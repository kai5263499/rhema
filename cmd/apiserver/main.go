package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/friendsofgo/graphiql"
	"github.com/gofrs/uuid"
	"github.com/gomodule/redigo/redis"
	"github.com/graphql-go/graphql"
	"github.com/sirupsen/logrus"

	"github.com/caarlos0/env"

	. "github.com/kai5263499/rhema"

	rg "github.com/redislabs/redisgraph-go"

	pb "github.com/kai5263499/rhema/generated"

	ast "github.com/graphql-go/graphql/language/ast"
)

type config struct {
	LogLevel      string `env:"LOG_LEVEL" envDefault:"info"`
	SubmittedWith string `env:"SUBMITTED_WITH" envDefault:"api"`
	Port          int    `env:"PORT" envDefault:"8080"`
	MQTTBroker    string `env:"MQTT_BROKER" envDefault:"tcp://172.17.0.3:1883"`
	MQTTClientID  string `env:"MQTT_CLIENT_ID" envDefault:"contentbot"`
	RedisHost     string `env:"REDIS_HOST"`
	RedisPort     string `env:"REDIS_PORT" envDefault:"6379"`
	RedisGraphKey string `env:"REDIS_GRAPH_KEY" envDefault:"rhema-content"`
}

var contentType = graphql.NewObject(
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
		req.Title = i.(string)
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

// Define the GraphQL Schema
func gqlSchema() graphql.Schema {
	fields := graphql.Fields{
		"contents": &graphql.Field{
			Type:        graphql.NewList(contentType),
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

				logrus.Debugf("gql query %s", query)

				result, queryErr := redisGraph.Query(query)
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
			Type:        contentType,
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

				logrus.Debugf("gql query %s", query)

				result, queryErr := redisGraph.Query(query)
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
			Type:        contentType,
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
					SubmittedWith: cfg.SubmittedWith,
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

				if err := mqttComms.SendRequest(req); err != nil {
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

var (
	cfg        config
	mqttComms  *MqttComms
	redisGraph *rg.Graph
)

type reqBody struct {
	Query string `json:"query"`
}

func gqlHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Body == nil {
			http.Error(w, "No query data", 400)
			return
		}

		var rBody reqBody
		err := json.NewDecoder(r.Body).Decode(&rBody)
		if err != nil {
			http.Error(w, "Error parsing JSON request body", 400)
		}

		fmt.Fprintf(w, "%s", processQuery(rBody.Query))

	})
}

func processQuery(query string) (result string) {
	logrus.Debug("process query")

	params := graphql.Params{Schema: gqlSchema(), RequestString: query}
	r := graphql.Do(params)
	if len(r.Errors) > 0 {
		fmt.Printf("failed to execute graphql operation, errors: %+v", r.Errors)
	}
	rJSON, _ := json.Marshal(r)

	return fmt.Sprintf("%s", rJSON)
}

func main() {
	cfg = config{}
	if err := env.Parse(&cfg); err != nil {
		logrus.WithError(err).Fatal("parse config")
	}

	if level, err := logrus.ParseLevel(cfg.LogLevel); err != nil {
		logrus.WithError(err).Fatal("parse log level")
	} else {
		logrus.SetLevel(level)
	}

	var mqttCommsErr error
	mqttComms, mqttCommsErr = NewMqttComms(cfg.MQTTClientID, cfg.MQTTBroker)
	if mqttCommsErr != nil {
		logrus.WithError(mqttCommsErr).Fatal("new mqtt comms")
	}

	redisConnStr := fmt.Sprintf("%s:%s", cfg.RedisHost, cfg.RedisPort)
	logrus.Debugf("connecting to redis %s", redisConnStr)
	redisConn, redisConnErr := redis.Dial("tcp", redisConnStr)
	if redisConnErr != nil {
		logrus.WithError(redisConnErr).Fatal("unable to connect to redis")
	}

	redisG := rg.GraphNew(cfg.RedisGraphKey, redisConn)
	redisGraph = &redisG

	graphiqlHandler, newGraphHandlerErr := graphiql.NewGraphiqlHandler("/graphql")
	if newGraphHandlerErr != nil {
		logrus.WithError(newGraphHandlerErr).Fatal("unable to create new graphiq handler")
	}

	logrus.Info("adding handlers")
	http.Handle("/graphiql", graphiqlHandler)
	http.Handle("/graphql", gqlHandler())

	logrus.Infof("listen and serve on %d", cfg.Port)
	if err := http.ListenAndServe(fmt.Sprintf(":%d", cfg.Port), nil); err != nil {
		logrus.WithError(err).Fatal("finished serving")
	}
}
