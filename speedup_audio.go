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

func NewSpeedupAudio(contentStorage *ContentStorage, localPath string) *SpeedupAudio {
	return &SpeedupAudio{
		contentStorage: contentStorage,
		localPath:      localPath,
		execCommand:    exec.Command,
	}
}

type SpeedupAudio struct {
	contentStorage domain.Storage
	localPath      string
	execCommand    func(command string, args ...string) *exec.Cmd
}

func (sa *SpeedupAudio) Convert(ci pb.Request) (pb.Request, error) {

	slowFilename, err := GetFilePath(ci)
	if err != nil {
		return ci, err
	}

	slowFullFilename := filepath.Join(sa.localPath, slowFilename)
	tmpFullFilename := fmt.Sprintf("%s%s", slowFullFilename[:len(slowFullFilename)-4], "-TMP.mp3")

	logrus.Debugf("before ffmpeg slowFullFilename=%s tmpFullFilename=%s", slowFullFilename, tmpFullFilename)

	ffmpegCmd := sa.execCommand("ffmpeg",
		"-y",
		"-i", slowFullFilename,
		"-filter:a", "atempo=2.0, volume=10dB",
		"-c:a", "libmp3lame", "-q:a", "4", tmpFullFilename)

	if err = ffmpegCmd.Run(); err != nil {
		return ci, err
	}
	ffmpegCmd.Wait()

	ci.Type = pb.Request_AUDIO
	mp3FileName, err := GetFilePath(ci)
	if err != nil {
		return ci, err
	}

	mp3FullFilename := filepath.Join(sa.localPath, mp3FileName)

	logrus.Debugf("before rename %s -> %s", tmpFullFilename, mp3FullFilename)

	err = os.Rename(tmpFullFilename, mp3FullFilename)
	if err != nil {
		return ci, err
	}

	sa.contentStorage.Store(ci)

	return ci, nil
}
