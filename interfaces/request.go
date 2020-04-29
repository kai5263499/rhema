package interfaces

import "time"

type Request interface {
	Presign(expire time.Duration) (string, error)
}
