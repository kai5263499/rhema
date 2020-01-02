package rhema

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/gofrs/uuid"
	"github.com/kai5263499/glossa"
	. "github.com/kai5263499/rhema/domain"
	pb "github.com/kai5263499/rhema/generated"
	"github.com/nlopes/slack"
	"github.com/sirupsen/logrus"
)

const (
	SUCCESS_EMOJI = ":heavy_check_mark:"
	FAILURE_EMOJI = ":X:"
)

type botCommand struct {
	action  pb.BotAction
	pattern glossa.Pattern
}

func NewBot(slackToken string, requestProcessor Processor, localPath string, chownTo int, channels []string) *Bot {
	bot := &Bot{
		slackToken:       slackToken,
		requestProcessor: requestProcessor,
		localPath:        localPath,
		chownTo:          chownTo,
		channels:         channels,
		patterns:         make([]botCommand, 0),
	}

	commands, _ := bot.initPatterns()
	bot.patterns = commands

	return bot
}

var (
	URL_REGEXP *regexp.Regexp
)

type Bot struct {
	api              *slack.Client
	rtm              *slack.RTM
	slackToken       string
	requestProcessor Processor
	localPath        string
	chownTo          int
	channels         []string
	slackChannels    []*slack.Channel
	patterns         []botCommand
}

func (b *Bot) processUri(uri string, user *slack.User, channel string, upload bool) {
	var err error

	newUUID := uuid.Must(uuid.NewV4())

	contentRequest := pb.Request{
		Uri:         uri,
		Type:        pb.Request_URI,
		Title:       newUUID.String(),
		SubmittedBy: user.Name,
		SubmittedAt: uint64(time.Now().UTC().Unix()),
		Created:     uint64(time.Now().UTC().Unix()),
		RequestHash: newUUID.String(),
	}

	resultingItem, err := b.requestProcessor.Process(contentRequest)
	if err != nil {
		b.rtm.SendMessage(b.rtm.NewOutgoingMessage(fmt.Sprintf("%s <@%s> I failed to process `%s` err=%+#v", FAILURE_EMOJI, user.Name, uri, err), channel))
		logrus.WithFields(logrus.Fields{
			"uri": uri,
			"err": err,
		}).Errorf("processing item")
		return
	}

	urlFilename, err := GetFilePath(resultingItem)
	if err != nil {
		b.rtm.SendMessage(b.rtm.NewOutgoingMessage(fmt.Sprintf("%s <@%s> I failed to get file path `%s` err=%+#v", FAILURE_EMOJI, user.Name, resultingItem.Uri, err), channel))
		logrus.WithFields(logrus.Fields{
			"uri": uri,
			"err": err,
		}).Errorf("get file path")
		return
	}

	baseUrlFilename := path.Base(urlFilename)

	urlFullFilename := filepath.Join(b.localPath, baseUrlFilename)

	err = DownloadUriToFile(resultingItem.DownloadURI, urlFullFilename)
	if err != nil {
		b.rtm.SendMessage(b.rtm.NewOutgoingMessage(fmt.Sprintf("%s <@%s> I failed to download `%s` -> `%s` err=%+#v", FAILURE_EMOJI, user.Name, resultingItem.DownloadURI, path.Base(urlFullFilename), err), channel))
		logrus.WithFields(logrus.Fields{
			"downloadURI":     resultingItem.DownloadURI,
			"urlFullFilename": urlFullFilename,
			"err":             err,
		}).Errorf("download uri to file")
		return
	}

	file, err := os.Open(urlFullFilename)
	if err != nil {
		b.rtm.SendMessage(b.rtm.NewOutgoingMessage(fmt.Sprintf("%s <@%s> I failed to open `%s` err=%+#v", FAILURE_EMOJI, user.Name, path.Base(urlFullFilename), err), channel))
		logrus.WithFields(logrus.Fields{
			"urlFullFilename": urlFullFilename,
			"err":             err,
		}).Errorf("open url file")
		return
	}

	fileInfo, err := file.Stat()
	if err != nil {
		b.rtm.SendMessage(b.rtm.NewOutgoingMessage(fmt.Sprintf("%s <@%s> I failed to stat `%s` err=%+#v", FAILURE_EMOJI, user.Name, path.Base(urlFullFilename), err), channel))
		logrus.WithFields(logrus.Fields{
			"urlFullFilename": urlFullFilename,
			"err":             err,
		}).Errorf("file info")
		return
	}

	var size int64 = fileInfo.Size()
	file.Close()

	err = os.Chown(urlFullFilename, b.chownTo, b.chownTo)
	if err != nil {
		b.rtm.SendMessage(b.rtm.NewOutgoingMessage(fmt.Sprintf("%s <@%s> I failed to chown `%s` to `%d` err=%+#v", FAILURE_EMOJI, user.Name, path.Base(urlFullFilename), b.chownTo, err), channel))
		logrus.WithFields(logrus.Fields{
			"urlFullFilename": urlFullFilename,
			"chownTo":         b.chownTo,
			"err":             err,
		}).Errorf("chown")
		return
	}

	b.rtm.SendMessage(b.rtm.NewOutgoingMessage(fmt.Sprintf("%s <@%s> I've processed `%s` which is `%d` bytes and downloadable from %s", SUCCESS_EMOJI, user.Name, path.Base(urlFullFilename), size, resultingItem.DownloadURI), channel))

	if upload {
		params := slack.FileUploadParameters{
			Title:          resultingItem.Title,
			File:           urlFullFilename,
			Channels:       []string{channel},
			InitialComment: fmt.Sprintf("%s <@%s> I've uploaded `%s`", SUCCESS_EMOJI, user.Name, urlFullFilename),
		}
		_, err := b.api.UploadFile(params)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"urlFullFilename": urlFullFilename,
				"err":             err,
			}).Errorf("file upload")
			return
		}
	}
}

func (b *Bot) processMessage(ev *slack.MessageEvent) {
	if strings.Contains(ev.Text, "uploaded a file") {
		return
	}

	user, err := b.api.GetUserInfo(ev.User)
	if err != nil {
		logrus.Errorf("error retrieving user info %v", err)
		return
	}

	for _, pat := range b.patterns {
		matched, args, err := pat.pattern.Match(ev.Text)

		if err != nil {
			logrus.WithFields(logrus.Fields{
				"err": err,
			}).Errorf("error processing command")
			continue
		}

		logrus.WithFields(logrus.Fields{
			"pattern":  pat.pattern,
			"matched":  matched,
			"action":   pat.action,
			"args_len": len(args),
		}).Debugf("pattern match")

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

	b.rtm.SendMessage(b.rtm.NewOutgoingMessage(fmt.Sprintf("%s <@%s> I'll process %d url", SUCCESS_EMOJI, user.Name, len(uris)), ev.Channel))

	for _, parsedUri := range uris {
		go b.processUri(parsedUri, user, ev.Channel, false)
	}
}

func (b *Bot) processFileUpload(ev *slack.FileSharedEvent) {
	logrus.Debugf("file share event %+v ev.File.URLPrivateDownload=%s\n", ev, ev.File.URLPrivate)

	file, _, _, err := b.api.GetFileInfo(ev.File.ID, 0, 0)
	if err != nil {
		fmt.Printf("error getting shared public url %+v\n", err)
		return
	}

	client := &http.Client{}
	req, _ := http.NewRequest("GET", file.URLPrivate, nil)
	req.Header.Set("Authorization", "Bearer "+b.slackToken)
	resp, err := client.Do(req)
	if err != nil {
		logrus.Errorf("ERROR: Failed to scrape %s", file.URLPrivate)
		return
	}

	body, _ := ioutil.ReadAll(resp.Body)
	bodyStr := string(body)

	if file.Title == "urls" {
		urls := URL_REGEXP.FindAllString(bodyStr, -1)

		logrus.Debugf("parsed %d urls %+v", len(urls), urls)

		for _, parsedUrl := range urls {
			newUUID := uuid.Must(uuid.NewV4())
			unescapedUrl, _ := url.QueryUnescape(parsedUrl)
			contentRequest := pb.Request{
				Uri:         unescapedUrl,
				Type:        pb.Request_URI,
				Title:       newUUID.String(),
				SubmittedBy: ev.File.User,
				SubmittedAt: uint64(time.Now().UTC().Unix()),
				RequestHash: newUUID.String(),
			}

			b.requestProcessor.Process(contentRequest)
		}
	} else {

		logrus.Debugf("treating upload as text")

		newUUID := uuid.Must(uuid.NewV4())

		contentRequest := pb.Request{
			Uri:         file.URLPrivate,
			Title:       file.Title,
			Type:        pb.Request_TEXT,
			Text:        bodyStr,
			SubmittedBy: ev.File.User,
			SubmittedAt: uint64(time.Now().UTC().Unix()),
			RequestHash: newUUID.String(),
		}

		b.requestProcessor.Process(contentRequest)
	}
}

func (b *Bot) slackReadLoop() {
	logrus.Debugf("start slack read loop")
Loop:
	for {
		select {
		case msg := <-b.rtm.IncomingEvents:
			logrus.Debugf("event received %#v", msg)

			switch ev := msg.Data.(type) {
			case *slack.ConnectedEvent:
				logrus.Infof("Connection counter: %d", ev.ConnectionCount)
			case *slack.MessageEvent:
				logrus.Debugf("message event received!")
				go b.processMessage(ev)
			case *slack.FileSharedEvent:
				logrus.Debugf("file shared event recieved!")
				go b.processFileUpload(ev)
			case *slack.RTMError:
				logrus.Errorf("RTMError: %s", ev.Error())
			case *slack.InvalidAuthEvent:
				logrus.Errorf("Invalid credentials!")
				break Loop
			default:
				//Take no action
			}
		}
	}

	logrus.Debugf("slack read loop finished")
}

func (b *Bot) joinChannels() {
	b.slackChannels = make([]*slack.Channel, len(b.channels))
	for _, chanStr := range b.channels {
		channel, err := b.api.JoinChannel(chanStr)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"channel": chanStr,
				"err":     err,
			}).Errorf("failed to join channel")
		} else {
			logrus.WithFields(logrus.Fields{
				"channel": chanStr,
			}).Debugf("joined channel")
			b.slackChannels = append(b.slackChannels, channel)
		}
	}
}

func (b *Bot) Start() {
	logrus.Debugf("connecting to slack")

	b.api = slack.New(
		b.slackToken,
	)

	b.rtm = b.api.NewRTM()
	go b.rtm.ManageConnection()
	go b.slackReadLoop()

	b.joinChannels()
}

func (b *Bot) getConfig(key string, user *slack.User, channel string) {
	found, val := b.requestProcessor.GetConfig(key)

	logrus.WithFields(logrus.Fields{
		"found": found,
		"val":   val,
	}).Debugf("get setting")

	if found {
		b.rtm.SendMessage(b.rtm.NewOutgoingMessage(fmt.Sprintf("%s <@%s> config property %s=%s", SUCCESS_EMOJI, user.Name, key, val), channel))
	} else {
		b.rtm.SendMessage(b.rtm.NewOutgoingMessage(fmt.Sprintf("%s <@%s> config property %s not found", FAILURE_EMOJI, user.Name, key), channel))
	}
}

func (b *Bot) setConfig(key string, val string, user *slack.User, channel string) {
	result := b.requestProcessor.SetConfig(key, val)

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

func init() {
	URL_REGEXP = regexp.MustCompile("(?m)(http[^ <>\n]*)")
}
