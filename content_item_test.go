package rhema

import (
	pb "github.com/kai5263499/rhema/generated"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

type titleTest struct {
	title string
	want  string
}

type extTypeTest struct {
	extType pb.Request_ContentType
	want    string
}

type uriParseTest struct {
	uri  string
	want string
}

type pathTest struct {
	req  pb.Request
	want string
}

var _ = Describe("content_item", func() {
	It("Should return the right title", func() {
		tests := []titleTest{
			{title: "test title", want: "test_title"},
		}

		for _, tc := range tests {
			Expect(cleanTitle(tc.title)).To(Equal(tc.want))
		}

	})
	It("Should return the right extension from a given type", func() {
		tests := []extTypeTest{
			{extType: pb.Request_TEXT, want: "txt"},
			{extType: pb.Request_AUDIO, want: "mp3"},
		}

		for _, tc := range tests {
			fileExt, err := getExtFromType(tc.extType)
			Expect(err).To(BeNil())
			Expect(fileExt).To(Equal(tc.want))
		}
	})
	It("Should return the right extension from a given uri", func() {
		tests := []uriParseTest{
			{uri: "http://somedomain.com/somefile.txt", want: "txt"},
			{uri: "http://anotherdomain.com/hot_new_lecture.mp3", want: "mp3"},
			{uri: "https://ia601407.us.archive.org/32/items/videoplayback_20191122_2202/videoplayback.mp4", want: "mp4"},
		}

		for _, tc := range tests {
			fileExt, err := getExtFromUri(tc.uri)
			Expect(err).To(BeNil())
			Expect(fileExt).To(Equal(tc.want))
		}
	})
	It("Should return the right title from a given uri", func() {
		tests := []uriParseTest{
			{uri: "http://somedomain.com/somefile.txt", want: "somefile"},
			{uri: "http://anotherdomain.com/hot_new_lecture.mp3", want: "hot_new_lecture"},
			{uri: "https://ia601407.us.archive.org/32/items/videoplayback_20191122_2202/videoplayback.mp4", want: "videoplayback"},
		}

		for _, tc := range tests {
			fileExt, err := parseTitleFromUri(tc.uri)
			Expect(err).To(BeNil())
			Expect(fileExt).To(Equal(tc.want))
		}
	})
	It("Should return the right local path", func() {
		tests := []pathTest{
			{req: pb.Request{Type: pb.Request_TEXT, Created: 383576400, Title: "test_title"}, want: "TEXT/1982/02/26/test_title.txt"},
			{req: pb.Request{Type: pb.Request_AUDIO, Created: 383576400, Title: "test_title"}, want: "AUDIO/1982/02/26/test_title.mp3"},
		}

		for _, tc := range tests {
			filePath, err := GetFilePath(tc.req)
			Expect(err).To(BeNil())
			Expect(filePath).To(Equal(tc.want))
		}
	})
	It("Should return the right s3 path", func() {
		tests := []pathTest{
			{req: pb.Request{Type: pb.Request_TEXT, Created: 383576400, Title: "test_title"}, want: "TEXT/1982/02/26/test_title.txt"},
			{req: pb.Request{Type: pb.Request_AUDIO, Created: 383576400, Title: "test_title"}, want: "AUDIO/1982/02/26/test_title.mp3"},
		}

		for _, tc := range tests {
			s3path, err := getPath(tc.req)
			Expect(err).To(BeNil())
			Expect(s3path).To(Equal(tc.want))
		}
	})
})
