package rhema

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"strconv"
	"testing"

	"github.com/gofrs/uuid"
	pb "github.com/kai5263499/rhema/generated"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

type rpTest struct {
	request pb.Request
	wanted  pb.Request
}

type rpUriTest struct {
	uri          string
	expectedType pb.ContentType
}

var processRPCmdInput func(args []string) (string, int)

func fakeRPExecCommand(command string, args ...string) *exec.Cmd {
	cs := []string{"-test.run=TestRPExecCommandHelper", "--", command}
	cs = append(cs, args...)
	cmd := exec.Command(os.Args[0], cs...)

	mockedStdout, es := processRPCmdInput(append([]string{command}, args...))

	cmd.Env = []string{"GO_WANT_HELPER_PROCESS=1",
		"STDOUT=" + mockedStdout,
		fmt.Sprintf("EXIT_STATUS=%d", es)}
	return cmd
}

func TestRPExecCommandHelper(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}

	fmt.Fprintf(os.Stdout, os.Getenv("STDOUT"))
	i, _ := strconv.Atoi(os.Getenv("EXIT_STATUS"))
	os.Exit(i)
}

var _ = Describe("request_processor", func() {

	var (
		rp RequestProcessor
	)

	BeforeEach(func() {
		rp = RequestProcessor{
			youtube:      &youtubeConverter{},
			scrape:       &textConverter{},
			text2mp3:     &audioConverter{},
			speedupAudio: &audioConverter{},
			localPath:    "/tmp",
		}
	})

	It("Should perform a basic processing", func() {
		var err error

		processRPCmdInput = func(args []string) (cmdOutput string, exitStatus int) {
			return "all ok", 0
		}

		newUUID := uuid.Must(uuid.NewV4())

		testText := "This should come from a file and contain real messy HTML examples"

		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintln(w, testText)
		}))
		defer ts.Close()

		ci := pb.Request{
			Created:     383576400,
			Type:        pb.ContentType_URI,
			Title:       newUUID.String(),
			RequestHash: "DEADBEEF",
			Uri:         ts.URL,
		}

		requestResult, err := rp.Process(ci)
		Expect(err).To(BeNil())
		Expect(requestResult.Type).To(Equal(pb.ContentType_AUDIO))
	})

	It("Should parse the right type from a URI", func() {
		tests := []rpUriTest{
			{
				uri:          "http://something.com/my_podcast.mp3",
				expectedType: pb.ContentType_AUDIO,
			},
			{
				uri:          "http://something.com/my_podcast.mp3?this=is&some=parameter&trash=bad",
				expectedType: pb.ContentType_AUDIO,
			},
		}

		for _, t := range tests {
			parsedType := rp.parseRequestTypeFromURI(t.uri)
			Expect(parsedType).To(Equal(t.expectedType))
		}

	})
})
