package rhema

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/kai5263499/rhema/domain"
	pb "github.com/kai5263499/rhema/generated"
	"github.com/sirupsen/logrus"
)

var _ domain.Converter = (*SpeedupAudio)(nil)

func NewSpeedupAudio(cfg *domain.Config, exec func(command string, args ...string) *exec.Cmd) *SpeedupAudio {
	return &SpeedupAudio{
		cfg:         cfg,
		execCommand: exec,
	}
}

type SpeedupAudio struct {
	cfg         *domain.Config
	execCommand func(command string, args ...string) *exec.Cmd
}

func (sa *SpeedupAudio) Convert(ci *pb.Request) error {

	slowFilename, err := GetFilePath(ci)
	if err != nil {
		return err
	}

	slowFullFilename := filepath.Join(sa.cfg.TmpPath, slowFilename)
	tmpFullFilename := fmt.Sprintf("%s%s", slowFullFilename[:len(slowFullFilename)-4], "-TMP.mp3")

	ffmpegCmd := sa.execCommand("ffmpeg",
		"-y",
		"-i", slowFullFilename,
		"-filter:a", fmt.Sprintf("atempo=%s", ci.ATempo),
		"-c:a", "libmp3lame", "-q:a", "4", tmpFullFilename)

	ffmpegCmd.Stderr = os.Stdout
	ffmpegCmd.Stdout = os.Stdout

	logrus.Debugf("running ffmpeg command with ffmpegCmd=%s", ffmpegCmd)
	if err := ffmpegCmd.Run(); err != nil {
		return err
	}
	ffmpegCmd.Wait()

	ci.Type = pb.ContentType_AUDIO
	mp3FileName, err := GetFilePath(ci)
	if err != nil {
		return err
	}

	mp3FullFilename := filepath.Join(sa.cfg.TmpPath, mp3FileName)

	if err := os.Rename(tmpFullFilename, mp3FullFilename); err != nil {
		return err
	}

	logrus.Debugf("renamed %s -> %s", tmpFullFilename, mp3FullFilename)

	return nil
}
