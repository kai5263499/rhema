package rhema

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/boltdb/bolt"
	"github.com/cayleygraph/cayley"
	"github.com/cayleygraph/cayley/graph"
	_ "github.com/cayleygraph/cayley/graph/kv/bolt"
	"github.com/cayleygraph/quad"
	"github.com/kai5263499/rhema/domain"
	pb "github.com/kai5263499/rhema/generated"
	"github.com/sirupsen/logrus"
	"google.golang.org/protobuf/proto"
)

var _ domain.Storage = (*storage)(nil)

// NewContentStorage returns an instance of ContentStorage
func NewContentStorage(
	cfg *domain.Config,
) (domain.Storage, error) {

	logrus.Debugf("initquadstore on %s", cfg.CayleyStoragePath)

	if err := graph.InitQuadStore("bolt", cfg.CayleyStoragePath, nil); err != nil {
		logrus.WithError(err).Error("error init quad store")
		if !strings.Contains(err.Error(), "exists") {
			return nil, err
		}
	}

	cayleyGraph, err := cayley.NewGraph("bolt", cfg.CayleyStoragePath, nil)
	if err != nil {
		return nil, err
	}

	boltdb, err := bolt.Open(cfg.BoltDBPath, 0600, nil)
	if err != nil {
		return nil, err
	}

	cs := &storage{
		cfg:         cfg,
		cayleyGraph: cayleyGraph,
		boltdb:      boltdb,
	}

	return cs, nil
}

// ContentStorage persists content artifacts to S3
type storage struct {
	cfg         *domain.Config
	cayleyGraph *cayley.Handle
	boltdb      *bolt.DB
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

	err = InsertRequestIntoCayley(ci, cs.cayleyGraph)
	if err != nil {
		return
	}

	err = InsertRequestIntoBolt(ci, cs.boltdb, cs.cfg.BoltDBBucket)
	if err != nil {
		return
	}

	logrus.Debugf("inserted requestHash=%s into both cayley and bolt", ci.RequestHash)

	return
}

func InsertRequestIntoBolt(request *pb.Request, boltdb *bolt.DB, boltBucket string) error {
	requestBytes, err := proto.Marshal(request)
	if err != nil {
		return err
	}

	return boltdb.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte(boltBucket))
		if err != nil {
			return err
		}
		if b == nil {
			return fmt.Errorf("Bucket not found")
		}

		// Set the value at the given key
		return b.Put([]byte(request.RequestHash), requestBytes)
	})
}

func InsertRequestIntoCayley(request *pb.Request, cayley *cayley.Handle) error {
	return cayley.AddQuad(quad.Make(request.RequestHash, "submittedBy", request.SubmittedBy, nil))
}

func (cs *storage) Close() (err error) {
	err = cs.boltdb.Close()
	err = cs.cayleyGraph.Close()
	return
}

func (cs *storage) Load(requestHash string) (req *pb.Request, err error) {
	return ReadRequestFromBolt(requestHash, cs.boltdb, cs.cfg.BoltDBBucket)
}

func (cs *storage) ListAll() (requests []*pb.Request) {
	var err error
	requests = make([]*pb.Request, 0)

	requestHashes, err := ReadAllRequestHashesFromCayley(cs.cayleyGraph)
	if err != nil {
		return
	}

	for _, requestHash := range requestHashes {
		req, err := ReadRequestFromBolt(requestHash, cs.boltdb, cs.cfg.BoltDBBucket)
		if err != nil {
			return
		}

		requests = append(requests, req)
	}

	return
}

func ReadAllRequestHashesFromCayley(cayleyGraph *cayley.Handle) ([]string, error) {
	// Create a slice to store the requestHashes
	requestHashes := []string{}

	// Use cayley.StartPath to query the graph for all nodes with the "Type" property
	it := cayley.StartPath(cayleyGraph).Has("submittedBy").Iterate(nil)

	// Iterate over the path and retrieve the requestHashes
	err := it.EachValue(nil, func(value quad.Value) {
		res := value.Native().(string)
		requestHashes = append(requestHashes, res)
	})
	if err != nil {
		return nil, fmt.Errorf("failed to iterate over path: %v", err)
	}

	return requestHashes, nil
}

func ReadRequestFromBolt(requestHash string, boltDb *bolt.DB, boltDbBucket string) (request *pb.Request, err error) {
	var requestBytes []byte

	// Start a read-only transaction
	err = boltDb.View(func(tx *bolt.Tx) error {
		// Retrieve the bucket named "MyBucket"
		b := tx.Bucket([]byte(boltDbBucket))
		if b == nil {
			return fmt.Errorf("Bucket not found")
		}

		// Retrieve the value at the given key
		requestBytes = b.Get([]byte(requestHash))
		if requestBytes == nil {
			return fmt.Errorf("Key not found")
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	request = &pb.Request{}
	err = proto.Unmarshal(requestBytes, request)

	return
}
