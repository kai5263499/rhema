package main

import (
	"context"
	"time"

	"github.com/caarlos0/env"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/gofrs/uuid"
	"github.com/sirupsen/logrus"

	. "github.com/kai5263499/rhema"

	"encoding/json"
	"net/http"

	pb "github.com/kai5263499/rhema/generated"

	"github.com/kai5263499/rhema/generated"

	_ "github.com/kai5263499/rhema/cmd/apiserver/docs" // docs is generated by Swag CLI, you have to import it.

	httpSwagger "github.com/swaggo/http-swagger"

	"github.com/gorilla/mux"

	firebase "firebase.google.com/go"
	"firebase.google.com/go/auth"
)

type config struct {
	MQTTBroker string `env:"MQTT_BROKER" envDefault:"tcp://172.17.0.3:1883"`
	LogLevel   string `env:"LOG_LEVEL" envDefault:"debug"`
}

const (
	esIndex = "requests"
)

var (
	cfg    config
	fbApp  *firebase.App
	fbAuth *auth.Client
	comms  *MqttComms
)

// @title Rhema API
// @version 1.0
// @description This is a REST interface for the Rhema content to audio system
// @termsOfService http://swagger.io/terms/
// @contact.name Wes Widner
// @contact.email wes@manwe.io
// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html
// @host localhost
// @BasePath /
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

	if app, err := firebase.NewApp(context.Background(), nil); err != nil {
		logrus.WithError(err).Fatal("new firebase app")
	} else {
		fbApp = app
	}

	// Access auth service from the default app
	if client, err := fbApp.Auth(context.Background()); err != nil {
		logrus.WithError(err).Fatal("new firebase auth")
	} else {
		fbAuth = client
	}

	opts := mqtt.NewClientOptions().AddBroker(cfg.MQTTBroker).SetClientID("apiserver")
	// opts.SetDefaultPublishHandler(mqttMessageHandler)
	opts.SetKeepAlive(2 * time.Second)
	opts.SetPingTimeout(1 * time.Second)

	mqttClient := mqtt.NewClient(opts)
	if token := mqttClient.Connect(); token.Wait() && token.Error() != nil {
		logrus.WithError(token.Error()).Fatal("mqtt new client")
	}

	router := mux.NewRouter()
	router.HandleFunc("/request", createRequest).Methods("POST")
	router.HandleFunc("/request/{requestHash}", getRequest).Methods("GET")
	router.HandleFunc("/requests/{submittedBy}", getRequests).Methods("GET")

	router.PathPrefix("/swagger").Handler(httpSwagger.WrapHandler)

	logrus.Infof("serving")
	if err := http.ListenAndServe(":8080", router); err != nil {
		logrus.WithError(err).Fatalf("finished serving")
	}
}

// CreateRequest godoc
// @Summary Create a new content processing request
// @Description Create a new content to mp3 conversion request with the input paylod
// @Tags requests
// @Accept  json
// @Produce  json
// @Success 200 {object} generated.Request
// @Router /request [post]
func createRequest(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var request generated.Request
	json.NewDecoder(r.Body).Decode(&request)

	newUUID := uuid.Must(uuid.NewV4())

	request.SubmittedAt = uint64(time.Now().UTC().Unix())

	if request.Created == 0 {
		request.Created = uint64(time.Now().UTC().Unix())
	}

	if request.RequestHash == "" {
		request.RequestHash = newUUID.String()
	}

	if request.Text != "" {
		request.Type = pb.Request_TEXT
	} else if request.Uri != "" {
		request.Type = pb.Request_URI
		request.Title = newUUID.String()
	}

	if err := comms.SendRequest(request); err != nil {
		request.Text = "Invalid request. Uri or Text field must be provided."
	}

	json.NewEncoder(w).Encode(request)
}

// GetOrders godoc
// @Summary Get details of all requests
// @Description Get details of all requests
// @Tags requests
// @Accept  json
// @Produce  json
// @Success 200 {array} generated.Request
// @Router /requests/{submittedBy} [get]
func getRequests(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	params := mux.Vars(r)
	submittedBy := params["submittedBy"]

	logrus.WithFields(logrus.Fields{
		"SubmittedBy": submittedBy,
	}).Debugf("performing request")

	var requests []generated.Request

	json.NewEncoder(w).Encode(requests)
}

// GetOrder godoc
// @Summary Get details for a given requestID
// @Description Get details of a content request for a given requestID
// @Tags requests
// @Accept  json
// @Produce  json
// @Param orderId path int true "ID of the order"
// @Success 200 {object} generated.Request
// @Router /request/{requestHash} [get]
func getRequest(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var request pb.Request

	json.NewEncoder(w).Encode(request)
}
