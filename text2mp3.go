package rhema

import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"

	"github.com/kai5263499/rhema/domain"
	pb "github.com/kai5263499/rhema/generated"
	"github.com/sirupsen/logrus"
)

var _ domain.Converter = (*Text2Mp3)(nil)

func NewText2Mp3(cfg *domain.Config) *Text2Mp3 {
	return &Text2Mp3{
		cfg:         cfg,
		execCommand: exec.Command,
	}
}

type Text2Mp3 struct {
	cfg         *domain.Config
	execCommand func(command string, args ...string) *exec.Cmd
}

func (tm *Text2Mp3) Convert(ci *pb.Request) (err error) {

	logrus.Debugf("converting text2mp3\n")

	if ci.Type != pb.ContentType_TEXT {
		err = fmt.Errorf("Invalid request type %s", ci.Type.String())
		return
	}

	if ci.Length < 3 {
		err = fmt.Errorf("Text length too short %#+v", ci)
		return
	}

	txtFilename, err := GetFilePath(ci)
	if err != nil {
		return
	}

	txtFullFilename := filepath.Join(tm.cfg.TmpPath, txtFilename)

	ci.Type = pb.ContentType_AUDIO

	mp3Filename, err := GetFilePath(ci)
	if err != nil {
		return
	}

	mp3FullFilename := filepath.Join(tm.cfg.TmpPath, mp3Filename)
	wavFullFilename := fmt.Sprintf("%s%s", mp3FullFilename[:len(mp3FullFilename)-3], "wav")

	if err = os.MkdirAll(path.Dir(txtFullFilename), os.ModePerm); err != nil {
		return
	}

	if err = os.MkdirAll(path.Dir(wavFullFilename), os.ModePerm); err != nil {
		return
	}

	if err = os.MkdirAll(path.Dir(mp3FullFilename), os.ModePerm); err != nil {
		return
	}

	if ci.WordsPerMinute == 0 {
		ci.WordsPerMinute = uint32(tm.cfg.WordsPerMinute)
	}

	if len(ci.ESpeakVoice) == 0 {
		ci.ESpeakVoice = tm.cfg.EspeakVoice
	}

	ttsCmd := tm.execCommand("espeak-ng", "-v", ci.ESpeakVoice, "-s", fmt.Sprintf("%d", ci.WordsPerMinute), "-m", "-w", wavFullFilename, "-f", txtFullFilename)
	logrus.Debugf("running tts command %s", ttsCmd)
	if err = ttsCmd.Run(); err != nil {
		return
	}
	ttsCmd.Stdout = os.Stdout
	ttsCmd.Stderr = os.Stdout
	ttsCmd.Wait()

	lameCmd := tm.execCommand("lame", "-S", "-m", "m", wavFullFilename, mp3FullFilename)
	logrus.Debugf("running lame command %s", lameCmd)
	if err = lameCmd.Run(); err != nil {
		return
	}
	lameCmd.Stdout = os.Stdout
	lameCmd.Stderr = os.Stdout
	lameCmd.Wait()

	if err = os.Remove(wavFullFilename); err != nil {
		return
	}

	file, err := os.Open(mp3FullFilename)
	if err != nil {
		return
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		return
	}

	ci.Size = uint64(fileInfo.Size())

	return
}
