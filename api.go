package rhema

import (
	"crypto/rsa"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	// External
	jwtmiddleware "github.com/auth0/go-jwt-middleware"
	"github.com/codegangsta/negroni"
	"github.com/form3tech-oss/jwt-go"
	"github.com/friendsofgo/graphiql"
	"github.com/gofrs/uuid"
	"github.com/gomodule/redigo/redis"
	"github.com/gorilla/mux"
	"github.com/graphql-go/graphql"
	ast "github.com/graphql-go/graphql/language/ast"
	"github.com/icza/gox/stringsx"
	rg "github.com/redislabs/redisgraph-go"
	log "github.com/sirupsen/logrus"

	// Internal
	"github.com/kai5263499/rhema/domain"
	pb "github.com/kai5263499/rhema/generated"
)

func NewApi(redisConn redis.Conn,
	redisGraph *rg.Graph,
	redisGraphKey string,
	comms domain.Comms,
	port int,
	auth0ClientId string,
	auth0ClientSecret string,
	auth0CallbackUrl string,
	auth0Domain string,
	enableGraphiql bool,
	submittedWith string,
) *Api {
	return &Api{
		redisConn:         redisConn,
		redisGraph:        redisGraph,
		redisGraphKey:     redisGraphKey,
		comms:             comms,
		port:              port,
		auth0ClientId:     auth0ClientId,
		auth0ClientSecret: auth0ClientSecret,
		auth0CallbackUrl:  auth0CallbackUrl,
		auth0Domain:       auth0Domain,
		enableGraphiql:    enableGraphiql,
		submittedWith:     submittedWith,
	}
}

type Api struct {
	redisConn         redis.Conn
	redisGraph        *rg.Graph
	redisGraphKey     string
	comms             domain.Comms
	auth0Jwks         *Jwks
	port              int
	auth0ClientId     string
	auth0ClientSecret string
	auth0CallbackUrl  string
	auth0Domain       string
	enableGraphiql    bool
	submittedWith     string
	router            *mux.Router
	contentType       *graphql.Object
}

type Response struct {
	Message string `json:"message"`
}

type Jwks struct {
	Keys []JSONWebKeys `json:"keys"`
}

type JSONWebKeys struct {
	Kty string   `json:"kty"`
	Kid string   `json:"kid"`
	Use string   `json:"use"`
	N   string   `json:"n"`
	E   string   `json:"e"`
	X5c []string `json:"x5c"`
}

func (a *Api) loadAuth0Jwks() error {
	var err error

	url := fmt.Sprintf("https://%s/.well-known/jwks.json", a.auth0Domain)
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	a.auth0Jwks = &Jwks{}
	if err = json.NewDecoder(resp.Body).Decode(a.auth0Jwks); err != nil {
		return err
	}

	return nil
}

func (a *Api) getPemCertFromToken(token *jwt.Token) (*rsa.PublicKey, error) {
	for k, _ := range a.auth0Jwks.Keys {
		if token.Header["kid"] == a.auth0Jwks.Keys[k].Kid {
			pemCert := "-----BEGIN CERTIFICATE-----\n" + a.auth0Jwks.Keys[k].X5c[0] + "\n-----END CERTIFICATE-----"

			result, err := jwt.ParseRSAPublicKeyFromPEM([]byte(pemCert))
			if err != nil {
				return nil, err
			}

			return result, nil
		}
	}

	return nil, errors.New("unable to find appropriate key")
}

type reqBody struct {
	Query string `json:"query"`
}

func (a *Api) processQuery(query string) (string, error) {
	log.Debug("process query")

	params := graphql.Params{Schema: a.gqlSchema(), RequestString: query}
	r := graphql.Do(params)
	if len(r.Errors) > 0 {
		return "", fmt.Errorf("failed to execute graphql operation, errors: %+v", r.Errors)
	}
	rJSON, err := json.Marshal(r)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s", rJSON), nil
}

func (a *Api) Setup() {
	a.contentType = graphql.NewObject(
		graphql.ObjectConfig{
			Name: "Content",
			Fields: graphql.Fields{
				"requesthash": &graphql.Field{
					Type: graphql.String,
				},
				"text": &graphql.Field{
					Type: graphql.String,
				},
				"title": &graphql.Field{
					Type: graphql.String,
				},
				"created": &graphql.Field{
					Type: graphql.Int,
				},
				"size": &graphql.Field{
					Type: graphql.Int,
				},
				"length": &graphql.Field{
					Type: graphql.Int,
				},
				"uri": &graphql.Field{
					Type: graphql.String,
				},
				"downloaduri": &graphql.Field{
					Type: graphql.String,
				},
				"wpm": &graphql.Field{
					Type: graphql.Int,
				},
				"espeakvoice": &graphql.Field{
					Type: graphql.String,
				},
				"atempo": &graphql.Field{
					Type: graphql.String,
				},
				"type": &graphql.Field{
					Type: graphql.Int,
				},
			},
		},
	)

	var gqlHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Body == nil {
			http.Error(w, "No query data", 400)
			return
		}

		var rBody reqBody
		if err := json.NewDecoder(r.Body).Decode(&rBody); err != nil {
			http.Error(w, "Error parsing JSON request body", 400)
			return
		}

		q, err := a.processQuery(rBody.Query)
		if err != nil {
			http.Error(w, "Error processing query", 400)
		}
		fmt.Fprintf(w, "%s", q)
	})

	jwtMiddleware := jwtmiddleware.New(jwtmiddleware.Options{
		ValidationKeyGetter: func(token *jwt.Token) (interface{}, error) {
			aud := a.auth0CallbackUrl
			checkAud := token.Claims.(jwt.MapClaims).VerifyAudience(aud, false)
			if !checkAud {
				return token, errors.New("invalid audience")
			}

			iss := fmt.Sprintf("https://%s/", a.auth0Domain)
			checkIss := token.Claims.(jwt.MapClaims).VerifyIssuer(iss, false)
			if !checkIss {
				return token, errors.New("invalid issuer")
			}

			pemCert, err := a.getPemCertFromToken(token)
			if err != nil {
				return token, errors.New("Unable to find PEM cert from token")
			}
			return pemCert, nil
		},
		SigningMethod: jwt.SigningMethodRS256,
	})

	graphiqlHandler, newGraphHandlerErr := graphiql.NewGraphiqlHandler("/graphql")
	if newGraphHandlerErr != nil {
		log.WithError(newGraphHandlerErr).Fatal("unable to create new graphiq handler")
	}

	if err := a.loadAuth0Jwks(); err != nil {
		log.WithError(err).Fatal("unable to load Auth0 JWKS")
	}

	log.Info("adding handlers")
	a.router = mux.NewRouter()

	if a.enableGraphiql {
		log.Info("enabling graphiql endpoint")
		a.router.Handle("/graphiql", graphiqlHandler)
	}

	a.router.Handle("/graphql", negroni.New(
		negroni.HandlerFunc(jwtMiddleware.HandlerWithNext),
		negroni.Wrap(http.HandlerFunc(gqlHandler)),
	))
}

func (a *Api) Start() {
	go a.httpListenerLoop()
}

func (a *Api) httpListenerLoop() {
	log.Infof("listen and serve on %d", a.port)
	if err := http.ListenAndServe(fmt.Sprintf(":%d", a.port), a.router); err != nil {
		log.WithError(err).Fatal("finished serving")
	}
}

// Define the GraphQL Schema
func (a *Api) gqlSchema() graphql.Schema {
	fields := graphql.Fields{
		"contents": &graphql.Field{
			Type:        graphql.NewList(a.contentType),
			Description: "All content items",
			Args: graphql.FieldConfigArgument{
				"submittedby": &graphql.ArgumentConfig{
					Type: graphql.NewNonNull(graphql.String),
				},
				"skip": &graphql.ArgumentConfig{
					Type: graphql.Int,
				},
				"take": &graphql.ArgumentConfig{
					Type: graphql.Int,
				},
			},
			Resolve: func(params graphql.ResolveParams) (interface{}, error) {
				submittedby := params.Args["submittedby"].(string)

				fields := GetSelectedFields([]string{"contents"}, params)
				selectedFields := fieldsToGQLSelect(fields)

				skip := 0
				if v, success := params.Args["skip"]; success {
					skip = v.(int)
				}

				limit := 10
				if v, success := params.Args["limit"]; success {
					limit = v.(int)
				}

				requests := make([]pb.Request, 0)
				query := fmt.Sprintf("MATCH (a:actor {submittedBy:'%s'})-[s:submitted]->(c:content) RETURN %s ORDER BY c.created DESC SKIP %d LIMIT %d", submittedby, selectedFields, skip, limit)

				log.Debugf("gql query %s", query)

				result, queryErr := a.redisGraph.Query(query)
				if queryErr != nil {
					return nil, fmt.Errorf("error querying for %s", submittedby)
				}

				for result.Next() {
					var req pb.Request
					if err := graphNodeToReq(result.Record(), &req); err != nil {
						return nil, errors.New("error converting graph node to request")
					}

					requests = append(requests, req)
				}

				return requests, nil
			},
		},
		"content": &graphql.Field{
			Type:        a.contentType,
			Description: "Get content by requesthash",
			Args: graphql.FieldConfigArgument{
				"requesthash": &graphql.ArgumentConfig{
					Type: graphql.NewNonNull(graphql.String),
				},
			},
			Resolve: func(params graphql.ResolveParams) (interface{}, error) {
				requesthash := params.Args["requesthash"].(string)

				fields := GetSelectedFields([]string{"content"}, params)
				selectedFields := fieldsToGQLSelect(fields)

				query := fmt.Sprintf("MATCH (c:content {requesthash:'%s'}) RETURN %s LIMIT 1", requesthash, selectedFields)

				log.Debugf("gql query %s", query)

				result, queryErr := a.redisGraph.Query(query)
				if queryErr != nil {
					return nil, fmt.Errorf("error querying for %s", requesthash)
				}

				var req pb.Request
				for result.Next() {
					if err := graphNodeToReq(result.Record(), &req); err != nil {
						return nil, errors.New("error converting graph node to request")
					}
				}

				return req, nil
			},
		},
		"request": &graphql.Field{
			Type:        a.contentType,
			Description: "Create a content request",
			Args: graphql.FieldConfigArgument{
				"text": &graphql.ArgumentConfig{
					Type: graphql.String,
				},
				"uri": &graphql.ArgumentConfig{
					Type: graphql.String,
				},
				"submittedby": &graphql.ArgumentConfig{
					Type: graphql.NewNonNull(graphql.String),
				},
				"atempo": &graphql.ArgumentConfig{
					Type: graphql.String,
				},
				"wpm": &graphql.ArgumentConfig{
					Type: graphql.Int,
				},
				"espeakvoice": &graphql.ArgumentConfig{
					Type: graphql.String,
				},
			},
			Resolve: func(params graphql.ResolveParams) (interface{}, error) {

				newUUID := uuid.Must(uuid.NewV4())
				req := pb.Request{
					Title:         newUUID.String(),
					SubmittedBy:   params.Args["submittedby"].(string),
					SubmittedAt:   uint64(time.Now().UTC().Unix()),
					SubmittedWith: a.submittedWith,
					Created:       uint64(time.Now().UTC().Unix()),
					RequestHash:   newUUID.String(),
				}

				if v, success := params.Args["uri"]; success {
					req.Uri = v.(string)
				}

				if v, success := params.Args["text"]; success {
					req.Text = v.(string)
				}

				if v, success := params.Args["atempo"]; success {
					req.ATempo = v.(string)
				}

				if v, success := params.Args["espeakvoice"]; success {
					req.ESpeakVoice = v.(string)
				}

				if v, success := params.Args["wordsperminute"]; success {
					req.WordsPerMinute = uint32(v.(int))
				}

				var reqType pb.ContentType
				if len(req.Text) > 0 {
					reqType = pb.ContentType_TEXT
				} else if len(req.Uri) > 0 {
					reqType = pb.ContentType_URI
				}

				if reqType != pb.ContentType_TEXT && reqType != pb.ContentType_URI {
					return nil, errors.New("request type cannot be detected. text and uri are empty")
				}

				req.Type = reqType

				if err := a.comms.SendRequest(req); err != nil {
					return nil, errors.New("error unable to send request")
				}

				return req, nil
			},
		},
	}

	rootQuery := graphql.ObjectConfig{Name: "RootQuery", Fields: fields}
	schemaConfig := graphql.SchemaConfig{Query: graphql.NewObject(rootQuery)}
	schema, err := graphql.NewSchema(schemaConfig)
	if err != nil {
		fmt.Printf("failed to create new schema, error: %v", err)
	}

	return schema
}

// GetSelectedFields returns the list of request fields listed under provider selection path in the Graphql query.
func GetSelectedFields(selectionPath []string, resolveParams graphql.ResolveParams) []string {
	fields := resolveParams.Info.FieldASTs

	for _, propName := range selectionPath {
		found := false

		for _, field := range fields {
			if field.Name.Value == propName {
				selections := field.SelectionSet.Selections
				fields = make([]*ast.Field, 0)

				for _, selection := range selections {
					fields = append(fields, selection.(*ast.Field))
				}

				found = true

				break
			}
		}

		if !found {
			return []string{}
		}
	}

	var collect []string

	for _, field := range fields {
		collect = append(collect, field.Name.Value)
	}

	return collect
}

func fieldsToGQLSelect(strs []string) string {
	var sb strings.Builder
	for idx, str := range strs {
		sb.WriteString("c.")
		sb.WriteString(str)
		if idx < len(strs)-1 {
			sb.WriteString(",")
		}
	}
	return sb.String()
}

func graphNodeToReq(rec *rg.Record, req *pb.Request) error {
	if i, _ := rec.Get("c.created"); i != nil {
		req.Created = uint64(i.(int))
	}

	if i, _ := rec.Get("c.downloaduri"); i != nil {
		req.DownloadURI = i.(string)
	}

	if i, _ := rec.Get("c.type"); i != nil {
		req.Type = pb.ContentType(pb.ContentType_value[i.(string)])
	}

	if i, _ := rec.Get("c.size"); i != nil {
		req.Size = uint64(i.(int))
	}

	if i, _ := rec.Get("c.length"); i != nil {
		req.Length = uint64(i.(int))
	}

	if i, _ := rec.Get("c.uri"); i != nil {
		req.Uri = i.(string)
	}

	if i, _ := rec.Get("c.title"); i != nil {
		req.Title = stringsx.Clean(i.(string))
	}

	if i, _ := rec.Get("c.wpm"); i != nil {
		req.WordsPerMinute = uint32(i.(int))
	}

	if i, _ := rec.Get("c.atempo"); i != nil {
		req.ATempo = i.(string)
	}

	if i, _ := rec.Get("c.text"); i != nil {
		req.Text = i.(string)
	}

	if i, _ := rec.Get("c.requesthash"); i != nil {
		req.RequestHash = i.(string)
	}

	if i, _ := rec.Get("c.espeakvoice"); i != nil {
		req.RequestHash = i.(string)
	}

	return nil
}
