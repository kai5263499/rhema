package rhema

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"

	pb "github.com/kai5263499/rhema/generated"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

type htmlScrapeTest struct {
	request      pb.Request
	dataFilename string
	wanted       pb.Request
}

var _ = Describe("scrape", func() {
	It("Should perform a basic scrape", func() {

		testText := "This should come from a file and contain real messy HTML examples"

		scrape := Scrape{
			localPath:      "/tmp",
			contentStorage: &fakeContentStore{},
		}

		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintln(w, testText)
		}))
		defer ts.Close()

		ci := pb.Request{
			Created: 383576400,
			Type:    pb.Request_URI,
			Uri:     ts.URL,
			Title:   "my title",
		}

		textRequest, err := scrape.Convert(ci)
		Expect(err).To(BeNil())

		Expect(textRequest.Text).To(Not(BeNil()))
		Expect(textRequest.Title).To(Equal("my title"))
		Expect(textRequest.Text).To(Equal(testText))
	})
	It("Should perform complex scrapes using stored HTML", func() {

		tests := []htmlScrapeTest{
			{
				request: pb.Request{
					Created:     383576400,
					RequestHash: "ABC123",
					Type:        pb.Request_URI,
				},
				dataFilename: "testdata/ito.html",
				wanted: pb.Request{
					Created:     383576400,
					RequestHash: "ABC123",
					Type:        pb.Request_TEXT,
					Title:       "Frozen II: Saved by Blessedly Superficial Viewers",
					Text:        "someText",
				},
			},
		}

		scrape := Scrape{
			localPath:      "/tmp",
			contentStorage: &fakeContentStore{},
		}

		for _, tc := range tests {

			content, err := ioutil.ReadFile(tc.dataFilename)
			Expect(err).To(BeNil())

			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				fmt.Fprintln(w, string(content))
			}))

			tc.request.Uri = ts.URL

			textRequest, err := scrape.Convert(tc.request)
			ts.Close()

			Expect(err).To(BeNil())

			Expect(textRequest.Text).To(Not(BeNil()))
			Expect(textRequest.Title).To(Equal(tc.wanted.Title))
			Expect(textRequest.Type).To(Equal(tc.wanted.Type))
			Expect(textRequest.RequestHash).To(Equal(tc.wanted.RequestHash))
		}
	})
})
