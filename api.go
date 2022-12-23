package rhema

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gomodule/redigo/redis"
	"github.com/gritzkoo/golang-health-checker/pkg/healthcheck"
	"github.com/kai5263499/rhema/domain"
	v1 "github.com/kai5263499/rhema/internal/v1"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	log "github.com/sirupsen/logrus"
	echoSwagger "github.com/swaggo/echo-swagger"
	"github.com/swaggo/swag"
)

var _ v1.ServerInterface = (*Api)(nil)

func NewApi(
	shutdownContext context.Context,
	shutdownCancelFunc context.CancelFunc,
	cfg *domain.Config,
	redisConn redis.Conn,
	requestProcessor domain.Processor,
	contentStorage domain.Storage,
) (*Api, error) {

	e := echo.New()
	e.Debug = true
	e.Use(middleware.Recover())

	a := &Api{
		shutdownContext:    shutdownContext,
		shutdownCancelFunc: shutdownCancelFunc,
		cfg:                cfg,
		redisConn:          redisConn,
		requestProcessor:   requestProcessor,
		contentStorage:     contentStorage,
		echo:               e,
	}

	all := v1.ServerInterfaceWrapper{
		Handler: a,
	}

	swagger, err := v1.GetSwagger()
	if err != nil {
		return nil, err
	}

	swaggerJson, err := swagger.MarshalJSON()
	if err != nil {
		return nil, err
	}

	var SwaggerInfo = &swag.Spec{
		Version:          "",
		Host:             "",
		BasePath:         "",
		Schemes:          []string{},
		Title:            "",
		Description:      "",
		InfoInstanceName: "swagger",
		SwaggerTemplate:  string(swaggerJson),
	}

	swag.Register(SwaggerInfo.InstanceName(), SwaggerInfo)

	e.GET("/", all.Ping)
	e.GET("/live", all.Live)
	e.GET("/swagger/*", echoSwagger.WrapHandler)

	log.Infof("listen and serve on %d", a.cfg.ApiHttpPort)

	return a, nil
}

type Api struct {
	shutdownContext    context.Context
	shutdownCancelFunc context.CancelFunc
	cfg                *domain.Config
	redisConn          redis.Conn
	requestProcessor   domain.Processor
	contentStorage     domain.Storage
	health             healthcheck.ApplicationConfig
	echo               *echo.Echo
}

func (a *Api) Start() {
	go a.httpListenerLoop()
}

func (a *Api) httpListenerLoop() {
	a.echo.Logger.Fatal(a.echo.Start(fmt.Sprintf(":%d", a.cfg.ApiHttpPort)))
}

func (a *Api) Ping(ctx echo.Context) error {
	return ctx.String(http.StatusOK, "OK")
}

func (a *Api) Live(ctx echo.Context) error {
	return ctx.JSON(http.StatusOK, v1.Liveness{
		Health: "OK",
		RealIp: ctx.RealIP(),
	})
}

func (a *Api) Ready(ctx echo.Context) error {
	h := healthcheck.HealthCheckerDetailed(a.health)

	return ctx.JSON(http.StatusOK, struct {
		Health healthcheck.ApplicationHealthDetailed `json:"health"`
		RealIP string                                `json:"real_ip"`
	}{
		Health: h,
		RealIP: ctx.RealIP(),
	})
}

func (a *Api) SubmitRequest(ctx echo.Context, params v1.SubmitRequestParams) error {
	return newHTTPError(http.StatusNotImplemented)
}

func (a *Api) RetrieveResultContent(ctx echo.Context, params v1.RetrieveResultContentParams) error {
	return newHTTPError(http.StatusNotImplemented)
}

func (a *Api) RetrieveResultStatus(ctx echo.Context, requestId string) error {
	return newHTTPError(http.StatusNotImplemented)
}

func newHTTPError(code int, errs ...error) error {
	if len(errs) == 0 {
		return echo.NewHTTPError(code)
	}
	err := errs[0]

	return echo.NewHTTPError(code).SetInternal(err)
}
