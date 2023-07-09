package rhema

import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"

	"github.com/icza/gox/stringsx"
	"github.com/kai5263499/rhema/domain"
	pb "github.com/kai5263499/rhema/generated"
	"github.com/sirupsen/logrus"
)

var _ domain.Converter = (*YouTube)(nil)

func NewYoutube(cfg *domain.Config, scrape domain.Converter, speedupAudio domain.Converter) *YouTube {
	return &YouTube{
		cfg:          cfg,
		scrape:       scrape,
		speedupAudio: speedupAudio,
		execCommand:  exec.Command,
	}
}

type YouTube struct {
	cfg          *domain.Config
	scrape       domain.Converter
	speedupAudio domain.Converter
	execCommand  func(command string, args ...string) *exec.Cmd
}

func (yt *YouTube) Convert(ci *pb.Request) error {

	if err := yt.scrape.Convert(ci); err != nil {
		logrus.WithFields(logrus.Fields{
			"uri":   ci.Uri,
			"title": ci.Title,
			"err":   err,
		}).Warnf("unable to scrape title")
	} else {
		if len(ci.Title) > 3 {
			ci.Title = stringsx.Clean(ci.Title)
		}
	}

	ci.Type = pb.ContentType_AUDIO
	fileName, err := GetFilePath(ci)
	if err != nil {
		return err
	}

	mp3FullFilename := filepath.Join(yt.cfg.TmpPath, fileName)

	mp3FullFilename = fmt.Sprintf("%s%s", mp3FullFilename[:len(mp3FullFilename)-4], "")

	if err := os.MkdirAll(path.Dir(mp3FullFilename), os.ModePerm); err != nil {
		return err
	}

	logrus.Debugf("before execCommand with mp3FullFilename=%s uri=%s", mp3FullFilename, ci.Uri)

	youtubeCmd := yt.execCommand("yt-dlp",
		"--extract-audio",
		"--add-metadata",
		"--audio-format", "mp3",
		"--restrict-filenames",
		"-o", fmt.Sprintf("%s.%%(ext)s", mp3FullFilename),
		ci.Uri)

	youtubeCmd.Stdout = os.Stdout
	youtubeCmd.Stderr = os.Stdout

	if err := youtubeCmd.Run(); err != nil {
		return err
	}
	youtubeCmd.Wait()

	return nil
}
