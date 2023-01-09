package rhema

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/gofrs/uuid"
	"github.com/kai5263499/glossa"
	"github.com/kai5263499/rhema/client"
	"github.com/kai5263499/rhema/domain"
	pb "github.com/kai5263499/rhema/generated"
	"github.com/patrickmn/go-cache"
	"github.com/sirupsen/logrus"
	"github.com/slack-go/slack"
)

const (
	SUCCESS_EMOJI = ":heavy_check_mark:"
	FAILURE_EMOJI = ":X:"
)

type botCommand struct {
	action  pb.BotAction
	pattern glossa.Pattern
}

func NewBot(cfg *domain.Config) (*Bot, error) {
	URL_REGEXP = regexp.MustCompile("(?m)(http[^ <>\n]*)")

	bot := &Bot{
		cfg:      cfg,
		patterns: make([]botCommand, 0),
		cache:    cache.New(5*time.Minute, 10*time.Minute),
		channels: make(map[string]*slack.Channel),
	}

	for _, channelName := range cfg.Channels {
		bot.channels[channelName] = nil
	}

	commands, err := bot.initPatterns()
	if err != nil {
		return nil, err
	}
	bot.patterns = commands

	return bot, nil
}

var (
	URL_REGEXP *regexp.Regexp
)

type Bot struct {
	cfg      *domain.Config
	api      *slack.Client
	rtm      *slack.RTM
	patterns []botCommand
	cache    *cache.Cache
	channels map[string]*slack.Channel
}

func (b *Bot) processUri(uri string, user *slack.User, channel string, upload bool) {

	logrus.Debugf("sending request for %s", uri)

	processorClient, err := client.NewClientWithResponses(b.cfg.RequestProcessorUri)
	if err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"uri": uri,
		}).Error("error creating processor client")
		return
	}

	sri := client.SubmitRequestInput{
		Uri: uri,
	}

	contentType := pb.ContentType_URI.String()
	sri.Type = &contentType

	uuid1 := uuid.Must(uuid.NewV4()).String()
	sri.Title = &uuid1

	sri.SubmittedBy = &user.Name

	now := uint64(time.Now().UTC().Unix())
	intnow := int(now)

	sri.SubmittedAt = &intnow
	sri.Created = &now

	uuid2 := uuid.Must(uuid.NewV4()).String()
	sri.RequestHash = &uuid2

	resp, err := processorClient.SubmitRequestWithResponse(context.Background(), nil, []client.SubmitRequestInput{
		sri,
	})
	if err != nil {
		return
	}

	requestHash := ""
	if resp.JSON202 != nil && len(*resp.JSON202) > 0 {
		requestHash = *(*resp.JSON202)[0].RequestHash
	}

	b.rtm.SendMessage(b.rtm.NewOutgoingMessage(fmt.Sprintf("%s <@%s> submitted uri %s has request hash of %s", SUCCESS_EMOJI, user.Name, uri, requestHash), channel))
}

func (b *Bot) processMessage(ev *slack.MessageEvent) {
	if strings.Contains(ev.Text, "uploaded a file") {
		return
	}

	user, err := b.api.GetUserInfo(ev.User)
	if err != nil {
		logrus.WithError(err).Errorf("error retrieving user info")
		return
	}

	for _, pat := range b.patterns {
		matched, args, err := pat.pattern.Match(ev.Text)

		if err != nil {
			logrus.WithFields(logrus.Fields{
				"err": err,
			}).Error("error processing command")
			continue
		}

		logrus.WithFields(logrus.Fields{
			"pattern":  pat.pattern,
			"matched":  matched,
			"action":   pat.action,
			"args_len": len(args),
		}).Debug("pattern match")

		if !matched {
			continue
		}

		switch pat.action {
		case pb.BotAction_GET_CONFIG:
			b.getConfig(args[0].(string), user, ev.Channel)
		case pb.BotAction_SET_CONFIG:
			b.setConfig(args[0].(string), args[1].(string), user, ev.Channel)
		}

		return
	}

	uris := URL_REGEXP.FindAllString(ev.Text, -1)

	if len(uris) < 1 {
		logrus.Tracef("no uris")
		return
	}

	plural := ""
	if len(uris) > 1 {
		plural = "s"
	}

	b.rtm.SendMessage(b.rtm.NewOutgoingMessage(fmt.Sprintf("%s <@%s> I'll process %d uri%s", SUCCESS_EMOJI, user.Name, len(uris), plural), ev.Channel))

	for _, parsedUri := range uris {
		go b.processUri(parsedUri, user, ev.Channel, false)
	}
}

func (b *Bot) processUploadRequest(ci *pb.Request, user *slack.User, channel string) {
	processorClient, err := client.NewClientWithResponses(b.cfg.RequestProcessorUri)
	if err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"requestHash": ci.RequestHash,
		}).Error("error creating processor client")
		return
	}

	sri := client.SubmitRequestInput{
		RequestHash:   &ci.RequestHash,
		Title:         &ci.Title,
		Text:          &ci.Text,
		Uri:           ci.Uri,
		SubmittedBy:   &user.Name,
		SubmittedWith: &ci.SubmittedWith,
	}

	contentType := pb.ContentType_TEXT.String()
	sri.Type = &contentType

	now := uint64(time.Now().UTC().Unix())
	intnow := int(now)
	sri.SubmittedAt = &intnow

	_, err = processorClient.SubmitRequestWithResponse(context.Background(), nil, []client.SubmitRequestInput{
		sri,
	})
	if err != nil {
		b.rtm.SendMessage(b.rtm.NewOutgoingMessage(fmt.Sprintf("%s <@%s> Error sending your request to the processor", FAILURE_EMOJI, user.Name), channel))
		return
	}

	b.rtm.SendMessage(b.rtm.NewOutgoingMessage(fmt.Sprintf("%s <@%s> I've sent your request to the processor", SUCCESS_EMOJI, user.Name), channel))
}

func (b *Bot) processFileUpload(ev *slack.FileSharedEvent) {
	var err error
	var channel string

	_, found := b.cache.Get(ev.File.ID)
	if found {
		logrus.Debugf("%s found in cache, skipping", ev.FileID)
		return
	}

	b.cache.Set(ev.File.ID, true, cache.DefaultExpiration)

	file, _, _, err := b.api.GetFileInfo(ev.File.ID, 0, 0)
	if err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"fileId": ev.File.ID,
		}).Error("error getting file info")
		return
	}

	if len(b.cfg.Channels) > 0 {
		channel = b.cfg.Channels[0]
	}

	user, err := b.api.GetUserInfo(file.User)
	if err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"fileId": ev.File.ID,
		}).Error("error getting file info")
		return
	}

	client := &http.Client{}
	req, err := http.NewRequest("GET", file.URLPrivate, nil)
	if err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"fileId": ev.File.ID,
		}).Error("error creating GET request to retrieve file")
		return
	}
	req.Header.Set("Authorization", "Bearer "+b.cfg.SlackToken)
	resp, err := client.Do(req)
	if err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"fileURLPrivate": file.URLPrivate,
		}).Error("error downloading shared file")
		return
	}

	body, _ := ioutil.ReadAll(resp.Body)
	bodyStr := string(body)

	if file.Title == "uris" {
		uris := URL_REGEXP.FindAllString(bodyStr, -1)

		logrus.Debugf("parsed %d uris", len(uris))

		for _, parsedUrl := range uris {
			unescapedUrl, _ := url.QueryUnescape(parsedUrl)
			go b.processUri(unescapedUrl, user, channel, false)
		}
	} else {

		newUUID := uuid.Must(uuid.NewV4())

		ci := &pb.Request{
			Uri:         file.URLPrivate,
			Title:       file.Title,
			Type:        pb.ContentType_TEXT,
			Text:        bodyStr,
			SubmittedBy: ev.File.User,
			Created:     uint64(time.Now().UTC().Unix()),
			SubmittedAt: uint64(time.Now().UTC().Unix()),
			RequestHash: newUUID.String(),
			Length:      uint64(len(bodyStr)),
		}

		b.rtm.SendMessage(b.rtm.NewOutgoingMessage(fmt.Sprintf("%s <@%s> I'll process the text snippet upload with %d bytes", SUCCESS_EMOJI, user.Name, len(bodyStr)), channel))

		go b.processUploadRequest(ci, user, channel)
	}
}

func (b *Bot) Process(ci *pb.Request) error {
	switch ci.Type {
	case pb.ContentType_AUDIO:
		b.rtm.SendMessage(b.rtm.NewOutgoingMessage(fmt.Sprintf("%s I've processed the audio result", SUCCESS_EMOJI), b.cfg.Channels[0]))
	}

	return nil
}

func (b *Bot) slackReadLoop() {
	logrus.Debug("start slack read loop")

	for {
		select {
		case msg := <-b.rtm.IncomingEvents:
			logrus.Debugf("event received %#v", msg)
			switch ev := msg.Data.(type) {
			case *slack.ConnectedEvent:
				logrus.Infof("Connection counter: %d", ev.ConnectionCount)
				for _, channel := range b.channels {
					b.rtm.SendMessage(b.rtm.NewOutgoingMessage(fmt.Sprintf("%s I'm connected now with connection count %d", SUCCESS_EMOJI, ev.ConnectionCount), channel.ID))
				}
			case *slack.MessageEvent:
				logrus.Debug("message event received!")
				go b.processMessage(ev)
			case *slack.FileSharedEvent:
				logrus.Debug("file shared event recieved!")
				go b.processFileUpload(ev)
			case *slack.RTMError:
				logrus.Errorf("RTMError: %s", ev.Error())
			case *slack.InvalidAuthEvent:
				logrus.Error("Invalid credentials!")
				return
			default:
				//Take no action
			}
		}
	}
}

func (b *Bot) joinChannels() {
	for channelName, origChan := range b.channels {
		channel, warning, warnings, err := b.api.JoinConversation(origChan.ID)
		logrus.Debugf("joinconversation warning=%s warnings=%v", warning, warnings)
		if err != nil {
			logrus.WithError(err).WithFields(logrus.Fields{
				"channel": channelName,
			}).Error("failed to join channel")
		} else {
			logrus.WithFields(logrus.Fields{
				"channel": channelName,
			}).Debug("joined channel")

			b.channels[channelName] = channel
		}

		logrus.Debugf("sending welcome message to channelName=%s channelID=%s", channel.Name, channel.ID)
		b.rtm.SendMessage(b.rtm.NewOutgoingMessage(fmt.Sprintf("%s bot has entered the chat", SUCCESS_EMOJI), channel.ID))
	}
}

func (b *Bot) Start() error {
	logrus.Debugf("connecting to slack with token %s", b.cfg.SlackToken)

	b.api = slack.New(
		b.cfg.SlackToken,
	)

	channels, _, err := b.api.GetConversations(&slack.GetConversationsParameters{
		ExcludeArchived: true,
		Limit:           100,
	})
	if err != nil {
		return err
	}

	for _, channel := range channels {
		if _, found := b.channels[channel.Name]; found {
			b.channels[channel.Name] = &channel
		}
	}

	b.rtm = b.api.NewRTM()

	go b.rtm.ManageConnection()
	b.joinChannels()
	go b.slackReadLoop()

	return nil
}

func (b *Bot) getConfig(key string, user *slack.User, channel string) {
	// TODO: Return user prefs
	found, val := false, ""

	logrus.WithFields(logrus.Fields{
		"found": found,
		"val":   val,
	}).Debug("get setting")

	if found {
		b.rtm.SendMessage(b.rtm.NewOutgoingMessage(fmt.Sprintf("%s <@%s> config property %s=%s", SUCCESS_EMOJI, user.Name, key, val), channel))
	} else {
		b.rtm.SendMessage(b.rtm.NewOutgoingMessage(fmt.Sprintf("%s <@%s> config property %s not found", FAILURE_EMOJI, user.Name, key), channel))
	}
}

func (b *Bot) setConfig(key string, val string, user *slack.User, channel string) {
	// TODO: Capture user prefs
	result := false

	if result {
		b.rtm.SendMessage(b.rtm.NewOutgoingMessage(fmt.Sprintf("%s <@%s> config property %s set to %s", SUCCESS_EMOJI, user.Name, key, val), channel))
	} else {
		b.rtm.SendMessage(b.rtm.NewOutgoingMessage(fmt.Sprintf("%s <@%s> config property %s not set to %s", FAILURE_EMOJI, user.Name, key, val), channel))
	}
}

func (b *Bot) initPatterns() ([]botCommand, error) {
	patterns := make([]botCommand, 0)

	parser, err := glossa.NewParser()
	if err != nil {
		return nil, err
	}

	pattern1, err := parser.Parse(`set <string> to <string>`)
	if err != nil {
		return nil, err
	}

	bc1 := botCommand{
		pattern: pattern1,
		action:  pb.BotAction_SET_CONFIG,
	}

	patterns = append(patterns, bc1)

	pattern2, err := parser.Parse(`get <string>`)
	if err != nil {
		return nil, err
	}

	bc2 := botCommand{
		pattern: pattern2,
		action:  pb.BotAction_GET_CONFIG,
	}

	patterns = append(patterns, bc2)

	return patterns, nil
}
