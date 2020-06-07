package rhema

import (
	"crypto/sha256"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	pb "github.com/kai5263499/rhema/generated"
	"github.com/sirupsen/logrus"
)

var (
	titleCleaningRegexp *regexp.Regexp
)

func cleanTitle(title string) string {
	return titleCleaningRegexp.ReplaceAllString(strings.Replace(title, " ", "_", -1), "")
}

func getExtFromType(contentType pb.Request_ContentType) (string, error) {
	switch contentType {
	case pb.Request_URI:
		return "uri", nil
	case pb.Request_AUDIO:
		return "mp3", nil
	case pb.Request_TEXT:
		return "txt", nil
	default:
		return "", fmt.Errorf("unknown extension for type %v", contentType)
	}
}

func getExtFromUri(requestUri string) (string, error) {
	logrus.Debugf("parsing ext from uri %s\n", requestUri)
	u, err := url.ParseRequestURI(requestUri)
	if err != nil {
		return "", err
	}

	uriExt := filepath.Ext(u.EscapedPath())

	return uriExt[1:], nil
}

func parseTitleFromUri(requestUri string) (string, error) {
	logrus.Debugf("parsing title from uri %s\n", requestUri)
	u, err := url.ParseRequestURI(requestUri)
	if err != nil {
		return "", err
	}

	titleStub := strings.Replace(filepath.Base(u.EscapedPath()), filepath.Ext(u.EscapedPath()), "", -1)

	logrus.Debugf("parsed titleStub %s\n", titleStub)

	return titleStub, nil
}

func hasContent(title string) bool {
	return len(cleanTitle(title)) > 0
}

func getPath(req pb.Request) (string, error) {
	var err error

	createdTime := time.Unix(int64(req.Created), 0)
	cleanTitle := cleanTitle(req.Title)

	if len(cleanTitle) < 2 {
		return "", fmt.Errorf("title length too small")
	}

	ext, err := getExtFromType(req.Type)
	if err != nil {
		ext, err = getExtFromUri(req.Uri)

		if err != nil {
			return "", fmt.Errorf("could not parse ext from either type %s or uri %s", req.Type.String(), req.Uri)
		}
	}

	return fmt.Sprintf("%s/%s/%s", req.Type.String(), createdTime.Format("2006/01/02"), fmt.Sprintf("%s.%s", cleanTitle, ext)), nil
}

// GetFilePath returns the filename for a given Request
func GetFilePath(req pb.Request) (string, error) {
	ext, err := getExtFromType(req.Type)
	if err != nil {
		ext, err = getExtFromUri(req.Uri)

		if err != nil {
			return "", fmt.Errorf("could not parse ext from either type %s or uri %s", req.Type.String(), req.Uri)
		}
	}

	createdTime := time.Unix(int64(req.Created), 0)
	cleanTitle := cleanTitle(req.Title)

	return filepath.Join(req.Type.String(), createdTime.Format("2006/01/02"), fmt.Sprintf("%s.%s", cleanTitle, ext)), nil
}

func getHash(thingToHash string) string {
	h := sha256.New()
	h.Write([]byte(thingToHash))
	bs := h.Sum(nil)

	return fmt.Sprintf("%x", bs)
}

// DownloadUriToFile downloads the contents of a uri to a local file
func DownloadUriToFile(uri string, uriFullFilename string) error {
	var err error
	if err = os.MkdirAll(path.Dir(uriFullFilename), os.ModePerm); err != nil {
		return err
	}

	logrus.Debugf("downloading %s to %s", uri, uriFullFilename)

	resp, err := http.Get(uri)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	out, err := os.Create(uriFullFilename)
	if err != nil {
		return err
	}

	defer out.Close()
	io.Copy(out, resp.Body)

	return nil
}

func init() {
	titleCleaningRegexp = regexp.MustCompile("[^a-zA-Z0-9\\-_]+")
}
