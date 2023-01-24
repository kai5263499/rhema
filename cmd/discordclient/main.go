package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/caarlos0/env/v6"
	"github.com/gofrs/uuid"
	"github.com/kai5263499/rhema/client"
	"github.com/kai5263499/rhema/domain"
	pb "github.com/kai5263499/rhema/generated"
	"github.com/sirupsen/logrus"
)

const (
	SUCCESS_EMOJI = ":heavy_check_mark:"
	FAILURE_EMOJI = ":X:"
)

var (
	cfg             *domain.Config
	processorClient *client.ClientWithResponses
)

func main() {

	var err error
	cfg = &domain.Config{}
	if err = env.Parse(cfg); err != nil {
		logrus.WithError(err).Fatal("parse configs")
	}

	if level, err := logrus.ParseLevel(cfg.LogLevel); err != nil {
		logrus.WithError(err).Fatal("parse log level")
	} else {
		logrus.SetLevel(level)
	}

	logrus.SetFormatter(&logrus.JSONFormatter{})

	logrus.SetReportCaller(true)

	processorClient, err = client.NewClientWithResponses(cfg.RequestProcessorUri)
	if err != nil {
		logrus.WithError(err).Fatal("error creating processor client")
	}

	// Create a new Discord session using the provided bot token.
	dg, err := discordgo.New("Bot " + cfg.DiscordBotToken)
	if err != nil {
		logrus.WithError(err).Fatal("error creating Discord session")
	}

	// Register the messageCreate func as a callback for MessageCreate events.
	dg.AddHandler(messageCreate)

	// In this example, we only care about receiving message events.
	dg.Identify.Intents = discordgo.IntentsGuildMessages

	// Open a websocket connection to Discord and begin listening.
	err = dg.Open()
	if err != nil {
		fmt.Println("error opening connection,", err)
		return
	}

	// Wait here until CTRL-C or other term signal is received.
	fmt.Println("Bot is now running. Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	// Cleanly close down the Discord session.
	dg.Close()
}

// This function will be called (due to AddHandler above) every time a new
// message is created on any channel that the authenticated bot has access to.
func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	var err error
	// Ignore all messages created by the bot itself
	// This isn't required in this specific example but it's a good practice.
	if m.Author.ID == s.State.User.ID {
		return
	}

	if m.Content == "!random" {
		_, err = s.ChannelMessageSend(m.ChannelID, "nope")
		if err != nil {
			fmt.Println(err)
		}

		// _, err = s.ChannelFileSend(m.ChannelID, "random-gopher.png", response.Body)
		// if err != nil {
		// 	fmt.Println(err)
		// }
	}

	if strings.HasPrefix(m.Content, "!req") {
		reqStr := strings.Replace(m.Content, "!req ", "", 1)
		uris := strings.Split(reqStr, " ")

		sri := []client.SubmitRequestInput{}

		for _, uri := range uris {
			logrus.Debugf("sending request for %s", uri)

			contentType := pb.ContentType_URI.String()
			uuid1 := uuid.Must(uuid.NewV4()).String()
			now := uint64(time.Now().UTC().Unix())
			intnow := int(now)
			uuid2 := uuid.Must(uuid.NewV4()).String()

			ri := client.SubmitRequestInput{
				Atempo:         &cfg.Atempo,
				WordsPerMinute: &cfg.WordsPerMinute,
				EspeakVoice:    &cfg.EspeakVoice,
				Uri:            uri,
				Created:        &now,
				Title:          &uuid1,
				Type:           &contentType,
				SubmittedAt:    &intnow,
				SubmittedBy:    &m.Author.Username,
				RequestHash:    &uuid2,
			}

			sri = append(sri, ri)
		}

		resp, err := processorClient.SubmitRequestWithResponse(context.Background(), &client.SubmitRequestParams{}, sri)
		if err != nil {
			logrus.WithError(err).Error("error submitting request")
			return
		}

		requestHashes := make([]string, 0)
		if resp.JSON202 != nil && len(*resp.JSON202) > 0 {
			requests := resp.JSON202
			for _, request := range *requests {
				requestHashes = append(requestHashes, *request.RequestHash)
			}
		}

		// Send a text message with the list of Gophers
		_, err = s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("%s %s submitted %d requests with hashes %s", SUCCESS_EMOJI, m.Author.Username, len(uris), strings.Join(requestHashes, ",")))
		if err != nil {
			fmt.Println(err)
		}
	}
}
