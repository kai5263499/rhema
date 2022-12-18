package rhema

import pb "github.com/kai5263499/rhema/generated"

type fakeContentStore struct{}

func (f *fakeContentStore) Store(ci *pb.Request) error { return nil }

type audioConverter struct{}

func (f *audioConverter) Convert(ci *pb.Request) error {
	ci.Type = pb.ContentType_AUDIO
	return nil
}

type textConverter struct{}

func (f *textConverter) Convert(ci *pb.Request) error {
	ci.Type = pb.ContentType_TEXT
	return nil
}

type youtubeConverter struct{}

func (f *youtubeConverter) Convert(ci *pb.Request) error {
	ci.Type = pb.ContentType_YOUTUBE
	return nil
}

type videoConverter struct{}

func (f *videoConverter) Convert(ci *pb.Request) error {
	ci.Type = pb.ContentType_VIDEO
	return nil
}
