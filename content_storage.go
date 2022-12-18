package rhema

import (
	"io"
	"os"
	"path/filepath"

	"github.com/kai5263499/rhema/domain"
	pb "github.com/kai5263499/rhema/generated"
	"github.com/sirupsen/logrus"

	"github.com/gomodule/redigo/redis"
)

// NewContentStorage returns an instance of ContentStorage
func NewContentStorage(
	cfg *domain.Config,
	redisConn *redis.Conn,
) (*ContentStorage, error) {

	cs := &ContentStorage{
		cfg:       cfg,
		redisConn: redisConn,
	}

	return cs, nil
}

// ContentStorage persists content artifacts to S3
type ContentStorage struct {
	cfg       *domain.Config
	redisConn *redis.Conn
}

func copyFileContents(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer func() {
		cerr := out.Close()
		if err == nil {
			err = cerr
		}
	}()
	if _, err = io.Copy(out, in); err != nil {
		return err
	}
	err = out.Sync()
	return nil
}

// Store persists a content item in S3
func (cs *ContentStorage) Store(ci *pb.Request) error {
	itemPath, err := getPath(ci)
	if err != nil {
		logrus.WithError(err).Error("unable to get path")
		return err
	}

	ci.StoragePath = itemPath

	logrus.Debugf("storing content item to %s", ci.StoragePath)

	if ci.Type == pb.ContentType_AUDIO {
		src := filepath.Join(cs.cfg.LocalPath, itemPath)
		dst := filepath.Join(cs.cfg.LocalPath, filepath.Base(itemPath))

		if err := os.MkdirAll(filepath.Dir(dst), os.ModePerm); err != nil {
			logrus.WithError(err).Errorf("unable to MkdirAll on %s", dst)
			return err
		}

		if err := copyFileContents(src, dst); err != nil {
			logrus.WithError(err).Errorf("unable to copy %s to %s", src, dst)
			return err
		}

		if err := os.Chown(dst, cs.cfg.ChownTo, cs.cfg.ChownTo); err != nil {
			logrus.WithError(err).Errorf("unable to chown %s to %d", dst, cs.cfg.ChownTo)
			return err
		}

		logrus.Debugf("%s -> %s", src, dst)
	}

	// TODO add entries to storage

	return nil
}
