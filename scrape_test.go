package rhema

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"

	"github.com/kai5263499/rhema/domain"
	pb "github.com/kai5263499/rhema/generated"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

type htmlScrapeTest struct {
	request      *pb.Request
	dataFilename string
	wanted       *pb.Request
}

var _ = Describe("scrape", func() {
	It("Should perform a basic scrape", func() {
		tmpDir, err := ioutil.TempDir("", "rhema-scrape-test-")
		Expect(err).To(BeNil())
		defer os.RemoveAll(tmpDir)

		testText := "This should come from a file and contain real messy HTML examples"

		scrape := Scrape{
			cfg: &domain.Config{
				TmpPath: tmpDir,
			},
		}

		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintln(w, testText)
		}))
		defer ts.Close()

		ci := &pb.Request{
			Created: 383576400,
			Type:    pb.ContentType_URI,
			Uri:     ts.URL,
			Title:   "my title",
		}

		// Process classifies URI requests before passing them to the scraper.
		processor := &RequestProcessor{}
		ci.Type = processor.parseRequestTypeFromURI(ci.Uri)
		Expect(ci.Type).To(Equal(pb.ContentType_TEXT))
		err = scrape.Convert(ci)
		Expect(err).To(BeNil())

		Expect(ci.Text).To(Not(BeNil()))
		Expect(ci.Title).To(Equal("my title"))
		Expect(ci.Text).To(Equal(testText))
	})
	It("Should perform complex scrapes using stored HTML", func() {
		tmpDir, err := ioutil.TempDir("", "rhema-scrape-html-test-")
		Expect(err).To(BeNil())
		defer os.RemoveAll(tmpDir)

		tests := []htmlScrapeTest{
			{
				request: &pb.Request{
					Created:     383576400,
					RequestHash: "ABC123",
					Type:        pb.ContentType_URI,
				},
				dataFilename: "testdata/ito.html",
				wanted: &pb.Request{
					Created:     383576400,
					RequestHash: "ABC123",
					Type:        pb.ContentType_TEXT,
					Title:       "Frozen II: Saved by Blessedly Superficial Viewers",
					Text:        "someText",
				},
			},
		}

		scrape := Scrape{
			cfg: &domain.Config{
				TmpPath:          tmpDir,
				TitleLengthLimit: 120,
			},
		}

		for _, tc := range tests {

			content, err := ioutil.ReadFile(tc.dataFilename)
			Expect(err).To(BeNil())

			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				fmt.Fprintln(w, string(content))
			}))

			tc.request.Uri = ts.URL
			processor := &RequestProcessor{}
			tc.request.Type = processor.parseRequestTypeFromURI(tc.request.Uri)
			Expect(tc.request.Type).To(Equal(pb.ContentType_TEXT))

			err = scrape.Convert(tc.request)
			ts.Close()

			Expect(err).To(BeNil())

			Expect(tc.request.Text).To(Not(BeNil()))
			Expect(tc.request.Text).ToNot(BeEmpty())
			Expect(tc.request.Title).To(Equal(tc.wanted.Title))
			Expect(tc.request.Type).To(Equal(tc.wanted.Type))
			Expect(tc.request.RequestHash).To(Equal(tc.wanted.RequestHash))
		}
	})
})
