package rhema

import (
	"fmt"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/sirupsen/logrus"

	"github.com/kai5263499/rhema/domain"
	pb "github.com/kai5263499/rhema/generated"
)

var _ domain.Processor = (*RequestProcessor)(nil)

func NewRequestProcessor(localPath string, scrape domain.Converter, youtube domain.Converter, text2mp3 domain.Converter, speedupAudio domain.Converter, titleLengthLimit int, comms domain.Comms) *RequestProcessor {
	return &RequestProcessor{
		youtube:          youtube,
		scrape:           scrape,
		text2mp3:         text2mp3,
		speedupAudio:     speedupAudio,
		localPath:        localPath,
		titleLengthLimit: titleLengthLimit,
		comms:            comms,
	}
}

type RequestProcessor struct {
	youtube          domain.Converter
	scrape           domain.Converter
	text2mp3         domain.Converter
	speedupAudio     domain.Converter
	localPath        string
	titleLengthLimit int
	comms            domain.Comms
}

func (rp *RequestProcessor) parseRequestTypeFromURI(requestUri string) pb.ContentType {
	if strings.Contains(requestUri, "youtu.be") ||
		strings.Contains(requestUri, "www.youtube.com") ||
		strings.Contains(requestUri, "facebook.com") ||
		strings.Contains(requestUri, "vimeo.com") {
		if strings.Contains(requestUri, "playlist") {
			return pb.ContentType_YOUTUBE_LIST
		} else {
			return pb.ContentType_YOUTUBE
		}
	} else if strings.Contains(requestUri, ".mp3") {
		return pb.ContentType_AUDIO
	} else if strings.Contains(requestUri, ".mp4") {
		return pb.ContentType_VIDEO
	} else if strings.Contains(requestUri, ".pdf") {
		return pb.ContentType_PDF
	}

	return pb.ContentType_TEXT
}

func (rp *RequestProcessor) downloadUri(ci pb.Request) error {
	urlFilename, err := GetFilePath(ci)
	if err != nil {
		return err
	}

	urlFullFilename := filepath.Join(rp.localPath, urlFilename)

	if err := DownloadUriToFile(ci.Uri, urlFullFilename); err != nil {
		return err
	}

	logrus.Debugf("done downloading\n")

	return nil
}

func (rp *RequestProcessor) Process(ci pb.Request) (pb.Request, error) {
	var err error
	var ci2 pb.Request
	var ci3 pb.Request

	if ci.Type == pb.ContentType_URI {
		ci.Type = rp.parseRequestTypeFromURI(ci.Uri)

		parsedTitle, err := parseTitleFromUri(ci.Uri)

		if err != nil {
			logrus.WithFields(logrus.Fields{
				"err": err,
				"uri": ci.Uri,
			}).Warnf("error parsing title from uri")
		} else if len(parsedTitle) > 4 {
			ci.Title = parsedTitle
		} else {
			logrus.WithFields(logrus.Fields{
				"err":         err,
				"uri":         ci.Uri,
				"parsedTitle": parsedTitle,
			}).Warnf("parsed title too short")
		}

		if len(ci.Title) > rp.titleLengthLimit {
			ci.Title = ci.Title[0:rp.titleLengthLimit]

			logrus.WithFields(logrus.Fields{
				"err":      err,
				"uri":      ci.Uri,
				"ci.Title": ci.Title,
			}).Warnf("parsed title too long")
		}
	}

	logrus.WithFields(logrus.Fields{
		"uri":   ci.Uri,
		"title": ci.Title,
		"type":  ci.Type,
	}).Infof("processing")

	switch ci.Type {
	case pb.ContentType_YOUTUBE:
		ci2, err = rp.youtube.Convert(ci)
		if err != nil {
			logrus.WithError(err).Errorf("error with youtube")
			return ci, err
		}

		ci3, err = rp.speedupAudio.Convert(ci2)
		if err != nil {
			logrus.WithError(err).Errorf("error with youtube audio")
			return ci2, err
		}

		rp.comms.SendRequest(ci3)

		return ci3, nil
	case pb.ContentType_TEXT:
		if len(ci.Text) < 1 {
			ci2, err = rp.scrape.Convert(ci)
			if err != nil {
				logrus.WithError(err).Errorf("error with text")
				return ci, err
			}
		} else {
			ci2 = ci
		}

		ci3, err = rp.text2mp3.Convert(ci2)
		if err != nil {
			logrus.WithError(err).Errorf("error with text to audio conversion")
			return ci2, err
		}

		rp.comms.SendRequest(ci3)

		return ci3, nil
	case pb.ContentType_AUDIO:
		err = rp.downloadUri(ci)
		if err != nil {
			logrus.WithError(err).Errorf("error downloading audio uri")
			return ci, err
		}

		ci2, err = rp.speedupAudio.Convert(ci)
		if err != nil {
			logrus.WithError(err).Errorf("error speeding up audio")
			return ci, err
		}

		rp.comms.SendRequest(ci2)

		return ci2, nil
	case pb.ContentType_VIDEO:
		err = rp.downloadUri(ci)
		if err != nil {
			logrus.WithError(err).Errorf("error downloading video uri")
			return ci, err
		}

		ci2, err = rp.speedupAudio.Convert(ci)
		if err != nil {
			logrus.WithError(err).Errorf("error speeding up video")
			return ci, err
		}

		rp.comms.SendRequest(ci2)

		return ci2, nil
	default:
		return ci, fmt.Errorf("unknown content type %s", ci.Type.String())
	}
}

func (rp *RequestProcessor) SetConfig(key string, value string) bool {
	ret1 := rp.youtube.SetConfig(key, value)
	ret2 := rp.scrape.SetConfig(key, value)
	ret3 := rp.text2mp3.SetConfig(key, value)
	ret4 := rp.speedupAudio.SetConfig(key, value)

	ret5 := false
	switch key {
	case "titlelengthlimit":
		v, err := strconv.Atoi(value)
		if err != nil {
			return false
		}
		rp.titleLengthLimit = v
	case "localpath":
		rp.localPath = value
		ret5 = true
	}

	return ret1 || ret2 || ret3 || ret4 || ret5
}

func (rp *RequestProcessor) GetConfig(key string) (bool, string) {
	var found bool
	var val string

	found, val = rp.youtube.GetConfig(key)
	if found {
		return found, val
	}

	found, val = rp.scrape.GetConfig(key)
	if found {
		return found, val
	}

	found, val = rp.text2mp3.GetConfig(key)
	if found {
		return found, val
	}

	found, val = rp.speedupAudio.GetConfig(key)
	if found {
		return found, val
	}

	switch key {
	case "localpath":
		return true, rp.localPath
	default:
		return false, ""
	}
}
