// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.28.1
// 	protoc        v3.17.3
// source: rhema.proto

package generated

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type ContentType int32

const (
	ContentType_URI          ContentType = 0
	ContentType_YOUTUBE      ContentType = 1
	ContentType_TEXT         ContentType = 2
	ContentType_AUDIO        ContentType = 3
	ContentType_VIDEO        ContentType = 4
	ContentType_PDF          ContentType = 5
	ContentType_YOUTUBE_LIST ContentType = 6
)

// Enum value maps for ContentType.
var (
	ContentType_name = map[int32]string{
		0: "URI",
		1: "YOUTUBE",
		2: "TEXT",
		3: "AUDIO",
		4: "VIDEO",
		5: "PDF",
		6: "YOUTUBE_LIST",
	}
	ContentType_value = map[string]int32{
		"URI":          0,
		"YOUTUBE":      1,
		"TEXT":         2,
		"AUDIO":        3,
		"VIDEO":        4,
		"PDF":          5,
		"YOUTUBE_LIST": 6,
	}
)

func (x ContentType) Enum() *ContentType {
	p := new(ContentType)
	*p = x
	return p
}

func (x ContentType) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (ContentType) Descriptor() protoreflect.EnumDescriptor {
	return file_rhema_proto_enumTypes[0].Descriptor()
}

func (ContentType) Type() protoreflect.EnumType {
	return &file_rhema_proto_enumTypes[0]
}

func (x ContentType) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use ContentType.Descriptor instead.
func (ContentType) EnumDescriptor() ([]byte, []int) {
	return file_rhema_proto_rawDescGZIP(), []int{0}
}

type ResponseCode int32

const (
	ResponseCode_ERROR    ResponseCode = 0
	ResponseCode_ACCEPTED ResponseCode = 1
)

// Enum value maps for ResponseCode.
var (
	ResponseCode_name = map[int32]string{
		0: "ERROR",
		1: "ACCEPTED",
	}
	ResponseCode_value = map[string]int32{
		"ERROR":    0,
		"ACCEPTED": 1,
	}
)

func (x ResponseCode) Enum() *ResponseCode {
	p := new(ResponseCode)
	*p = x
	return p
}

func (x ResponseCode) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (ResponseCode) Descriptor() protoreflect.EnumDescriptor {
	return file_rhema_proto_enumTypes[1].Descriptor()
}

func (ResponseCode) Type() protoreflect.EnumType {
	return &file_rhema_proto_enumTypes[1]
}

func (x ResponseCode) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use ResponseCode.Descriptor instead.
func (ResponseCode) EnumDescriptor() ([]byte, []int) {
	return file_rhema_proto_rawDescGZIP(), []int{1}
}

type BotAction int32

const (
	BotAction_SET_CONFIG    BotAction = 0
	BotAction_GET_CONFIG    BotAction = 1
	BotAction_SHOW_SETTINGS BotAction = 2
	BotAction_HELP          BotAction = 4
)

// Enum value maps for BotAction.
var (
	BotAction_name = map[int32]string{
		0: "SET_CONFIG",
		1: "GET_CONFIG",
		2: "SHOW_SETTINGS",
		4: "HELP",
	}
	BotAction_value = map[string]int32{
		"SET_CONFIG":    0,
		"GET_CONFIG":    1,
		"SHOW_SETTINGS": 2,
		"HELP":          4,
	}
)

func (x BotAction) Enum() *BotAction {
	p := new(BotAction)
	*p = x
	return p
}

func (x BotAction) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (BotAction) Descriptor() protoreflect.EnumDescriptor {
	return file_rhema_proto_enumTypes[2].Descriptor()
}

func (BotAction) Type() protoreflect.EnumType {
	return &file_rhema_proto_enumTypes[2]
}

func (x BotAction) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use BotAction.Descriptor instead.
func (BotAction) EnumDescriptor() ([]byte, []int) {
	return file_rhema_proto_rawDescGZIP(), []int{2}
}

type Request struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Type                ContentType `protobuf:"varint,1,opt,name=Type,proto3,enum=ContentType" json:"Type,omitempty"`
	RequestHash         string      `protobuf:"bytes,2,opt,name=RequestHash,proto3" json:"RequestHash,omitempty"`
	Text                string      `protobuf:"bytes,3,opt,name=Text,proto3" json:"Text,omitempty"`
	Title               string      `protobuf:"bytes,4,opt,name=Title,proto3" json:"Title,omitempty"`
	Created             uint64      `protobuf:"varint,5,opt,name=Created,proto3" json:"Created,omitempty"`
	Size                uint64      `protobuf:"varint,6,opt,name=Size,proto3" json:"Size,omitempty"`
	Length              uint64      `protobuf:"varint,7,opt,name=Length,proto3" json:"Length,omitempty"`
	Uri                 string      `protobuf:"bytes,8,opt,name=Uri,proto3" json:"Uri,omitempty"`
	SubmittedBy         string      `protobuf:"bytes,9,opt,name=SubmittedBy,proto3" json:"SubmittedBy,omitempty"`
	SubmittedAt         uint64      `protobuf:"varint,10,opt,name=SubmittedAt,proto3" json:"SubmittedAt,omitempty"`
	NumberOfConversions uint32      `protobuf:"varint,11,opt,name=NumberOfConversions,proto3" json:"NumberOfConversions,omitempty"`
	DownloadURI         string      `protobuf:"bytes,12,opt,name=DownloadURI,proto3" json:"DownloadURI,omitempty"`
	WordsPerMinute      uint32      `protobuf:"varint,13,opt,name=WordsPerMinute,proto3" json:"WordsPerMinute,omitempty"`
	ESpeakVoice         string      `protobuf:"bytes,14,opt,name=ESpeakVoice,proto3" json:"ESpeakVoice,omitempty"`
	ATempo              string      `protobuf:"bytes,15,opt,name=ATempo,proto3" json:"ATempo,omitempty"`
	SubmittedWith       string      `protobuf:"bytes,16,opt,name=SubmittedWith,proto3" json:"SubmittedWith,omitempty"`
	StoragePath         string      `protobuf:"bytes,17,opt,name=StoragePath,proto3" json:"StoragePath,omitempty"`
	Tags                []string    `protobuf:"bytes,18,rep,name=Tags,proto3" json:"Tags,omitempty"`
}

func (x *Request) Reset() {
	*x = Request{}
	if protoimpl.UnsafeEnabled {
		mi := &file_rhema_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Request) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Request) ProtoMessage() {}

func (x *Request) ProtoReflect() protoreflect.Message {
	mi := &file_rhema_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Request.ProtoReflect.Descriptor instead.
func (*Request) Descriptor() ([]byte, []int) {
	return file_rhema_proto_rawDescGZIP(), []int{0}
}

func (x *Request) GetType() ContentType {
	if x != nil {
		return x.Type
	}
	return ContentType_URI
}

func (x *Request) GetRequestHash() string {
	if x != nil {
		return x.RequestHash
	}
	return ""
}

func (x *Request) GetText() string {
	if x != nil {
		return x.Text
	}
	return ""
}

func (x *Request) GetTitle() string {
	if x != nil {
		return x.Title
	}
	return ""
}

func (x *Request) GetCreated() uint64 {
	if x != nil {
		return x.Created
	}
	return 0
}

func (x *Request) GetSize() uint64 {
	if x != nil {
		return x.Size
	}
	return 0
}

func (x *Request) GetLength() uint64 {
	if x != nil {
		return x.Length
	}
	return 0
}

func (x *Request) GetUri() string {
	if x != nil {
		return x.Uri
	}
	return ""
}

func (x *Request) GetSubmittedBy() string {
	if x != nil {
		return x.SubmittedBy
	}
	return ""
}

func (x *Request) GetSubmittedAt() uint64 {
	if x != nil {
		return x.SubmittedAt
	}
	return 0
}

func (x *Request) GetNumberOfConversions() uint32 {
	if x != nil {
		return x.NumberOfConversions
	}
	return 0
}

func (x *Request) GetDownloadURI() string {
	if x != nil {
		return x.DownloadURI
	}
	return ""
}

func (x *Request) GetWordsPerMinute() uint32 {
	if x != nil {
		return x.WordsPerMinute
	}
	return 0
}

func (x *Request) GetESpeakVoice() string {
	if x != nil {
		return x.ESpeakVoice
	}
	return ""
}

func (x *Request) GetATempo() string {
	if x != nil {
		return x.ATempo
	}
	return ""
}

func (x *Request) GetSubmittedWith() string {
	if x != nil {
		return x.SubmittedWith
	}
	return ""
}

func (x *Request) GetStoragePath() string {
	if x != nil {
		return x.StoragePath
	}
	return ""
}

func (x *Request) GetTags() []string {
	if x != nil {
		return x.Tags
	}
	return nil
}

type Response struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	RequestHash string       `protobuf:"bytes,1,opt,name=RequestHash,proto3" json:"RequestHash,omitempty"`
	Code        ResponseCode `protobuf:"varint,2,opt,name=Code,proto3,enum=ResponseCode" json:"Code,omitempty"`
}

func (x *Response) Reset() {
	*x = Response{}
	if protoimpl.UnsafeEnabled {
		mi := &file_rhema_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Response) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Response) ProtoMessage() {}

func (x *Response) ProtoReflect() protoreflect.Message {
	mi := &file_rhema_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Response.ProtoReflect.Descriptor instead.
func (*Response) Descriptor() ([]byte, []int) {
	return file_rhema_proto_rawDescGZIP(), []int{1}
}

func (x *Response) GetRequestHash() string {
	if x != nil {
		return x.RequestHash
	}
	return ""
}

func (x *Response) GetCode() ResponseCode {
	if x != nil {
		return x.Code
	}
	return ResponseCode_ERROR
}

var File_rhema_proto protoreflect.FileDescriptor

var file_rhema_proto_rawDesc = []byte{
	0x0a, 0x0b, 0x72, 0x68, 0x65, 0x6d, 0x61, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0xa5, 0x04,
	0x0a, 0x07, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x20, 0x0a, 0x04, 0x54, 0x79, 0x70,
	0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x0c, 0x2e, 0x43, 0x6f, 0x6e, 0x74, 0x65, 0x6e,
	0x74, 0x54, 0x79, 0x70, 0x65, 0x52, 0x04, 0x54, 0x79, 0x70, 0x65, 0x12, 0x20, 0x0a, 0x0b, 0x52,
	0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x48, 0x61, 0x73, 0x68, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x0b, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x48, 0x61, 0x73, 0x68, 0x12, 0x12, 0x0a,
	0x04, 0x54, 0x65, 0x78, 0x74, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x54, 0x65, 0x78,
	0x74, 0x12, 0x14, 0x0a, 0x05, 0x54, 0x69, 0x74, 0x6c, 0x65, 0x18, 0x04, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x05, 0x54, 0x69, 0x74, 0x6c, 0x65, 0x12, 0x18, 0x0a, 0x07, 0x43, 0x72, 0x65, 0x61, 0x74,
	0x65, 0x64, 0x18, 0x05, 0x20, 0x01, 0x28, 0x04, 0x52, 0x07, 0x43, 0x72, 0x65, 0x61, 0x74, 0x65,
	0x64, 0x12, 0x12, 0x0a, 0x04, 0x53, 0x69, 0x7a, 0x65, 0x18, 0x06, 0x20, 0x01, 0x28, 0x04, 0x52,
	0x04, 0x53, 0x69, 0x7a, 0x65, 0x12, 0x16, 0x0a, 0x06, 0x4c, 0x65, 0x6e, 0x67, 0x74, 0x68, 0x18,
	0x07, 0x20, 0x01, 0x28, 0x04, 0x52, 0x06, 0x4c, 0x65, 0x6e, 0x67, 0x74, 0x68, 0x12, 0x10, 0x0a,
	0x03, 0x55, 0x72, 0x69, 0x18, 0x08, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x55, 0x72, 0x69, 0x12,
	0x20, 0x0a, 0x0b, 0x53, 0x75, 0x62, 0x6d, 0x69, 0x74, 0x74, 0x65, 0x64, 0x42, 0x79, 0x18, 0x09,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x0b, 0x53, 0x75, 0x62, 0x6d, 0x69, 0x74, 0x74, 0x65, 0x64, 0x42,
	0x79, 0x12, 0x20, 0x0a, 0x0b, 0x53, 0x75, 0x62, 0x6d, 0x69, 0x74, 0x74, 0x65, 0x64, 0x41, 0x74,
	0x18, 0x0a, 0x20, 0x01, 0x28, 0x04, 0x52, 0x0b, 0x53, 0x75, 0x62, 0x6d, 0x69, 0x74, 0x74, 0x65,
	0x64, 0x41, 0x74, 0x12, 0x30, 0x0a, 0x13, 0x4e, 0x75, 0x6d, 0x62, 0x65, 0x72, 0x4f, 0x66, 0x43,
	0x6f, 0x6e, 0x76, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e, 0x73, 0x18, 0x0b, 0x20, 0x01, 0x28, 0x0d,
	0x52, 0x13, 0x4e, 0x75, 0x6d, 0x62, 0x65, 0x72, 0x4f, 0x66, 0x43, 0x6f, 0x6e, 0x76, 0x65, 0x72,
	0x73, 0x69, 0x6f, 0x6e, 0x73, 0x12, 0x20, 0x0a, 0x0b, 0x44, 0x6f, 0x77, 0x6e, 0x6c, 0x6f, 0x61,
	0x64, 0x55, 0x52, 0x49, 0x18, 0x0c, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0b, 0x44, 0x6f, 0x77, 0x6e,
	0x6c, 0x6f, 0x61, 0x64, 0x55, 0x52, 0x49, 0x12, 0x26, 0x0a, 0x0e, 0x57, 0x6f, 0x72, 0x64, 0x73,
	0x50, 0x65, 0x72, 0x4d, 0x69, 0x6e, 0x75, 0x74, 0x65, 0x18, 0x0d, 0x20, 0x01, 0x28, 0x0d, 0x52,
	0x0e, 0x57, 0x6f, 0x72, 0x64, 0x73, 0x50, 0x65, 0x72, 0x4d, 0x69, 0x6e, 0x75, 0x74, 0x65, 0x12,
	0x20, 0x0a, 0x0b, 0x45, 0x53, 0x70, 0x65, 0x61, 0x6b, 0x56, 0x6f, 0x69, 0x63, 0x65, 0x18, 0x0e,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x0b, 0x45, 0x53, 0x70, 0x65, 0x61, 0x6b, 0x56, 0x6f, 0x69, 0x63,
	0x65, 0x12, 0x16, 0x0a, 0x06, 0x41, 0x54, 0x65, 0x6d, 0x70, 0x6f, 0x18, 0x0f, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x06, 0x41, 0x54, 0x65, 0x6d, 0x70, 0x6f, 0x12, 0x24, 0x0a, 0x0d, 0x53, 0x75, 0x62,
	0x6d, 0x69, 0x74, 0x74, 0x65, 0x64, 0x57, 0x69, 0x74, 0x68, 0x18, 0x10, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x0d, 0x53, 0x75, 0x62, 0x6d, 0x69, 0x74, 0x74, 0x65, 0x64, 0x57, 0x69, 0x74, 0x68, 0x12,
	0x20, 0x0a, 0x0b, 0x53, 0x74, 0x6f, 0x72, 0x61, 0x67, 0x65, 0x50, 0x61, 0x74, 0x68, 0x18, 0x11,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x0b, 0x53, 0x74, 0x6f, 0x72, 0x61, 0x67, 0x65, 0x50, 0x61, 0x74,
	0x68, 0x12, 0x12, 0x0a, 0x04, 0x54, 0x61, 0x67, 0x73, 0x18, 0x12, 0x20, 0x03, 0x28, 0x09, 0x52,
	0x04, 0x54, 0x61, 0x67, 0x73, 0x22, 0x4f, 0x0a, 0x08, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73,
	0x65, 0x12, 0x20, 0x0a, 0x0b, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x48, 0x61, 0x73, 0x68,
	0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0b, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x48,
	0x61, 0x73, 0x68, 0x12, 0x21, 0x0a, 0x04, 0x43, 0x6f, 0x64, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28,
	0x0e, 0x32, 0x0d, 0x2e, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x43, 0x6f, 0x64, 0x65,
	0x52, 0x04, 0x43, 0x6f, 0x64, 0x65, 0x2a, 0x5e, 0x0a, 0x0b, 0x43, 0x6f, 0x6e, 0x74, 0x65, 0x6e,
	0x74, 0x54, 0x79, 0x70, 0x65, 0x12, 0x07, 0x0a, 0x03, 0x55, 0x52, 0x49, 0x10, 0x00, 0x12, 0x0b,
	0x0a, 0x07, 0x59, 0x4f, 0x55, 0x54, 0x55, 0x42, 0x45, 0x10, 0x01, 0x12, 0x08, 0x0a, 0x04, 0x54,
	0x45, 0x58, 0x54, 0x10, 0x02, 0x12, 0x09, 0x0a, 0x05, 0x41, 0x55, 0x44, 0x49, 0x4f, 0x10, 0x03,
	0x12, 0x09, 0x0a, 0x05, 0x56, 0x49, 0x44, 0x45, 0x4f, 0x10, 0x04, 0x12, 0x07, 0x0a, 0x03, 0x50,
	0x44, 0x46, 0x10, 0x05, 0x12, 0x10, 0x0a, 0x0c, 0x59, 0x4f, 0x55, 0x54, 0x55, 0x42, 0x45, 0x5f,
	0x4c, 0x49, 0x53, 0x54, 0x10, 0x06, 0x2a, 0x27, 0x0a, 0x0c, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e,
	0x73, 0x65, 0x43, 0x6f, 0x64, 0x65, 0x12, 0x09, 0x0a, 0x05, 0x45, 0x52, 0x52, 0x4f, 0x52, 0x10,
	0x00, 0x12, 0x0c, 0x0a, 0x08, 0x41, 0x43, 0x43, 0x45, 0x50, 0x54, 0x45, 0x44, 0x10, 0x01, 0x2a,
	0x48, 0x0a, 0x09, 0x42, 0x6f, 0x74, 0x41, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x12, 0x0e, 0x0a, 0x0a,
	0x53, 0x45, 0x54, 0x5f, 0x43, 0x4f, 0x4e, 0x46, 0x49, 0x47, 0x10, 0x00, 0x12, 0x0e, 0x0a, 0x0a,
	0x47, 0x45, 0x54, 0x5f, 0x43, 0x4f, 0x4e, 0x46, 0x49, 0x47, 0x10, 0x01, 0x12, 0x11, 0x0a, 0x0d,
	0x53, 0x48, 0x4f, 0x57, 0x5f, 0x53, 0x45, 0x54, 0x54, 0x49, 0x4e, 0x47, 0x53, 0x10, 0x02, 0x12,
	0x08, 0x0a, 0x04, 0x48, 0x45, 0x4c, 0x50, 0x10, 0x04, 0x42, 0x0d, 0x5a, 0x0b, 0x2e, 0x3b, 0x67,
	0x65, 0x6e, 0x65, 0x72, 0x61, 0x74, 0x65, 0x64, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_rhema_proto_rawDescOnce sync.Once
	file_rhema_proto_rawDescData = file_rhema_proto_rawDesc
)

func file_rhema_proto_rawDescGZIP() []byte {
	file_rhema_proto_rawDescOnce.Do(func() {
		file_rhema_proto_rawDescData = protoimpl.X.CompressGZIP(file_rhema_proto_rawDescData)
	})
	return file_rhema_proto_rawDescData
}

var file_rhema_proto_enumTypes = make([]protoimpl.EnumInfo, 3)
var file_rhema_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_rhema_proto_goTypes = []interface{}{
	(ContentType)(0),  // 0: ContentType
	(ResponseCode)(0), // 1: ResponseCode
	(BotAction)(0),    // 2: BotAction
	(*Request)(nil),   // 3: Request
	(*Response)(nil),  // 4: Response
}
var file_rhema_proto_depIdxs = []int32{
	0, // 0: Request.Type:type_name -> ContentType
	1, // 1: Response.Code:type_name -> ResponseCode
	2, // [2:2] is the sub-list for method output_type
	2, // [2:2] is the sub-list for method input_type
	2, // [2:2] is the sub-list for extension type_name
	2, // [2:2] is the sub-list for extension extendee
	0, // [0:2] is the sub-list for field type_name
}

func init() { file_rhema_proto_init() }
func file_rhema_proto_init() {
	if File_rhema_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_rhema_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Request); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_rhema_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Response); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_rhema_proto_rawDesc,
			NumEnums:      3,
			NumMessages:   2,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_rhema_proto_goTypes,
		DependencyIndexes: file_rhema_proto_depIdxs,
		EnumInfos:         file_rhema_proto_enumTypes,
		MessageInfos:      file_rhema_proto_msgTypes,
	}.Build()
	File_rhema_proto = out.File
	file_rhema_proto_rawDesc = nil
	file_rhema_proto_goTypes = nil
	file_rhema_proto_depIdxs = nil
}
