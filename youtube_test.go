package rhema

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/gofrs/uuid"
	"github.com/kai5263499/rhema/domain"
	pb "github.com/kai5263499/rhema/generated"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

type ytTest struct {
	request *pb.Request
	wanted  *pb.Request
}

var processYTCmdInput func(args []string) (string, int)

func fakeYTExecCommand(command string, args ...string) *exec.Cmd {
	cs := []string{"-test.run=TestYTExecCommandHelper", "--", command}
	cs = append(cs, args...)
	cmd := exec.Command(os.Args[0], cs...)

	mockedStdout, es := processYTCmdInput(append([]string{command}, args...))

	cmd.Env = []string{"GO_WANT_HELPER_PROCESS=1",
		"STDOUT=" + mockedStdout,
		fmt.Sprintf("EXIT_STATUS=%d", es)}
	return cmd
}

func TestYTExecCommandHelper(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}

	fmt.Fprintf(os.Stdout, os.Getenv("STDOUT"))
	i, _ := strconv.Atoi(os.Getenv("EXIT_STATUS"))
	os.Exit(i)
}

var _ = Describe("youtube", func() {
	It("Should perform a basic conversion", func() {
		var err error

		processYTCmdInput = func(args []string) (cmdOutput string, exitStatus int) {
			return "all ok", 0
		}

		newUUID := uuid.Must(uuid.NewV4())

		yt := YouTube{
			cfg: &domain.Config{
				LocalPath: "/tmp",
			},
			execCommand:  fakeYTExecCommand,
			scrape:       &textConverter{},
			speedupAudio: &audioConverter{},
		}

		requestContent := "this is a test"

		ci := &pb.Request{
			Created:     383576400,
			Type:        pb.ContentType_AUDIO,
			Title:       newUUID.String(),
			RequestHash: "DEADBEEF",
			Length:      uint64(len(requestContent)),
			Text:        requestContent,
		}

		fileName, err := GetFilePath(ci)
		Expect(err).To(BeNil())

		mp3Filename := filepath.Join(yt.cfg.LocalPath, fileName)

		tmpFilename := fmt.Sprintf("%s%s", mp3Filename[:len(mp3Filename)-3], "flv")

		err = os.MkdirAll(path.Dir(mp3Filename), os.ModePerm)
		Expect(err).To(BeNil())

		err = ioutil.WriteFile(tmpFilename, []byte("test"), 0644)
		Expect(err).To(BeNil())

		err = ioutil.WriteFile(mp3Filename, []byte("test"), 0644)
		Expect(err).To(BeNil())

		err = os.MkdirAll(path.Dir(mp3Filename), os.ModePerm)
		Expect(err).To(BeNil())

		err = yt.Convert(ci)
		Expect(err).To(BeNil())
		Expect(ci.Type).To(Equal(pb.ContentType_AUDIO))
		Expect(ci.RequestHash).To(Equal(ci.RequestHash))

		err = os.Remove(tmpFilename)
		Expect(err).To(BeNil())

		// This should return an error because the convert function should clean up after itself
		err = os.Remove(mp3Filename)
		Expect(err).To(BeNil())
	})
})
