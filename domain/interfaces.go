package domain

import (
	pb "github.com/kai5263499/rhema/generated"
)

type Converter interface {
	Convert(*pb.Request) error
}

type Storage interface {
	Store(*pb.Request) error
	Load(requestHash string) (*pb.Request, error)
	ListAll() []*pb.Request
	Close() error
}

type Processor interface {
	Process(*pb.Request) error
}

type Comms interface {
	RequestChan() chan pb.Request
	SendRequest(req *pb.Request) error
}
