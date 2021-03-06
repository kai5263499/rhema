package rhema

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/gofrs/uuid"
	"github.com/kai5263499/glossa"
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

func NewBot(slackToken string, channels []string, tmpPath string, chownTo int, comms domain.Comms, atempo string, wpm int, espeakvoice string, submittedwith string) *Bot {
	URL_REGEXP = regexp.MustCompile("(?m)(http[^ <>\n]*)")

	bot := &Bot{
		slackToken:     slackToken,
		tmpPath:        tmpPath,
		chownTo:        chownTo,
		channels:       channels,
		patterns:       make([]botCommand, 0),
		cache:          cache.New(5*time.Minute, 10*time.Minute),
		comms:          comms,
		atempo:         atempo,
		wordsperminute: wpm,
		espeakvoice:    espeakvoice,
		submittedwith:  submittedwith,
	}

	commands, _ := bot.initPatterns()
	bot.patterns = commands

	return bot
}

var (
	URL_REGEXP *regexp.Regexp
)

type Bot struct {
	api            *slack.Client
	rtm            *slack.RTM
	slackToken     string
	channels       []string
	slackChannels  []*slack.Channel
	patterns       []botCommand
	cache          *cache.Cache
	tmpPath        string
	chownTo        int
	comms          domain.Comms
	atempo         string
	wordsperminute int
	espeakvoice    string
	submittedwith  string
}

func (b *Bot) processUri(uri string, user *slack.User, channel string, upload bool) {
	newUUID := uuid.Must(uuid.NewV4())

	contentRequest := pb.Request{
		Uri:            uri,
		Type:           pb.ContentType_URI,
		Title:          newUUID.String(),
		SubmittedBy:    user.Name,
		SubmittedAt:    uint64(time.Now().UTC().Unix()),
		Created:        uint64(time.Now().UTC().Unix()),
		RequestHash:    newUUID.String(),
		SubmittedWith:  b.submittedwith,
		ATempo:         b.atempo,
		WordsPerMinute: uint32(b.wordsperminute),
		ESpeakVoice:    b.espeakvoice,
	}

	if err := b.comms.SendRequest(contentRequest); err != nil {
		b.rtm.SendMessage(b.rtm.NewOutgoingMessage(fmt.Sprintf("%s <@%s> I failed to send request for processing `%s` err=%+#v", FAILURE_EMOJI, user.Name, uri, err), channel))
		logrus.WithError(err).WithFields(logrus.Fields{
			"uri": uri,
		}).Error("send request")
		return
	}
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

func (b *Bot) processUploadRequest(ci pb.Request, user *slack.User, channel string) {
	if err := b.comms.SendRequest(ci); err != nil {
		b.rtm.SendMessage(b.rtm.NewOutgoingMessage(fmt.Sprintf("%s <@%s> I failed to send request to the processor err=%+#v", FAILURE_EMOJI, user.Name, err), channel))
		logrus.WithError(err).Error("send request")
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

	if len(b.channels) > 0 {
		channel = b.channels[0]
	}

	user, err := b.api.GetUserInfo(file.User)
	if err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"fileId": ev.File.ID,
		}).Error("error getting file info")
		return
	}

	client := &http.Client{}
	req, _ := http.NewRequest("GET", file.URLPrivate, nil)
	req.Header.Set("Authorization", "Bearer "+b.slackToken)
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

		ci := pb.Request{
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

func (b *Bot) Process(ci pb.Request) error {
	switch ci.Type {
	case pb.ContentType_AUDIO:
		b.rtm.SendMessage(b.rtm.NewOutgoingMessage(fmt.Sprintf("%s I've processed the audio result", SUCCESS_EMOJI), b.channels[0]))
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
	b.slackChannels = make([]*slack.Channel, len(b.channels))
	for _, chanStr := range b.channels {
		channel, err := b.api.JoinChannel(chanStr)
		if err != nil {
			logrus.WithError(err).WithFields(logrus.Fields{
				"channel": chanStr,
			}).Error("failed to join channel")
		} else {
			logrus.WithFields(logrus.Fields{
				"channel": chanStr,
			}).Debug("joined channel")
			b.slackChannels = append(b.slackChannels, channel)
		}
	}
}

func (b *Bot) Start() {
	logrus.Debug("connecting to slack")

	b.api = slack.New(
		b.slackToken,
	)

	b.rtm = b.api.NewRTM()
	go b.rtm.ManageConnection()
	b.joinChannels()
	go b.slackReadLoop()
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
