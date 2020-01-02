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

var _ domain.Converter = (*YouTube)(nil)

func NewYoutube(scrape domain.Converter, contentStorage domain.Storage, speedupAudio domain.Converter, localPath string) *YouTube {
	return &YouTube{
		scrape:         scrape,
		contentStorage: contentStorage,
		speedupAudio:   speedupAudio,
		localPath:      localPath,
		execCommand:    exec.Command,
	}
}

type YouTube struct {
	scrape         domain.Converter
	contentStorage domain.Storage
	speedupAudio   domain.Converter
	localPath      string
	execCommand    func(command string, args ...string) *exec.Cmd
}

func (yt *YouTube) Convert(ci pb.Request) (pb.Request, error) {

	var err error

	scrapeReq, err := yt.scrape.Convert(ci)
	if err != nil {
		return ci, err
	}

	if len(scrapeReq.Title) > 3 {
		ci.Title = scrapeReq.Title
	}

	ci.Type = pb.Request_AUDIO
	fileName, err := GetFilePath(ci)
	if err != nil {
		return ci, err
	}

	mp3FullFilename := filepath.Join(yt.localPath, fileName)

	mp3FullFilename = fmt.Sprintf("%s%s", mp3FullFilename[:len(mp3FullFilename)-4], "")

	if err = os.MkdirAll(path.Dir(mp3FullFilename), os.ModePerm); err != nil {
		return ci, err
	}

	logrus.Debugf("before execCommand with mp3FullFilename=%s uri=%s", mp3FullFilename, ci.Uri)

	youtubeCmd := yt.execCommand("youtube-dl",
		"--extract-audio",
		"--add-metadata",
		"--audio-format", "mp3",
		"--restrict-filenames",
		"-o", fmt.Sprintf("%s.%%(ext)s", mp3FullFilename),
		ci.Uri)

	youtubeCmd.Stdout = os.Stdout
	youtubeCmd.Stderr = os.Stderr

	if err = youtubeCmd.Run(); err != nil {
		return ci, err
	}
	youtubeCmd.Wait()

	return ci, nil
}

func (yt *YouTube) SetConfig(key string, value string) bool {
	switch key {
	case "localpath":
		yt.localPath = value
		return true
	default:
		return false
	}
}

func (yt *YouTube) GetConfig(key string) (bool, string) {
	switch key {
	case "localpath":
		return true, yt.localPath
	default:
		return false, ""
	}
}
