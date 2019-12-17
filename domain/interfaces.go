package domain

import (
	pb "github.com/kai5263499/rhema/generated"
)

type Converter interface {
	Convert(pb.Request) (pb.Request, error)
}

type Storage interface {
	Store(pb.Request) (pb.Request, error)
}

type Processor interface {
	Process(pb.Request) (pb.Request, error)
}
