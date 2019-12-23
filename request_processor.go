package rhema

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/sirupsen/logrus"

	"github.com/kai5263499/rhema/domain"
	pb "github.com/kai5263499/rhema/generated"
)

var _ domain.Processor = (*RequestProcessor)(nil)

func NewRequestProcessor(localPath string, scrape domain.Converter, youtube domain.Converter, text2mp3 domain.Converter, speedupAudio domain.Converter) *RequestProcessor {
	return &RequestProcessor{
		youtube:      youtube,
		scrape:       scrape,
		text2mp3:     text2mp3,
		speedupAudio: speedupAudio,
		localPath:    localPath,
	}
}

type RequestProcessor struct {
	youtube      domain.Converter
	scrape       domain.Converter
	text2mp3     domain.Converter
	speedupAudio domain.Converter
	localPath    string
}

func (rp *RequestProcessor) parseRequestTypeFromURI(requestUri string) pb.Request_ContentType {
	if strings.Contains(requestUri, "youtu.be") ||
		strings.Contains(requestUri, "www.youtube.com") ||
		strings.Contains(requestUri, "facebook.com") {
		if strings.Contains(requestUri, "playlist") {
			return pb.Request_YOUTUBE_LIST
		} else {
			return pb.Request_YOUTUBE
		}
	} else if strings.Contains(requestUri, ".mp3") {
		return pb.Request_AUDIO
	} else if strings.Contains(requestUri, ".mp4") {
		return pb.Request_VIDEO
	} else if strings.Contains(requestUri, ".pdf") {
		return pb.Request_PDF
	}

	return pb.Request_TEXT
}

func (rp *RequestProcessor) downloadUri(ci pb.Request) error {
	urlFilename, err := GetFilePath(ci)
	if err != nil {
		return err
	}

	urlFullFilename := filepath.Join(rp.localPath, urlFilename)

	err = DownloadUriToFile(ci.Uri, urlFullFilename)
	if err != nil {
		return err
	}

	logrus.Debugf("done downloading\n")

	return nil
}

func (rp *RequestProcessor) Process(ci pb.Request) (pb.Request, error) {
	logrus.Debugf("processing %#v\n", ci)

	var err error
	var ci2 pb.Request
	var ci3 pb.Request

	ci.Type = rp.parseRequestTypeFromURI(ci.Uri)
	parsedTitle, err := parseTitleFromUri(ci.Uri)
	if err != nil && len(parsedTitle) > 4 {
		ci.Title = parsedTitle
	}

	switch ci.Type {
	case pb.Request_YOUTUBE:
		ci2, err = rp.youtube.Convert(ci)
		if err != nil {
			return ci, err
		}

		ci3, err = rp.speedupAudio.Convert(ci2)
		if err != nil {
			return ci2, err
		}

		return ci3, nil
	case pb.Request_TEXT:
		ci2, err = rp.scrape.Convert(ci)
		if err != nil {
			return ci, err
		}

		ci3, err = rp.text2mp3.Convert(ci2)
		if err != nil {
			return ci2, err
		}

		return ci3, nil
	case pb.Request_AUDIO:
		err = rp.downloadUri(ci)
		if err != nil {
			return ci, err
		}

		ci2, err = rp.speedupAudio.Convert(ci)
		if err != nil {
			return ci, err
		}

		return ci2, nil
	case pb.Request_VIDEO:
		err = rp.downloadUri(ci)
		if err != nil {
			return ci, err
		}

		ci2, err = rp.speedupAudio.Convert(ci)
		if err != nil {
			return ci, err
		}

		return ci2, nil
	default:
		return ci, fmt.Errorf("unknown content type %s", ci.Type.String())
	}
}
