package rhema

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/icza/gox/stringsx"
	"github.com/sirupsen/logrus"

	"github.com/kai5263499/rhema/domain"
	pb "github.com/kai5263499/rhema/generated"
)

var _ domain.Processor = (*RequestProcessor)(nil)

var (
	processingKeyTTL = 30 * time.Second
)

func NewRequestProcessor(cfg *domain.Config, scrape domain.Converter, youtube domain.Converter, text2mp3 domain.Converter, speedupAudio domain.Converter, contentStorage domain.Storage) *RequestProcessor {
	return &RequestProcessor{
		cfg:            cfg,
		youtube:        youtube,
		scrape:         scrape,
		text2mp3:       text2mp3,
		speedupAudio:   speedupAudio,
		contentStorage: contentStorage,
	}
}

type RequestProcessor struct {
	cfg            *domain.Config
	youtube        domain.Converter
	scrape         domain.Converter
	text2mp3       domain.Converter
	speedupAudio   domain.Converter
	contentStorage domain.Storage
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

func (rp *RequestProcessor) downloadUri(ci *pb.Request) error {
	urlFilename, err := GetFilePath(ci)
	if err != nil {
		return err
	}

	urlFullFilename := filepath.Join(rp.cfg.TmpPath, urlFilename)

	if err := DownloadUriToFile(ci.Uri, urlFullFilename); err != nil {
		return err
	}

	logrus.Debug("done downloading")

	return nil
}

func (rp *RequestProcessor) Process(ci *pb.Request) (err error) {
	if ci.Type == pb.ContentType_URI {
		ci.Type = rp.parseRequestTypeFromURI(ci.Uri)

		parsedTitle, err := parseTitleFromUri(ci.Uri)

		if err != nil {
			logrus.WithFields(logrus.Fields{
				"err": err,
				"uri": ci.Uri,
			}).Warn("error parsing title from uri")
		} else if len(parsedTitle) > 4 {
			if len(ci.Title) < 1 {
				ci.Title = stringsx.Clean(parsedTitle)
			}
		} else {
			logrus.WithFields(logrus.Fields{
				"err":         err,
				"uri":         ci.Uri,
				"parsedTitle": parsedTitle,
			}).Warn("parsed title too short")
		}

		if len(ci.Title) > int(rp.cfg.TitleLengthLimit) {
			ci.Title = stringsx.Clean(ci.Title[0:rp.cfg.TitleLengthLimit])

			logrus.WithFields(logrus.Fields{
				"err":      err,
				"uri":      ci.Uri,
				"ci.Title": ci.Title,
			}).Warn("parsed title too long")
		}
	}

	logrus.WithFields(logrus.Fields{
		"uri":   ci.Uri,
		"title": ci.Title,
		"type":  ci.Type,
	}).Info("processing")

	switch ci.Type {
	case pb.ContentType_YOUTUBE:
		logrus.Debugf("converting youtube")
		err = rp.youtube.Convert(ci)
		if err != nil {
			logrus.WithError(err).Error("error with youtube")
			return
		}

		err = rp.speedupAudio.Convert(ci)
		if err != nil {
			logrus.WithError(err).Error("error with youtube audio")
			return
		}

		if err = rp.contentStorage.Store(ci); err != nil {
			logrus.WithError(err).Error("error storing item")
			return
		}

		return
	case pb.ContentType_TEXT:
		if len(ci.Text) < 1 {
			if err = rp.scrape.Convert(ci); err != nil {
				logrus.WithError(err).Error("error with text")
				return
			}
		}

		if len(ci.Title) < 1 {
			// Use the first 64 characters from the text as the title
			ci.Title = ci.Text[:64]
		}

		if ci.Length < 1 {
			ci.Length = uint64(len(ci.Text))
		}

		if ci.Size < 1 {
			ci.Length = uint64(len(ci.Text))
		}

		var localFilename string
		localFilename, err = GetFilePath(ci)
		if err != nil {
			return
		}

		fullFilename := filepath.Join(rp.cfg.TmpPath, localFilename)

		err = os.MkdirAll(path.Dir(fullFilename), os.ModePerm)
		if err != nil {
			return err
		}

		if err = ioutil.WriteFile(fullFilename, []byte(ci.Text), 0644); err != nil {
			return err
		}

		if err = rp.text2mp3.Convert(ci); err != nil {
			logrus.WithError(err).Error("error with text to audio conversion")
			return
		}

		if err = rp.contentStorage.Store(ci); err != nil {
			logrus.WithError(err).Error("error storing item")
			return
		}

		return
	case pb.ContentType_AUDIO:
		if err = rp.downloadUri(ci); err != nil {
			logrus.WithError(err).Error("error downloading audio uri")
			return
		}

		if err = rp.speedupAudio.Convert(ci); err != nil {
			logrus.WithError(err).Error("error speeding up audio")
			return
		}

		if err = rp.contentStorage.Store(ci); err != nil {
			logrus.WithError(err).Error("error storing item")
			return
		}

		return
	case pb.ContentType_VIDEO:
		if err = rp.downloadUri(ci); err != nil {
			logrus.WithError(err).Error("error downloading video uri")
			return
		}

		if err = rp.speedupAudio.Convert(ci); err != nil {
			logrus.WithError(err).Errorf("error speeding up video")
			return
		}

		if err = rp.contentStorage.Store(ci); err != nil {
			logrus.WithError(err).Error("error storing item")
			return
		}

		return
	default:
		err = fmt.Errorf("unknown content type %s", ci.Type.String())
		return
	}
}
