package rhema

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gomodule/redigo/redis"
	"github.com/gorilla/mux"
	"github.com/kai5263499/rhema/domain"
	log "github.com/sirupsen/logrus"
)

func NewApi(
	shutdownContext context.Context,
	shutdownCancelFunc context.CancelFunc,
	cfg *domain.Config,
	redisConn redis.Conn,
) *Api {
	return &Api{
		shutdownContext:    shutdownContext,
		shutdownCancelFunc: shutdownCancelFunc,
		redisConn:          redisConn,
	}
}

type Api struct {
	shutdownContext    context.Context
	shutdownCancelFunc context.CancelFunc
	cfg                *domain.Config
	redisConn          redis.Conn
	router             *mux.Router
}

func (a *Api) Start() {
	go a.httpListenerLoop()
}

func (a *Api) httpListenerLoop() {
	log.Infof("listen and serve on %d", a.cfg.ApiHttpPort)
	if err := http.ListenAndServe(fmt.Sprintf(":%d", a.cfg.ApiHttpPort), a.router); err != nil {
		log.WithError(err).Fatal("finished serving")
	}
}
