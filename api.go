package rhema

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/felixge/fgprof"
	"github.com/gofrs/uuid"
	"github.com/gomodule/redigo/redis"
	"github.com/gritzkoo/golang-health-checker/pkg/healthcheck"
	"github.com/kai5263499/rhema/domain"
	"github.com/kai5263499/rhema/generated"
	v1 "github.com/kai5263499/rhema/internal/v1"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/sirupsen/logrus"
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
	e.GET("/debug/fgprof", func(c echo.Context) error {
		fgprof.Handler().ServeHTTP(c.Response().Writer, c.Request())
		return nil
	})

	e.POST("/v1/request", all.SubmitRequest)
	e.GET("/v1/status/:request_id", all.RetrieveResultStatus)
	e.GET("/v1/result/:type/:request_id", all.RetrieveResultContent)
	e.GET("/v1/list-requests", all.ListAllRequests)

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

	var requests v1.SubmitRequestJSONRequestBody

	if err := ctx.Bind(&requests); err != nil {
		logrus.WithError(err).Error("error binding request body to v1.SubmitRequestJSONRequestBody")
		return newHTTPError(http.StatusBadRequest)
	}

	contentRequests := convertParamsToProto(&requests)

	responses := make([]*v1.SubmitRequestInput, len(requests))

	for idx, contentRequest := range contentRequests {
		responses[idx] = convertProtoToParams(contentRequest)

		go func(cr *generated.Request) {
			if err := a.requestProcessor.Process(cr); err != nil {
				logrus.WithError(err).WithFields(logrus.Fields{
					"requestHash": cr.RequestHash,
					"uri":         cr.Uri,
				}).Errorf("error processing request")
			}
		}(contentRequest)

	}

	return ctx.JSON(http.StatusAccepted, responses)
}

// (GET /result/{type}/{request_id})
func (a *Api) RetrieveResultContent(ctx echo.Context, pType string, requestId string) error {
	return ctx.JSON(http.StatusOK, a.contentStorage.ListAll())
}

// (GET /status/{request_id})
func (a *Api) RetrieveResultStatus(ctx echo.Context, requestId string) error {
	logrus.Debugf("looking up requestId=%s", requestId)

	req, err := a.contentStorage.Load(requestId)
	if err != nil {
		logrus.WithError(err).Errorf("error looking up requestId=%s", requestId)
		return newHTTPError(http.StatusNotFound)
	}

	return ctx.JSON(http.StatusOK, convertProtoToParams(req))
}

func newHTTPError(code int, errs ...error) error {
	if len(errs) == 0 {
		return echo.NewHTTPError(code)
	}
	err := errs[0]

	return echo.NewHTTPError(code).SetInternal(err)
}

func convertParamsToProto(submitRequests *v1.SubmitRequestJSONRequestBody) (requests []*generated.Request) {
	requests = make([]*generated.Request, len(*submitRequests))

	for idx, submitRequest := range *submitRequests {
		requests[idx] = &generated.Request{
			RequestHash: uuid.Must(uuid.NewV4()).String(),
			Uri:         submitRequest.Uri,
			SubmittedAt: uint64(time.Now().UTC().Unix()),
			Created:     uint64(time.Now().UTC().Unix()),
		}

		if submitRequest.Text != nil {
			requests[idx].Text = *submitRequest.Text
		}
		if submitRequest.Title != nil {
			requests[idx].Title = *submitRequest.Title
		}
		if submitRequest.EspeakVoice != nil {
			requests[idx].ESpeakVoice = *submitRequest.EspeakVoice
		}
		if submitRequest.Atempo != nil {
			requests[idx].ATempo = *submitRequest.Atempo
		}
		if submitRequest.WordsPerMinute != nil {
			requests[idx].WordsPerMinute = *submitRequest.WordsPerMinute
		}
		if submitRequest.SubmittedBy != nil {
			requests[idx].SubmittedBy = *submitRequest.SubmittedBy
		}
	}

	return
}

func convertProtoToParams(r *generated.Request) (o *v1.SubmitRequestInput) {
	contentType := r.Type.String()
	o = &v1.SubmitRequestInput{
		Uri:                 r.Uri,
		RequestHash:         &r.RequestHash,
		Title:               &r.Title,
		Atempo:              &r.ATempo,
		WordsPerMinute:      &r.WordsPerMinute,
		Length:              &r.Length,
		Size:                &r.Size,
		Text:                &r.Text,
		NumberOfConversions: &r.NumberOfConversions,
		Type:                &contentType,
		Created:             &r.Created,
	}

	if r.SubmittedAt > 0 {
		submittedAt := int(r.SubmittedAt)
		o.SubmittedAt = &submittedAt
	}

	return
}

// (GET /list-requests)
func (a *Api) ListAllRequests(ctx echo.Context) error {
	requests := a.contentStorage.ListAll()

	responses := make([]*v1.SubmitRequestInput, len(requests))

	for idx, request := range requests {
		responses[idx] = convertProtoToParams(request)
	}

	return ctx.JSON(http.StatusOK, responses)
}
