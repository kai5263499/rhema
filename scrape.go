package rhema

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/kai5263499/rhema/domain"
	pb "github.com/kai5263499/rhema/generated"

	"github.com/headzoo/surf"
	"github.com/headzoo/surf/agent"
	"github.com/headzoo/surf/browser"
	"github.com/headzoo/surf/jar"
	"golang.org/x/net/html"
)

var _ domain.Converter = (*Scrape)(nil)

// NewScrape returns a new Scrape instance
func NewScrape(minTextBlockSize uint32, localPath string, titleLengthLimit int) *Scrape {
	return &Scrape{
		minTextBlockSize: minTextBlockSize,
		localPath:        localPath,
		titleLengthLimit: titleLengthLimit,
	}
}

// Scrape represents everything needed to scrape content from a URL
type Scrape struct {
	minTextBlockSize uint32
	localPath        string
	titleLengthLimit int
}

func (s *Scrape) extractText(ci pb.Request) (string, *bytes.Buffer, error) {
	buffer := new(bytes.Buffer)

	bow := surf.NewBrowser()
	bow.SetUserAgent(agent.Chrome())
	bow.SetAttribute(browser.SendReferer, true)
	bow.SetAttribute(browser.MetaRefreshHandling, true)
	bow.SetAttribute(browser.FollowRedirects, true)
	bow.SetCookieJar(jar.NewMemoryCookies())
	err := bow.Open(ci.Uri)
	if err != nil {
		return "", nil, err
	}

	domDocTest := html.NewTokenizer(strings.NewReader(bow.Body()))
	previousStartTokenTest := domDocTest.Token()

loopDomTest:
	for {
		tt := domDocTest.Next()
		switch {
		case tt == html.ErrorToken:
			break loopDomTest // End of the document,  done
		case tt == html.StartTagToken:
			previousStartTokenTest = domDocTest.Token()
		case tt == html.TextToken:
			prevStartToken := strings.ToLower(previousStartTokenTest.Data)
			if prevStartToken == "script" || prevStartToken == "style" || prevStartToken == "meta" || prevStartToken == "link" {
				continue
			}
			TxtContent := strings.TrimSpace(html.UnescapeString(string(domDocTest.Text())))

			if uint32(len(TxtContent)) > s.minTextBlockSize {
				buffer.WriteString(TxtContent)
			}
		}
	}

	return bow.Title(), buffer, nil
}

// Convert takes a URL, reads the content, stores the relevant body of the website,
// and returns a new TEXT
func (s *Scrape) Convert(ci pb.Request) (pb.Request, error) {
	var err error

	title, bodyBuf, err := s.extractText(ci)
	if err != nil {
		return ci, err
	}

	if len(title) > 3 {
		if s.titleLengthLimit > 0 && len(title) > s.titleLengthLimit {
			ci.Title = title[:s.titleLengthLimit]
		} else {
			ci.Title = title
		}
	}

	ci.Type = pb.ContentType_TEXT

	createdTime := time.Now()
	ci.Created = uint64(createdTime.Unix())
	ci.Text = bodyBuf.String()
	ci.Size = uint64(bodyBuf.Len())
	ci.Length = uint64(bodyBuf.Len())

	localFilename, err := GetFilePath(ci)
	if err != nil {
		return ci, err
	}

	fullFilename := filepath.Join(s.localPath, localFilename)
	if err != nil {
		return ci, err
	}
	err = os.MkdirAll(path.Dir(fullFilename), os.ModePerm)
	if err != nil {
		return ci, err
	}

	err = ioutil.WriteFile(fullFilename, []byte(bodyBuf.String()), 0644)
	if err != nil {
		return ci, err
	}

	return ci, nil
}

func (s *Scrape) SetConfig(key string, value string) bool {
	switch key {
	case "mintextblocksize":
		v, err := strconv.Atoi(value)
		if err != nil {
			return false
		}
		s.minTextBlockSize = uint32(v)
		return true
	case "localpath":
		s.localPath = value
		return true
	default:
		return false
	}
}

func (s *Scrape) GetConfig(key string) (bool, string) {
	switch key {
	case "mintextblocksize":
		return true, fmt.Sprintf("%d", s.minTextBlockSize)
	case "localpath":
		return true, s.localPath
	default:
		return false, ""
	}
}
