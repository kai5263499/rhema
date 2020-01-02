package rhema

import pb "github.com/kai5263499/rhema/generated"

type fakeContentStore struct{}

func (f *fakeContentStore) Store(ci pb.Request) (pb.Request, error) { return ci, nil }
func (f *fakeContentStore) GetConfig(k string) (bool, string)       { return false, "" }
func (f *fakeContentStore) SetConfig(k string, v string) bool       { return false }

type audioConverter struct{}

func (f *audioConverter) Convert(ci pb.Request) (pb.Request, error) {
	ci.Type = pb.Request_AUDIO
	return ci, nil
}
func (c *audioConverter) GetConfig(k string) (bool, string) { return false, "" }
func (c *audioConverter) SetConfig(k string, v string) bool { return false }

type textConverter struct{}

func (f *textConverter) Convert(ci pb.Request) (pb.Request, error) {
	ci.Type = pb.Request_TEXT
	return ci, nil
}
func (f *textConverter) GetConfig(k string) (bool, string) { return false, "" }
func (f *textConverter) SetConfig(k string, v string) bool { return false }

type youtubeConverter struct{}

func (f *youtubeConverter) Convert(ci pb.Request) (pb.Request, error) {
	ci.Type = pb.Request_YOUTUBE
	return ci, nil
}
func (f *youtubeConverter) GetConfig(k string) (bool, string) { return false, "" }
func (f *youtubeConverter) SetConfig(k string, v string) bool { return false }

type videoConverter struct{}

func (f *videoConverter) Convert(ci pb.Request) (pb.Request, error) {
	ci.Type = pb.Request_VIDEO
	return ci, nil
}
func (f *videoConverter) GetConfig(k string) (bool, string) { return false, "" }
func (f *videoConverter) SetConfig(k string, v string) bool { return false }
