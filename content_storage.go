package rhema

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/boltdb/bolt"
	"github.com/kai5263499/rhema/domain"
	pb "github.com/kai5263499/rhema/generated"
	"github.com/sirupsen/logrus"
	"google.golang.org/protobuf/proto"

	"github.com/gomodule/redigo/redis"
)

var _ domain.Storage = (*storage)(nil)

// NewContentStorage returns an instance of ContentStorage
func NewContentStorage(
	cfg *domain.Config,
	redisConn *redis.Conn,
	boltdb *bolt.DB,
) (domain.Storage, error) {

	cs := &storage{
		cfg:       cfg,
		redisConn: redisConn,
		boltdb:    boltdb,
	}

	return cs, nil
}

// ContentStorage persists content artifacts to S3
type storage struct {
	cfg       *domain.Config
	redisConn *redis.Conn
	boltdb    *bolt.DB
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
		src := filepath.Join(cs.cfg.LocalPath, itemPath)
		dst := filepath.Join(cs.cfg.LocalPath, filepath.Base(itemPath))

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

	// serialize proto
	protoBytes, err := proto.Marshal(ci)
	if err != nil {
		logrus.WithError(err).Errorf("error marshalling proto to bytes")
		return
	}

	// TODO add entries to storage
	err = cs.boltdb.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte(cs.cfg.BoltDBBucket))
		if err != nil {
			return err
		}

		if err = b.Put([]byte(ci.RequestHash), protoBytes); err != nil {
			return err
		}

		return nil
	})

	return
}

func (cs *storage) Close() error {
	return cs.boltdb.Close()
}

func (cs *storage) Load(requestHash string) (req *pb.Request, err error) {
	var protoBytes []byte
	req = &pb.Request{}

	err = cs.boltdb.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(cs.cfg.BoltDBBucket))
		if b == nil {
			return fmt.Errorf("Bucket not found")
		}

		// Retrieve the value at key "greeting"
		protoBytes = b.Get([]byte(requestHash))

		return nil
	})
	if err != nil {
		return
	}

	if err = proto.Unmarshal(protoBytes, req); err != nil {
		return
	}

	return
}

func (cs *storage) ListAll() []*pb.Request {
	requests := make([]*pb.Request, 0)

	if err := cs.boltdb.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(cs.cfg.BoltDBBucket))
		if b == nil {
			return fmt.Errorf("Bucket not found")
		}

		// Iterate over all keys in the bucket
		err := b.ForEach(func(k, v []byte) error {
			req := &pb.Request{}
			if err := proto.Unmarshal(v, req); err != nil {
				return err
			}

			requests = append(requests, req)

			return nil
		})
		if err != nil {
			return err
		}

		return nil
	}); err != nil {
		logrus.WithError(err).Errorf("error listing keys in boltdb")
	}

	return requests
}
