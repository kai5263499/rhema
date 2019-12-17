package rhema

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/kai5263499/rhema/domain"
	pb "github.com/kai5263499/rhema/generated"
	"github.com/nlopes/slack"
	"github.com/sirupsen/logrus"
)

func NewBot(slackToken string, requestProcessor domain.Processor) *Bot {
	return &Bot{
		slackToken:       slackToken,
		requestProcessor: requestProcessor,
	}
}

var (
	URL_REGEXP *regexp.Regexp
)

type Bot struct {
	api              *slack.Client
	rtm              *slack.RTM
	slackToken       string
	requestProcessor domain.Processor
}

func (b *Bot) processMessage(ev *slack.MessageEvent) {
	logrus.Debugf("Message: %+v", ev)

	if strings.Contains(ev.Text, "uploaded a file") {
		return
	}

	urls := URL_REGEXP.FindAllString(ev.Text, -1)

	for _, parsedUrl := range urls {
		contentRequest := &pb.Request{
			Uri:         parsedUrl,
			Type:        pb.Request_URI,
			SubmittedBy: ev.User,
			SubmittedAt: uint64(time.Now().UTC().Unix()),
		}
		// TODO: Process request
		logrus.Debugf("contentRequest=%#+v", contentRequest)
	}

	user, err := b.api.GetUserInfo(ev.User)
	if err != nil {
		logrus.Errorf("error retrieving user info %v", err)
		return
	}

	if len(urls) > 0 {
		b.rtm.SendMessage(b.rtm.NewOutgoingMessage(fmt.Sprintf("ok <@%s> I'll process %d url", user.Name, len(urls)), ev.Channel))
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
			unescapedUrl, _ := url.QueryUnescape(parsedUrl)
			contentRequest := &pb.Request{
				Uri:         unescapedUrl,
				Type:        pb.Request_URI,
				SubmittedBy: ev.File.User,
				SubmittedAt: uint64(time.Now().UTC().Unix()),
			}

			// TODO: Process request
			logrus.Debugf("contentRequest=%#+v", contentRequest)
		}
	} else {

		contentRequest := &pb.Request{
			Uri:         file.URLPrivate,
			Title:       file.Title,
			Type:        pb.Request_TEXT,
			Text:        bodyStr,
			SubmittedBy: ev.File.User,
			SubmittedAt: uint64(time.Now().UTC().Unix()),
		}

		// TODO: Process request
		logrus.Debugf("contentRequest=%#+v", contentRequest)
	}
}

func (b *Bot) slackReadLoop() {
Loop:
	for {
		select {
		case msg := <-b.rtm.IncomingEvents:
			logrus.Debugf("event received %#v", msg)

			switch ev := msg.Data.(type) {
			case *slack.ConnectedEvent:
				logrus.Infof("Connection counter: %d", ev.ConnectionCount)
			case *slack.MessageEvent:
				go b.processMessage(ev)
			case *slack.FileSharedEvent:
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

	logrus.Infof("bot read loop finished")
}

func (b *Bot) Start() {
	logrus.Debugf("connecting to slack")

	b.api = slack.New(b.slackToken)
	b.api.SetUserAsActive()
	b.rtm = b.api.NewRTM()
	go b.rtm.ManageConnection()
	go b.slackReadLoop()
}

func init() {
	URL_REGEXP = regexp.MustCompile("(?m)(http[^ <>\n]*)")
}
