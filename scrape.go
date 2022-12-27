package rhema

import (
	"bytes"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/icza/gox/stringsx"
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
func NewScrape(cfg *domain.Config) *Scrape {
	return &Scrape{
		cfg: cfg,
	}
}

// Scrape represents everything needed to scrape content from a URL
type Scrape struct {
	cfg *domain.Config
}

func (s *Scrape) extractText(ci *pb.Request) (string, *bytes.Buffer, error) {
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

			if uint32(len(TxtContent)) > uint32(s.cfg.MinTextBlockSize) {
				buffer.WriteString(TxtContent)
			}
		}
	}

	return bow.Title(), buffer, nil
}

// Convert takes a URL, reads the content, stores the relevant body of the website,
// and returns a new TEXT
func (s *Scrape) Convert(ci *pb.Request) error {
	var err error

	title, bodyBuf, err := s.extractText(ci)
	if err != nil {
		return err
	}

	if len(title) > 3 && len(ci.Title) < 1 {
		if s.cfg.TitleLengthLimit > 0 && uint32(len(title)) > s.cfg.TitleLengthLimit {
			ci.Title = stringsx.Clean(title[:s.cfg.TitleLengthLimit])
		} else {
			ci.Title = stringsx.Clean(title)
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
		return err
	}

	fullFilename := filepath.Join(s.cfg.LocalPath, localFilename)
	if err != nil {
		return err
	}
	err = os.MkdirAll(path.Dir(fullFilename), os.ModePerm)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(fullFilename, []byte(bodyBuf.String()), 0644)
	if err != nil {
		return err
	}

	return nil
}
