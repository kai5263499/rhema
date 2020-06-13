package rhema

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"

	"github.com/kai5263499/rhema/domain"
	pb "github.com/kai5263499/rhema/generated"
	"github.com/sirupsen/logrus"
)

var _ domain.Converter = (*SpeedupAudio)(nil)

func NewSpeedupAudio(localPath string, atempo float32) *SpeedupAudio {
	return &SpeedupAudio{
		localPath:   localPath,
		execCommand: exec.Command,
		atempo:      atempo,
	}
}

type SpeedupAudio struct {
	localPath   string
	execCommand func(command string, args ...string) *exec.Cmd
	atempo      float32
}

func (sa *SpeedupAudio) Convert(ci pb.Request) (pb.Request, error) {

	slowFilename, err := GetFilePath(ci)
	if err != nil {
		return ci, err
	}

	slowFullFilename := filepath.Join(sa.localPath, slowFilename)
	tmpFullFilename := fmt.Sprintf("%s%s", slowFullFilename[:len(slowFullFilename)-4], "-TMP.mp3")

	ffmpegCmd := sa.execCommand("ffmpeg",
		"-y",
		"-i", slowFullFilename,
		"-filter:a", fmt.Sprintf("atempo=%.1f, volume=10dB", sa.atempo),
		"-c:a", "libmp3lame", "-q:a", "4", tmpFullFilename)

	if err := ffmpegCmd.Run(); err != nil {
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

	if err := os.Rename(tmpFullFilename, mp3FullFilename); err != nil {
		return ci, err
	}

	return ci, nil
}

func (sa *SpeedupAudio) SetConfig(key string, value string) bool {
	switch key {
	case "atempo":
		f, err := strconv.ParseFloat(value, 32)
		if err != nil {
			return false
		}
		sa.atempo = float32(f)
		return true
	case "localpath":
		sa.localPath = value
		return true
	default:
		return false
	}
}

func (sa *SpeedupAudio) GetConfig(key string) (bool, string) {
	switch key {
	case "atempo":
		return true, fmt.Sprintf("%.1f", sa.atempo)
	case "localpath":
		return true, sa.localPath
	default:
		return false, ""
	}
}
