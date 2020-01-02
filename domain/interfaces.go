package domain

import (
	pb "github.com/kai5263499/rhema/generated"
)

type Converter interface {
	SetConfig(string, string) bool
	GetConfig(string) (bool, string)
	Convert(pb.Request) (pb.Request, error)
}

type Storage interface {
	SetConfig(string, string) bool
	GetConfig(string) (bool, string)
	Store(pb.Request) (pb.Request, error)
}

type Processor interface {
	SetConfig(string, string) bool
	GetConfig(string) (bool, string)
	Process(pb.Request) (pb.Request, error)
}
