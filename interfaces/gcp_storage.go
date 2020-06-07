package interfaces

import "cloud.google.com/go/storage"

type GCPStorage interface {
	Bucket(name string) *storage.BucketHandle
}
