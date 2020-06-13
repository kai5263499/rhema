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
	pb "github.com/kai5263499/rhema/generated"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

type speedUpAudioTest struct {
	request pb.Request
	wanted  pb.Request
}

var processSpeedUpCmdInput func(args []string) (string, int)

func fakeSUAExecCommand(command string, args ...string) *exec.Cmd {
	cs := []string{"-test.run=TestSUAExecCommandHelper", "--", command}
	cs = append(cs, args...)
	cmd := exec.Command(os.Args[0], cs...)

	mockedStdout, es := processSpeedUpCmdInput(append([]string{command}, args...))

	cmd.Env = []string{"GO_WANT_HELPER_PROCESS=1",
		"STDOUT=" + mockedStdout,
		fmt.Sprintf("EXIT_STATUS=%d", es)}
	return cmd
}

func TestSUAExecCommandHelper(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}

	fmt.Fprintf(os.Stdout, os.Getenv("STDOUT"))
	i, _ := strconv.Atoi(os.Getenv("EXIT_STATUS"))
	os.Exit(i)
}

var _ = Describe("speedup_audio", func() {
	It("Should perform a basic conversion", func() {
		var err error

		processSpeedUpCmdInput = func(args []string) (cmdOutput string, exitStatus int) {
			return "all ok", 0
		}

		newUUID := uuid.Must(uuid.NewV4())

		tm := SpeedupAudio{
			localPath:   "/tmp",
			execCommand: fakeSUAExecCommand,
			atempo:      2.0,
		}

		requestContent := "this is a test"

		ci := pb.Request{
			Created: 383576400,
			Type:    pb.Request_AUDIO,
			Title:   newUUID.String(),
			Length:  uint64(len(requestContent)),
			Text:    requestContent,
		}

		slowFilename, err := GetFilePath(ci)
		Expect(err).To(BeNil())

		slowFullFilename := filepath.Join(tm.localPath, slowFilename)
		tmpFullFilename := fmt.Sprintf("%s%s", slowFullFilename[:len(slowFullFilename)-4], "-TMP.mp3")

		err = os.MkdirAll(path.Dir(slowFullFilename), os.ModePerm)
		Expect(err).To(BeNil())

		err = ioutil.WriteFile(slowFullFilename, []byte("test"), 0644)
		Expect(err).To(BeNil())

		err = ioutil.WriteFile(tmpFullFilename, []byte("test"), 0644)
		Expect(err).To(BeNil())

		audioRequest, err := tm.Convert(ci)
		Expect(err).To(BeNil())
		Expect(audioRequest.Type).To(Equal(pb.Request_AUDIO))
	})
})
