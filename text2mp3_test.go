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

type text2mp3Test struct {
	request pb.Request
	wanted  pb.Request
}

var processTMCmdInput func(args []string) (string, int)

func fakeExecCommand(command string, args ...string) *exec.Cmd {
	cs := []string{"-test.run=TestTMExecCommandHelper", "--", command}
	cs = append(cs, args...)
	cmd := exec.Command(os.Args[0], cs...)

	mockedStdout, es := processTMCmdInput(append([]string{command}, args...))

	cmd.Env = []string{"GO_WANT_HELPER_PROCESS=1",
		"STDOUT=" + mockedStdout,
		fmt.Sprintf("EXIT_STATUS=%d", es)}
	return cmd
}

func TestTMExecCommandHelper(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}

	fmt.Fprintf(os.Stdout, os.Getenv("STDOUT"))
	i, _ := strconv.Atoi(os.Getenv("EXIT_STATUS"))
	os.Exit(i)
}

var _ = Describe("text2mp3", func() {
	It("Should perform a basic conversion", func() {
		var err error

		processTMCmdInput = func(args []string) (cmdOutput string, exitStatus int) {
			return "all ok", 0
		}

		newUUID := uuid.Must(uuid.NewV4())

		tm := Text2Mp3{
			localPath:      "/tmp",
			execCommand:    fakeExecCommand,
			contentStorage: &fakeContentStore{},
			wordsPerMinute: 350,
			espeakVoice:    "f5",
		}

		requestContent := "this is a test"

		ci := pb.Request{
			Created: 383576400,
			Type:    pb.Request_TEXT,
			Title:   newUUID.String(),
			Length:  uint64(len(requestContent)),
			Text:    requestContent,
		}

		txtFilename, err := GetFilePath(ci)
		Expect(err).To(BeNil())

		txtFullFilename := filepath.Join(tm.localPath, txtFilename)

		ci.Type = pb.Request_AUDIO
		mp3FileName, err := GetFilePath(ci)
		Expect(err).To(BeNil())

		ci.Type = pb.Request_TEXT

		mp3FullFilename := filepath.Join(tm.localPath, mp3FileName)
		wavFullFilename := fmt.Sprintf("%s%s", mp3FullFilename[:len(mp3FullFilename)-3], "wav")

		err = os.MkdirAll(path.Dir(txtFullFilename), os.ModePerm)
		Expect(err).To(BeNil())

		err = ioutil.WriteFile(txtFullFilename, []byte("test"), 0644)
		Expect(err).To(BeNil())

		err = os.MkdirAll(path.Dir(wavFullFilename), os.ModePerm)
		Expect(err).To(BeNil())

		err = ioutil.WriteFile(wavFullFilename, []byte("test"), 0644)
		Expect(err).To(BeNil())

		err = os.MkdirAll(path.Dir(mp3FullFilename), os.ModePerm)
		Expect(err).To(BeNil())

		err = ioutil.WriteFile(mp3FullFilename, []byte("test"), 0644)
		Expect(err).To(BeNil())

		audioRequest, err := tm.Convert(ci)
		Expect(err).To(BeNil())
		Expect(audioRequest.Type).To(Equal(pb.Request_AUDIO))

		err = os.Remove(txtFullFilename)
		Expect(err).To(BeNil())

		// This should return an error because the convert function should clean up after itself
		err = os.Remove(wavFullFilename)
		Expect(err).To(Not(BeNil()))

		err = os.Remove(mp3FullFilename)
		Expect(err).To(BeNil())
	})
})
