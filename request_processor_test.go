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
	It("Should perform a basic processing", func() {
		var err error

		processRPCmdInput = func(args []string) (cmdOutput string, exitStatus int) {
			return "all ok", 0
		}

		newUUID := uuid.Must(uuid.NewV4())

		rp := RequestProcessor{
			youtube:      &youtubeConverter{},
			scrape:       &textConverter{},
			text2mp3:     &audioConverter{},
			speedupAudio: &audioConverter{},
			localPath:    "/tmp",
		}

		testText := "This should come from a file and contain real messy HTML examples"

		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintln(w, testText)
		}))
		defer ts.Close()

		ci := pb.Request{
			Created:     383576400,
			Type:        pb.Request_URI,
			Title:       newUUID.String(),
			RequestHash: "DEADBEEF",
			Uri:         ts.URL,
		}

		requestResult, err := rp.Process(ci)
		Expect(err).To(BeNil())
		Expect(requestResult.Type).To(Equal(pb.Request_AUDIO))
	})
})
