package rhema

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gomodule/redigo/redis"
	"github.com/gorilla/mux"
	"github.com/graphql-go/graphql"
	"github.com/icza/gox/stringsx"
	"github.com/kai5263499/rhema/domain"
	pb "github.com/kai5263499/rhema/generated"
	rg "github.com/redislabs/redisgraph-go"
	log "github.com/sirupsen/logrus"
)

func NewApi(
	shutdownContext context.Context,
	shutdownCancelFunc context.CancelFunc,
	redisConn redis.Conn,
	redisGraph *rg.Graph,
	redisGraphKey string,
	comms domain.Comms,
	port int,
	auth0ClientId string,
	auth0ClientSecret string,
	auth0CallbackUrl string,
	auth0Domain string,
	submittedWith string,
) *Api {
	return &Api{
		shutdownContext:    shutdownContext,
		shutdownCancelFunc: shutdownCancelFunc,
		redisConn:          redisConn,
		redisGraph:         redisGraph,
		redisGraphKey:      redisGraphKey,
		comms:              comms,
		port:               port,
		auth0ClientId:      auth0ClientId,
		auth0ClientSecret:  auth0ClientSecret,
		auth0CallbackUrl:   auth0CallbackUrl,
		auth0Domain:        auth0Domain,
		submittedWith:      submittedWith,
	}
}

type Api struct {
	shutdownContext    context.Context
	shutdownCancelFunc context.CancelFunc
	redisConn          redis.Conn
	redisGraph         *rg.Graph
	redisGraphKey      string
	comms              domain.Comms
	port               int
	auth0ClientId      string
	auth0ClientSecret  string
	auth0CallbackUrl   string
	auth0Domain        string
	submittedWith      string
	router             *mux.Router
	contentType        *graphql.Object
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

/*
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
*/

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
