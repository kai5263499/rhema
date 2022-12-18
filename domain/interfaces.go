package domain

import (
	pb "github.com/kai5263499/rhema/generated"
)

type Converter interface {
	Convert(*pb.Request) error
}

type Storage interface {
	Store(*pb.Request) error
}

type Processor interface {
	Process(*pb.Request) error
}

type Comms interface {
	RequestChan() chan pb.Request
	SendRequest(req *pb.Request) error
}
