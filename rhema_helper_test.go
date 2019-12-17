package rhema

import pb "github.com/kai5263499/rhema/generated"

type fakeContentStore struct{}

func (fs *fakeContentStore) Store(ci pb.Request) (pb.Request, error) { return ci, nil }

type audioConverter struct{}

func (sc *audioConverter) Convert(ci pb.Request) (pb.Request, error) {
	ci.Type = pb.Request_AUDIO
	return ci, nil
}

type textConverter struct{}

func (sc *textConverter) Convert(ci pb.Request) (pb.Request, error) {
	ci.Type = pb.Request_TEXT
	return ci, nil
}

type youtubeConverter struct{}

func (sc *youtubeConverter) Convert(ci pb.Request) (pb.Request, error) {
	ci.Type = pb.Request_YOUTUBE
	return ci, nil
}

type videoConverter struct{}

func (sc *videoConverter) Convert(ci pb.Request) (pb.Request, error) {
	ci.Type = pb.Request_VIDEO
	return ci, nil
}
