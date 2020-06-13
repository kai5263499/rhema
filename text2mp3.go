package rhema

import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strconv"

	"github.com/kai5263499/rhema/domain"
	pb "github.com/kai5263499/rhema/generated"
	"github.com/sirupsen/logrus"
)

var _ domain.Converter = (*Text2Mp3)(nil)

func NewText2Mp3(localPath string, wordsPerMinute int, espeakVoice string) *Text2Mp3 {
	return &Text2Mp3{
		localPath:      localPath,
		execCommand:    exec.Command,
		wordsPerMinute: wordsPerMinute,
		espeakVoice:    espeakVoice,
	}
}

type Text2Mp3 struct {
	localPath      string
	execCommand    func(command string, args ...string) *exec.Cmd
	wordsPerMinute int
	espeakVoice    string
}

func (tm *Text2Mp3) Convert(ci pb.Request) (pb.Request, error) {
	var err error

	logrus.Debugf("converting text2mp3\n")

	if ci.Type != pb.Request_TEXT {
		return ci, fmt.Errorf("Invalid request type %s", ci.Type.String())
	}

	if ci.Length < 3 {
		return ci, fmt.Errorf("Text length too short %#+v", ci)
	}

	txtFilename, err := GetFilePath(ci)
	if err != nil {
		return ci, err
	}

	txtFullFilename := filepath.Join(tm.localPath, txtFilename)

	ci.Type = pb.Request_AUDIO

	mp3Filename, err := GetFilePath(ci)
	if err != nil {
		return ci, err
	}

	mp3FullFilename := filepath.Join(tm.localPath, mp3Filename)
	wavFullFilename := fmt.Sprintf("%s%s", mp3FullFilename[:len(mp3FullFilename)-3], "wav")

	if err := os.MkdirAll(path.Dir(txtFullFilename), os.ModePerm); err != nil {
		return ci, err
	}

	if err := os.MkdirAll(path.Dir(wavFullFilename), os.ModePerm); err != nil {
		return ci, err
	}

	if err := os.MkdirAll(path.Dir(mp3FullFilename), os.ModePerm); err != nil {
		return ci, err
	}

	logrus.Debugf("running tts command with wavFilename=%s txtFilename=%s\n", wavFullFilename, txtFullFilename)
	ttsCmd := tm.execCommand("espeak-ng", "-v", tm.espeakVoice, "-s", fmt.Sprintf("%d", tm.wordsPerMinute), "-m", "-w", wavFullFilename, "-f", txtFullFilename)
	if err := ttsCmd.Run(); err != nil {
		return ci, err
	}
	ttsCmd.Stdout = os.Stdout
	ttsCmd.Stderr = os.Stderr
	ttsCmd.Wait()

	logrus.Debugf("running lame command with wavFilename=%s mp3Filename=%s\n", wavFullFilename, mp3FullFilename)
	lameCmd := tm.execCommand("lame", "-S", "-m", "m", wavFullFilename, mp3FullFilename)
	if err := lameCmd.Run(); err != nil {
		return ci, err
	}
	lameCmd.Stdout = os.Stdout
	lameCmd.Stderr = os.Stderr
	lameCmd.Wait()

	if err := os.Remove(wavFullFilename); err != nil {
		return ci, err
	}

	file, err := os.Open(mp3FullFilename)
	if err != nil {
		return ci, err
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		return ci, err
	}

	ci.Size = uint64(fileInfo.Size())

	return ci, nil
}

func (tm *Text2Mp3) SetConfig(key string, value string) bool {
	switch key {
	case "espeakvoice":
		tm.espeakVoice = value
		return true
	case "wordsperminute":
		v, err := strconv.Atoi(value)
		if err != nil {
			return false
		}

		tm.wordsPerMinute = v
		return true
	case "localpath":
		tm.localPath = value
		return true
	default:
		return false
	}
}

func (tm *Text2Mp3) GetConfig(key string) (bool, string) {
	switch key {
	case "espeakvoice":
		return true, tm.espeakVoice
	case "wordsperminute":
		return true, fmt.Sprintf("%d", tm.wordsPerMinute)
	case "localpath":
		return true, tm.localPath
	default:
		return false, ""
	}
}
