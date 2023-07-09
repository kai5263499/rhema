package rhema

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/kai5263499/rhema/domain"
	pb "github.com/kai5263499/rhema/generated"
	"github.com/sirupsen/logrus"
)

var _ domain.Storage = (*storage)(nil)

// NewContentStorage returns an instance of ContentStorage
func NewContentStorage(
	cfg *domain.Config,
) (domain.Storage, error) {

	cs := &storage{
		cfg: cfg,
	}

	return cs, nil
}

// ContentStorage persists content artifacts to S3
type storage struct {
	cfg *domain.Config
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
func (cs *storage) Store(ci *pb.Request) (err error) {
	itemPath, err := getPath(ci)
	if err != nil {
		logrus.WithError(err).Error("unable to get path")
		return
	}

	ci.StoragePath = itemPath

	logrus.Debugf("storing content item to %s", ci.StoragePath)

	if ci.Type == pb.ContentType_AUDIO {
		src := filepath.Join(cs.cfg.TmpPath, itemPath)
		dst := filepath.Join(cs.cfg.LocalPath, filepath.Base(itemPath))

		logrus.Debugf("mkdirall on %s", dst)
		if err = os.MkdirAll(filepath.Dir(dst), os.ModePerm); err != nil {
			logrus.WithError(err).Errorf("unable to MkdirAll on %s", dst)
			return
		}

		if err = copyFileContents(src, dst); err != nil {
			logrus.WithError(err).Errorf("unable to copy %s to %s", src, dst)
			return
		}

		if err = os.Chown(dst, cs.cfg.ChownTo, cs.cfg.ChownTo); err != nil {
			logrus.WithError(err).Errorf("unable to chown %s to %d", dst, cs.cfg.ChownTo)
			return
		}

		logrus.Debugf("%s -> %s", src, dst)
	}

	// TODO store content into text file on disk

	logrus.Debugf("inserted requestHash=%s into both cayley and bolt", ci.RequestHash)

	return
}

func (cs *storage) Close() (err error) {
	return
}

func (cs *storage) Load(requestHash string) (req *pb.Request, err error) {
	return nil, fmt.Errorf("not implemented")
}

func (cs *storage) ListAll() (requests []*pb.Request) {
	requests = make([]*pb.Request, 0)
	return
}
