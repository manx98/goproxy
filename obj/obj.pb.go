// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.28.0
// 	protoc        v3.21.2
// source: obj.proto

package obj

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

type DirEntry struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Id      []byte    `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	Entries []*Dirent `protobuf:"bytes,2,rep,name=entries,proto3" json:"entries,omitempty"`
}

func (x *DirEntry) Reset() {
	*x = DirEntry{}
	if protoimpl.UnsafeEnabled {
		mi := &file_obj_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *DirEntry) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*DirEntry) ProtoMessage() {}

func (x *DirEntry) ProtoReflect() protoreflect.Message {
	mi := &file_obj_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use DirEntry.ProtoReflect.Descriptor instead.
func (*DirEntry) Descriptor() ([]byte, []int) {
	return file_obj_proto_rawDescGZIP(), []int{0}
}

func (x *DirEntry) GetId() []byte {
	if x != nil {
		return x.Id
	}
	return nil
}

func (x *DirEntry) GetEntries() []*Dirent {
	if x != nil {
		return x.Entries
	}
	return nil
}

type Dirent struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Id    []byte `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	Name  string `protobuf:"bytes,2,opt,name=name,proto3" json:"name,omitempty"`
	IsDir bool   `protobuf:"varint,3,opt,name=is_dir,json=isDir,proto3" json:"is_dir,omitempty"`
	Mtime int64  `protobuf:"varint,4,opt,name=mtime,proto3" json:"mtime,omitempty"`
	Size  int64  `protobuf:"varint,5,opt,name=size,proto3" json:"size,omitempty"`
}

func (x *Dirent) Reset() {
	*x = Dirent{}
	if protoimpl.UnsafeEnabled {
		mi := &file_obj_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Dirent) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Dirent) ProtoMessage() {}

func (x *Dirent) ProtoReflect() protoreflect.Message {
	mi := &file_obj_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Dirent.ProtoReflect.Descriptor instead.
func (*Dirent) Descriptor() ([]byte, []int) {
	return file_obj_proto_rawDescGZIP(), []int{1}
}

func (x *Dirent) GetId() []byte {
	if x != nil {
		return x.Id
	}
	return nil
}

func (x *Dirent) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *Dirent) GetIsDir() bool {
	if x != nil {
		return x.IsDir
	}
	return false
}

func (x *Dirent) GetMtime() int64 {
	if x != nil {
		return x.Mtime
	}
	return 0
}

func (x *Dirent) GetSize() int64 {
	if x != nil {
		return x.Size
	}
	return 0
}

type CheckPoint struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Parent []byte `protobuf:"bytes,1,opt,name=parent,proto3" json:"parent,omitempty"`
	Id     []byte `protobuf:"bytes,2,opt,name=id,proto3" json:"id,omitempty"`
	Mtime  int64  `protobuf:"varint,3,opt,name=mtime,proto3" json:"mtime,omitempty"`
	Desc   string `protobuf:"bytes,4,opt,name=desc,proto3" json:"desc,omitempty"`
}

func (x *CheckPoint) Reset() {
	*x = CheckPoint{}
	if protoimpl.UnsafeEnabled {
		mi := &file_obj_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *CheckPoint) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CheckPoint) ProtoMessage() {}

func (x *CheckPoint) ProtoReflect() protoreflect.Message {
	mi := &file_obj_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use CheckPoint.ProtoReflect.Descriptor instead.
func (*CheckPoint) Descriptor() ([]byte, []int) {
	return file_obj_proto_rawDescGZIP(), []int{2}
}

func (x *CheckPoint) GetParent() []byte {
	if x != nil {
		return x.Parent
	}
	return nil
}

func (x *CheckPoint) GetId() []byte {
	if x != nil {
		return x.Id
	}
	return nil
}

func (x *CheckPoint) GetMtime() int64 {
	if x != nil {
		return x.Mtime
	}
	return 0
}

func (x *CheckPoint) GetDesc() string {
	if x != nil {
		return x.Desc
	}
	return ""
}

var File_obj_proto protoreflect.FileDescriptor

var file_obj_proto_rawDesc = []byte{
	0x0a, 0x09, 0x6f, 0x62, 0x6a, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x3d, 0x0a, 0x08, 0x44,
	0x69, 0x72, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x12, 0x0e, 0x0a, 0x02, 0x69, 0x64, 0x18, 0x01, 0x20,
	0x01, 0x28, 0x0c, 0x52, 0x02, 0x69, 0x64, 0x12, 0x21, 0x0a, 0x07, 0x65, 0x6e, 0x74, 0x72, 0x69,
	0x65, 0x73, 0x18, 0x02, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x07, 0x2e, 0x44, 0x69, 0x72, 0x65, 0x6e,
	0x74, 0x52, 0x07, 0x65, 0x6e, 0x74, 0x72, 0x69, 0x65, 0x73, 0x22, 0x6d, 0x0a, 0x06, 0x44, 0x69,
	0x72, 0x65, 0x6e, 0x74, 0x12, 0x0e, 0x0a, 0x02, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0c,
	0x52, 0x02, 0x69, 0x64, 0x12, 0x12, 0x0a, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x02, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x12, 0x15, 0x0a, 0x06, 0x69, 0x73, 0x5f, 0x64,
	0x69, 0x72, 0x18, 0x03, 0x20, 0x01, 0x28, 0x08, 0x52, 0x05, 0x69, 0x73, 0x44, 0x69, 0x72, 0x12,
	0x14, 0x0a, 0x05, 0x6d, 0x74, 0x69, 0x6d, 0x65, 0x18, 0x04, 0x20, 0x01, 0x28, 0x03, 0x52, 0x05,
	0x6d, 0x74, 0x69, 0x6d, 0x65, 0x12, 0x12, 0x0a, 0x04, 0x73, 0x69, 0x7a, 0x65, 0x18, 0x05, 0x20,
	0x01, 0x28, 0x03, 0x52, 0x04, 0x73, 0x69, 0x7a, 0x65, 0x22, 0x5e, 0x0a, 0x0a, 0x43, 0x68, 0x65,
	0x63, 0x6b, 0x50, 0x6f, 0x69, 0x6e, 0x74, 0x12, 0x16, 0x0a, 0x06, 0x70, 0x61, 0x72, 0x65, 0x6e,
	0x74, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x06, 0x70, 0x61, 0x72, 0x65, 0x6e, 0x74, 0x12,
	0x0e, 0x0a, 0x02, 0x69, 0x64, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x02, 0x69, 0x64, 0x12,
	0x14, 0x0a, 0x05, 0x6d, 0x74, 0x69, 0x6d, 0x65, 0x18, 0x03, 0x20, 0x01, 0x28, 0x03, 0x52, 0x05,
	0x6d, 0x74, 0x69, 0x6d, 0x65, 0x12, 0x12, 0x0a, 0x04, 0x64, 0x65, 0x73, 0x63, 0x18, 0x04, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x04, 0x64, 0x65, 0x73, 0x63, 0x42, 0x07, 0x5a, 0x05, 0x2e, 0x2f, 0x6f,
	0x62, 0x6a, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_obj_proto_rawDescOnce sync.Once
	file_obj_proto_rawDescData = file_obj_proto_rawDesc
)

func file_obj_proto_rawDescGZIP() []byte {
	file_obj_proto_rawDescOnce.Do(func() {
		file_obj_proto_rawDescData = protoimpl.X.CompressGZIP(file_obj_proto_rawDescData)
	})
	return file_obj_proto_rawDescData
}

var file_obj_proto_msgTypes = make([]protoimpl.MessageInfo, 3)
var file_obj_proto_goTypes = []interface{}{
	(*DirEntry)(nil),   // 0: DirEntry
	(*Dirent)(nil),     // 1: Dirent
	(*CheckPoint)(nil), // 2: CheckPoint
}
var file_obj_proto_depIdxs = []int32{
	1, // 0: DirEntry.entries:type_name -> Dirent
	1, // [1:1] is the sub-list for method output_type
	1, // [1:1] is the sub-list for method input_type
	1, // [1:1] is the sub-list for extension type_name
	1, // [1:1] is the sub-list for extension extendee
	0, // [0:1] is the sub-list for field type_name
}

func init() { file_obj_proto_init() }
func file_obj_proto_init() {
	if File_obj_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_obj_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*DirEntry); i {
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
		file_obj_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Dirent); i {
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
		file_obj_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*CheckPoint); i {
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
			RawDescriptor: file_obj_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   3,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_obj_proto_goTypes,
		DependencyIndexes: file_obj_proto_depIdxs,
		MessageInfos:      file_obj_proto_msgTypes,
	}.Build()
	File_obj_proto = out.File
	file_obj_proto_rawDesc = nil
	file_obj_proto_goTypes = nil
	file_obj_proto_depIdxs = nil
}
