// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.32.0
// 	protoc        v3.21.0
// source: data_model.proto

package vearchpb

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

type FieldType int32

const (
	FieldType_INT         FieldType = 0
	FieldType_LONG        FieldType = 1
	FieldType_FLOAT       FieldType = 2
	FieldType_DOUBLE      FieldType = 3
	FieldType_STRING      FieldType = 4
	FieldType_VECTOR      FieldType = 5
	FieldType_BOOL        FieldType = 6
	FieldType_DATE        FieldType = 7
	FieldType_STRINGARRAY FieldType = 8
)

// Enum value maps for FieldType.
var (
	FieldType_name = map[int32]string{
		0: "INT",
		1: "LONG",
		2: "FLOAT",
		3: "DOUBLE",
		4: "STRING",
		5: "VECTOR",
		6: "BOOL",
		7: "DATE",
		8: "STRINGARRAY",
	}
	FieldType_value = map[string]int32{
		"INT":         0,
		"LONG":        1,
		"FLOAT":       2,
		"DOUBLE":      3,
		"STRING":      4,
		"VECTOR":      5,
		"BOOL":        6,
		"DATE":        7,
		"STRINGARRAY": 8,
	}
)

func (x FieldType) Enum() *FieldType {
	p := new(FieldType)
	*p = x
	return p
}

func (x FieldType) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (FieldType) Descriptor() protoreflect.EnumDescriptor {
	return file_data_model_proto_enumTypes[0].Descriptor()
}

func (FieldType) Type() protoreflect.EnumType {
	return &file_data_model_proto_enumTypes[0]
}

func (x FieldType) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use FieldType.Descriptor instead.
func (FieldType) EnumDescriptor() ([]byte, []int) {
	return file_data_model_proto_rawDescGZIP(), []int{0}
}

// Whether index this field
type FieldOption int32

const (
	FieldOption_Null  FieldOption = 0
	FieldOption_Index FieldOption = 1
)

// Enum value maps for FieldOption.
var (
	FieldOption_name = map[int32]string{
		0: "Null",
		1: "Index",
	}
	FieldOption_value = map[string]int32{
		"Null":  0,
		"Index": 1,
	}
)

func (x FieldOption) Enum() *FieldOption {
	p := new(FieldOption)
	*p = x
	return p
}

func (x FieldOption) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (FieldOption) Descriptor() protoreflect.EnumDescriptor {
	return file_data_model_proto_enumTypes[1].Descriptor()
}

func (FieldOption) Type() protoreflect.EnumType {
	return &file_data_model_proto_enumTypes[1]
}

func (x FieldOption) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use FieldOption.Descriptor instead.
func (FieldOption) EnumDescriptor() ([]byte, []int) {
	return file_data_model_proto_rawDescGZIP(), []int{1}
}

type VectorMetaInfo_ValueType int32

const (
	VectorMetaInfo_FLOAT VectorMetaInfo_ValueType = 0
	VectorMetaInfo_UINT8 VectorMetaInfo_ValueType = 1 // binary
)

// Enum value maps for VectorMetaInfo_ValueType.
var (
	VectorMetaInfo_ValueType_name = map[int32]string{
		0: "FLOAT",
		1: "UINT8",
	}
	VectorMetaInfo_ValueType_value = map[string]int32{
		"FLOAT": 0,
		"UINT8": 1,
	}
)

func (x VectorMetaInfo_ValueType) Enum() *VectorMetaInfo_ValueType {
	p := new(VectorMetaInfo_ValueType)
	*p = x
	return p
}

func (x VectorMetaInfo_ValueType) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (VectorMetaInfo_ValueType) Descriptor() protoreflect.EnumDescriptor {
	return file_data_model_proto_enumTypes[2].Descriptor()
}

func (VectorMetaInfo_ValueType) Type() protoreflect.EnumType {
	return &file_data_model_proto_enumTypes[2]
}

func (x VectorMetaInfo_ValueType) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use VectorMetaInfo_ValueType.Descriptor instead.
func (VectorMetaInfo_ValueType) EnumDescriptor() ([]byte, []int) {
	return file_data_model_proto_rawDescGZIP(), []int{3, 0}
}

type VectorMetaInfo_StoreType int32

const (
	VectorMetaInfo_MEMORYONLY VectorMetaInfo_StoreType = 0
	VectorMetaInfo_ROCKSDB    VectorMetaInfo_StoreType = 1
)

// Enum value maps for VectorMetaInfo_StoreType.
var (
	VectorMetaInfo_StoreType_name = map[int32]string{
		0: "MEMORYONLY",
		1: "ROCKSDB",
	}
	VectorMetaInfo_StoreType_value = map[string]int32{
		"MEMORYONLY": 0,
		"ROCKSDB":    1,
	}
)

func (x VectorMetaInfo_StoreType) Enum() *VectorMetaInfo_StoreType {
	p := new(VectorMetaInfo_StoreType)
	*p = x
	return p
}

func (x VectorMetaInfo_StoreType) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (VectorMetaInfo_StoreType) Descriptor() protoreflect.EnumDescriptor {
	return file_data_model_proto_enumTypes[3].Descriptor()
}

func (VectorMetaInfo_StoreType) Type() protoreflect.EnumType {
	return &file_data_model_proto_enumTypes[3]
}

func (x VectorMetaInfo_StoreType) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use VectorMetaInfo_StoreType.Descriptor instead.
func (VectorMetaInfo_StoreType) EnumDescriptor() ([]byte, []int) {
	return file_data_model_proto_rawDescGZIP(), []int{3, 1}
}

type Field struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Name   string      `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	Type   FieldType   `protobuf:"varint,2,opt,name=type,proto3,enum=vearchpb.FieldType" json:"type,omitempty"`
	Value  []byte      `protobuf:"bytes,3,opt,name=value,proto3" json:"value,omitempty"`
	Option FieldOption `protobuf:"varint,4,opt,name=option,proto3,enum=vearchpb.FieldOption" json:"option,omitempty"`
}

func (x *Field) Reset() {
	*x = Field{}
	if protoimpl.UnsafeEnabled {
		mi := &file_data_model_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Field) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Field) ProtoMessage() {}

func (x *Field) ProtoReflect() protoreflect.Message {
	mi := &file_data_model_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Field.ProtoReflect.Descriptor instead.
func (*Field) Descriptor() ([]byte, []int) {
	return file_data_model_proto_rawDescGZIP(), []int{0}
}

func (x *Field) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *Field) GetType() FieldType {
	if x != nil {
		return x.Type
	}
	return FieldType_INT
}

func (x *Field) GetValue() []byte {
	if x != nil {
		return x.Value
	}
	return nil
}

func (x *Field) GetOption() FieldOption {
	if x != nil {
		return x.Option
	}
	return FieldOption_Null
}

type Document struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	PKey   string   `protobuf:"bytes,1,opt,name=p_key,json=pKey,proto3" json:"p_key,omitempty"`
	Fields []*Field `protobuf:"bytes,2,rep,name=fields,proto3" json:"fields,omitempty"`
}

func (x *Document) Reset() {
	*x = Document{}
	if protoimpl.UnsafeEnabled {
		mi := &file_data_model_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Document) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Document) ProtoMessage() {}

func (x *Document) ProtoReflect() protoreflect.Message {
	mi := &file_data_model_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Document.ProtoReflect.Descriptor instead.
func (*Document) Descriptor() ([]byte, []int) {
	return file_data_model_proto_rawDescGZIP(), []int{1}
}

func (x *Document) GetPKey() string {
	if x != nil {
		return x.PKey
	}
	return ""
}

func (x *Document) GetFields() []*Field {
	if x != nil {
		return x.Fields
	}
	return nil
}

type Item struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Err *Error    `protobuf:"bytes,1,opt,name=err,proto3" json:"err,omitempty"`
	Doc *Document `protobuf:"bytes,2,opt,name=doc,proto3" json:"doc,omitempty"`
	Msg string    `protobuf:"bytes,3,opt,name=msg,proto3" json:"msg,omitempty"`
}

func (x *Item) Reset() {
	*x = Item{}
	if protoimpl.UnsafeEnabled {
		mi := &file_data_model_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Item) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Item) ProtoMessage() {}

func (x *Item) ProtoReflect() protoreflect.Message {
	mi := &file_data_model_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Item.ProtoReflect.Descriptor instead.
func (*Item) Descriptor() ([]byte, []int) {
	return file_data_model_proto_rawDescGZIP(), []int{2}
}

func (x *Item) GetErr() *Error {
	if x != nil {
		return x.Err
	}
	return nil
}

func (x *Item) GetDoc() *Document {
	if x != nil {
		return x.Doc
	}
	return nil
}

func (x *Item) GetMsg() string {
	if x != nil {
		return x.Msg
	}
	return ""
}

type VectorMetaInfo struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Dimension  int32                    `protobuf:"varint,1,opt,name=dimension,proto3" json:"dimension,omitempty"`
	ValueType  VectorMetaInfo_ValueType `protobuf:"varint,2,opt,name=value_type,json=valueType,proto3,enum=vearchpb.VectorMetaInfo_ValueType" json:"value_type,omitempty"`
	StoreType  VectorMetaInfo_StoreType `protobuf:"varint,3,opt,name=store_type,json=storeType,proto3,enum=vearchpb.VectorMetaInfo_StoreType" json:"store_type,omitempty"`
	StoreParam string                   `protobuf:"bytes,4,opt,name=store_param,json=storeParam,proto3" json:"store_param,omitempty"`
}

func (x *VectorMetaInfo) Reset() {
	*x = VectorMetaInfo{}
	if protoimpl.UnsafeEnabled {
		mi := &file_data_model_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *VectorMetaInfo) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*VectorMetaInfo) ProtoMessage() {}

func (x *VectorMetaInfo) ProtoReflect() protoreflect.Message {
	mi := &file_data_model_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use VectorMetaInfo.ProtoReflect.Descriptor instead.
func (*VectorMetaInfo) Descriptor() ([]byte, []int) {
	return file_data_model_proto_rawDescGZIP(), []int{3}
}

func (x *VectorMetaInfo) GetDimension() int32 {
	if x != nil {
		return x.Dimension
	}
	return 0
}

func (x *VectorMetaInfo) GetValueType() VectorMetaInfo_ValueType {
	if x != nil {
		return x.ValueType
	}
	return VectorMetaInfo_FLOAT
}

func (x *VectorMetaInfo) GetStoreType() VectorMetaInfo_StoreType {
	if x != nil {
		return x.StoreType
	}
	return VectorMetaInfo_MEMORYONLY
}

func (x *VectorMetaInfo) GetStoreParam() string {
	if x != nil {
		return x.StoreParam
	}
	return ""
}

type FieldMetaInfo struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Name           string          `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	DataType       FieldType       `protobuf:"varint,2,opt,name=data_type,json=dataType,proto3,enum=vearchpb.FieldType" json:"data_type,omitempty"`
	IsIndex        bool            `protobuf:"varint,3,opt,name=is_index,json=isIndex,proto3" json:"is_index,omitempty"`
	VectorMetaInfo *VectorMetaInfo `protobuf:"bytes,4,opt,name=vector_meta_info,json=vectorMetaInfo,proto3" json:"vector_meta_info,omitempty"` // nil if data_type is not vector
}

func (x *FieldMetaInfo) Reset() {
	*x = FieldMetaInfo{}
	if protoimpl.UnsafeEnabled {
		mi := &file_data_model_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *FieldMetaInfo) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*FieldMetaInfo) ProtoMessage() {}

func (x *FieldMetaInfo) ProtoReflect() protoreflect.Message {
	mi := &file_data_model_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use FieldMetaInfo.ProtoReflect.Descriptor instead.
func (*FieldMetaInfo) Descriptor() ([]byte, []int) {
	return file_data_model_proto_rawDescGZIP(), []int{4}
}

func (x *FieldMetaInfo) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *FieldMetaInfo) GetDataType() FieldType {
	if x != nil {
		return x.DataType
	}
	return FieldType_INT
}

func (x *FieldMetaInfo) GetIsIndex() bool {
	if x != nil {
		return x.IsIndex
	}
	return false
}

func (x *FieldMetaInfo) GetVectorMetaInfo() *VectorMetaInfo {
	if x != nil {
		return x.VectorMetaInfo
	}
	return nil
}

type TableMetaInfo struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	PrimaryKeyName string           `protobuf:"bytes,1,opt,name=primary_key_name,json=primaryKeyName,proto3" json:"primary_key_name,omitempty"`
	PrimaryKeyType FieldType        `protobuf:"varint,2,opt,name=primary_key_type,json=primaryKeyType,proto3,enum=vearchpb.FieldType" json:"primary_key_type,omitempty"`
	PartitionsNum  int32            `protobuf:"varint,3,opt,name=partitions_num,json=partitionsNum,proto3" json:"partitions_num,omitempty"`
	ReplicasNum    int32            `protobuf:"varint,4,opt,name=replicas_num,json=replicasNum,proto3" json:"replicas_num,omitempty"`
	FieldMetaInfo  []*FieldMetaInfo `protobuf:"bytes,5,rep,name=field_meta_info,json=fieldMetaInfo,proto3" json:"field_meta_info,omitempty"`
}

func (x *TableMetaInfo) Reset() {
	*x = TableMetaInfo{}
	if protoimpl.UnsafeEnabled {
		mi := &file_data_model_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *TableMetaInfo) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*TableMetaInfo) ProtoMessage() {}

func (x *TableMetaInfo) ProtoReflect() protoreflect.Message {
	mi := &file_data_model_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use TableMetaInfo.ProtoReflect.Descriptor instead.
func (*TableMetaInfo) Descriptor() ([]byte, []int) {
	return file_data_model_proto_rawDescGZIP(), []int{5}
}

func (x *TableMetaInfo) GetPrimaryKeyName() string {
	if x != nil {
		return x.PrimaryKeyName
	}
	return ""
}

func (x *TableMetaInfo) GetPrimaryKeyType() FieldType {
	if x != nil {
		return x.PrimaryKeyType
	}
	return FieldType_INT
}

func (x *TableMetaInfo) GetPartitionsNum() int32 {
	if x != nil {
		return x.PartitionsNum
	}
	return 0
}

func (x *TableMetaInfo) GetReplicasNum() int32 {
	if x != nil {
		return x.ReplicasNum
	}
	return 0
}

func (x *TableMetaInfo) GetFieldMetaInfo() []*FieldMetaInfo {
	if x != nil {
		return x.FieldMetaInfo
	}
	return nil
}

type Table struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Name          string         `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	TableMetaInfo *TableMetaInfo `protobuf:"bytes,2,opt,name=table_meta_info,json=tableMetaInfo,proto3" json:"table_meta_info,omitempty"`
}

func (x *Table) Reset() {
	*x = Table{}
	if protoimpl.UnsafeEnabled {
		mi := &file_data_model_proto_msgTypes[6]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Table) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Table) ProtoMessage() {}

func (x *Table) ProtoReflect() protoreflect.Message {
	mi := &file_data_model_proto_msgTypes[6]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Table.ProtoReflect.Descriptor instead.
func (*Table) Descriptor() ([]byte, []int) {
	return file_data_model_proto_rawDescGZIP(), []int{6}
}

func (x *Table) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *Table) GetTableMetaInfo() *TableMetaInfo {
	if x != nil {
		return x.TableMetaInfo
	}
	return nil
}

type DB struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Name             string            `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	Tables           []*Table          `protobuf:"bytes,2,rep,name=tables,proto3" json:"tables,omitempty"`
	UserPasswordPair map[string]string `protobuf:"bytes,3,rep,name=user_password_pair,json=userPasswordPair,proto3" json:"user_password_pair,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
}

func (x *DB) Reset() {
	*x = DB{}
	if protoimpl.UnsafeEnabled {
		mi := &file_data_model_proto_msgTypes[7]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *DB) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*DB) ProtoMessage() {}

func (x *DB) ProtoReflect() protoreflect.Message {
	mi := &file_data_model_proto_msgTypes[7]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use DB.ProtoReflect.Descriptor instead.
func (*DB) Descriptor() ([]byte, []int) {
	return file_data_model_proto_rawDescGZIP(), []int{7}
}

func (x *DB) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *DB) GetTables() []*Table {
	if x != nil {
		return x.Tables
	}
	return nil
}

func (x *DB) GetUserPasswordPair() map[string]string {
	if x != nil {
		return x.UserPasswordPair
	}
	return nil
}

var File_data_model_proto protoreflect.FileDescriptor

var file_data_model_proto_rawDesc = []byte{
	0x0a, 0x10, 0x64, 0x61, 0x74, 0x61, 0x5f, 0x6d, 0x6f, 0x64, 0x65, 0x6c, 0x2e, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x12, 0x08, 0x76, 0x65, 0x61, 0x72, 0x63, 0x68, 0x70, 0x62, 0x1a, 0x0c, 0x65, 0x72,
	0x72, 0x6f, 0x72, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x89, 0x01, 0x0a, 0x05, 0x46,
	0x69, 0x65, 0x6c, 0x64, 0x12, 0x12, 0x0a, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x01, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x12, 0x27, 0x0a, 0x04, 0x74, 0x79, 0x70, 0x65,
	0x18, 0x02, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x13, 0x2e, 0x76, 0x65, 0x61, 0x72, 0x63, 0x68, 0x70,
	0x62, 0x2e, 0x46, 0x69, 0x65, 0x6c, 0x64, 0x54, 0x79, 0x70, 0x65, 0x52, 0x04, 0x74, 0x79, 0x70,
	0x65, 0x12, 0x14, 0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0c,
	0x52, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x12, 0x2d, 0x0a, 0x06, 0x6f, 0x70, 0x74, 0x69, 0x6f,
	0x6e, 0x18, 0x04, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x15, 0x2e, 0x76, 0x65, 0x61, 0x72, 0x63, 0x68,
	0x70, 0x62, 0x2e, 0x46, 0x69, 0x65, 0x6c, 0x64, 0x4f, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x52, 0x06,
	0x6f, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x22, 0x48, 0x0a, 0x08, 0x44, 0x6f, 0x63, 0x75, 0x6d, 0x65,
	0x6e, 0x74, 0x12, 0x13, 0x0a, 0x05, 0x70, 0x5f, 0x6b, 0x65, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x04, 0x70, 0x4b, 0x65, 0x79, 0x12, 0x27, 0x0a, 0x06, 0x66, 0x69, 0x65, 0x6c, 0x64,
	0x73, 0x18, 0x02, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x0f, 0x2e, 0x76, 0x65, 0x61, 0x72, 0x63, 0x68,
	0x70, 0x62, 0x2e, 0x46, 0x69, 0x65, 0x6c, 0x64, 0x52, 0x06, 0x66, 0x69, 0x65, 0x6c, 0x64, 0x73,
	0x22, 0x61, 0x0a, 0x04, 0x49, 0x74, 0x65, 0x6d, 0x12, 0x21, 0x0a, 0x03, 0x65, 0x72, 0x72, 0x18,
	0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x0f, 0x2e, 0x76, 0x65, 0x61, 0x72, 0x63, 0x68, 0x70, 0x62,
	0x2e, 0x45, 0x72, 0x72, 0x6f, 0x72, 0x52, 0x03, 0x65, 0x72, 0x72, 0x12, 0x24, 0x0a, 0x03, 0x64,
	0x6f, 0x63, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x12, 0x2e, 0x76, 0x65, 0x61, 0x72, 0x63,
	0x68, 0x70, 0x62, 0x2e, 0x44, 0x6f, 0x63, 0x75, 0x6d, 0x65, 0x6e, 0x74, 0x52, 0x03, 0x64, 0x6f,
	0x63, 0x12, 0x10, 0x0a, 0x03, 0x6d, 0x73, 0x67, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03,
	0x6d, 0x73, 0x67, 0x22, 0xa2, 0x02, 0x0a, 0x0e, 0x56, 0x65, 0x63, 0x74, 0x6f, 0x72, 0x4d, 0x65,
	0x74, 0x61, 0x49, 0x6e, 0x66, 0x6f, 0x12, 0x1c, 0x0a, 0x09, 0x64, 0x69, 0x6d, 0x65, 0x6e, 0x73,
	0x69, 0x6f, 0x6e, 0x18, 0x01, 0x20, 0x01, 0x28, 0x05, 0x52, 0x09, 0x64, 0x69, 0x6d, 0x65, 0x6e,
	0x73, 0x69, 0x6f, 0x6e, 0x12, 0x41, 0x0a, 0x0a, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x5f, 0x74, 0x79,
	0x70, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x22, 0x2e, 0x76, 0x65, 0x61, 0x72, 0x63,
	0x68, 0x70, 0x62, 0x2e, 0x56, 0x65, 0x63, 0x74, 0x6f, 0x72, 0x4d, 0x65, 0x74, 0x61, 0x49, 0x6e,
	0x66, 0x6f, 0x2e, 0x56, 0x61, 0x6c, 0x75, 0x65, 0x54, 0x79, 0x70, 0x65, 0x52, 0x09, 0x76, 0x61,
	0x6c, 0x75, 0x65, 0x54, 0x79, 0x70, 0x65, 0x12, 0x41, 0x0a, 0x0a, 0x73, 0x74, 0x6f, 0x72, 0x65,
	0x5f, 0x74, 0x79, 0x70, 0x65, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x22, 0x2e, 0x76, 0x65,
	0x61, 0x72, 0x63, 0x68, 0x70, 0x62, 0x2e, 0x56, 0x65, 0x63, 0x74, 0x6f, 0x72, 0x4d, 0x65, 0x74,
	0x61, 0x49, 0x6e, 0x66, 0x6f, 0x2e, 0x53, 0x74, 0x6f, 0x72, 0x65, 0x54, 0x79, 0x70, 0x65, 0x52,
	0x09, 0x73, 0x74, 0x6f, 0x72, 0x65, 0x54, 0x79, 0x70, 0x65, 0x12, 0x1f, 0x0a, 0x0b, 0x73, 0x74,
	0x6f, 0x72, 0x65, 0x5f, 0x70, 0x61, 0x72, 0x61, 0x6d, 0x18, 0x04, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x0a, 0x73, 0x74, 0x6f, 0x72, 0x65, 0x50, 0x61, 0x72, 0x61, 0x6d, 0x22, 0x21, 0x0a, 0x09, 0x56,
	0x61, 0x6c, 0x75, 0x65, 0x54, 0x79, 0x70, 0x65, 0x12, 0x09, 0x0a, 0x05, 0x46, 0x4c, 0x4f, 0x41,
	0x54, 0x10, 0x00, 0x12, 0x09, 0x0a, 0x05, 0x55, 0x49, 0x4e, 0x54, 0x38, 0x10, 0x01, 0x22, 0x28,
	0x0a, 0x09, 0x53, 0x74, 0x6f, 0x72, 0x65, 0x54, 0x79, 0x70, 0x65, 0x12, 0x0e, 0x0a, 0x0a, 0x4d,
	0x45, 0x4d, 0x4f, 0x52, 0x59, 0x4f, 0x4e, 0x4c, 0x59, 0x10, 0x00, 0x12, 0x0b, 0x0a, 0x07, 0x52,
	0x4f, 0x43, 0x4b, 0x53, 0x44, 0x42, 0x10, 0x01, 0x22, 0xb4, 0x01, 0x0a, 0x0d, 0x46, 0x69, 0x65,
	0x6c, 0x64, 0x4d, 0x65, 0x74, 0x61, 0x49, 0x6e, 0x66, 0x6f, 0x12, 0x12, 0x0a, 0x04, 0x6e, 0x61,
	0x6d, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x12, 0x30,
	0x0a, 0x09, 0x64, 0x61, 0x74, 0x61, 0x5f, 0x74, 0x79, 0x70, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28,
	0x0e, 0x32, 0x13, 0x2e, 0x76, 0x65, 0x61, 0x72, 0x63, 0x68, 0x70, 0x62, 0x2e, 0x46, 0x69, 0x65,
	0x6c, 0x64, 0x54, 0x79, 0x70, 0x65, 0x52, 0x08, 0x64, 0x61, 0x74, 0x61, 0x54, 0x79, 0x70, 0x65,
	0x12, 0x19, 0x0a, 0x08, 0x69, 0x73, 0x5f, 0x69, 0x6e, 0x64, 0x65, 0x78, 0x18, 0x03, 0x20, 0x01,
	0x28, 0x08, 0x52, 0x07, 0x69, 0x73, 0x49, 0x6e, 0x64, 0x65, 0x78, 0x12, 0x42, 0x0a, 0x10, 0x76,
	0x65, 0x63, 0x74, 0x6f, 0x72, 0x5f, 0x6d, 0x65, 0x74, 0x61, 0x5f, 0x69, 0x6e, 0x66, 0x6f, 0x18,
	0x04, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x18, 0x2e, 0x76, 0x65, 0x61, 0x72, 0x63, 0x68, 0x70, 0x62,
	0x2e, 0x56, 0x65, 0x63, 0x74, 0x6f, 0x72, 0x4d, 0x65, 0x74, 0x61, 0x49, 0x6e, 0x66, 0x6f, 0x52,
	0x0e, 0x76, 0x65, 0x63, 0x74, 0x6f, 0x72, 0x4d, 0x65, 0x74, 0x61, 0x49, 0x6e, 0x66, 0x6f, 0x22,
	0x83, 0x02, 0x0a, 0x0d, 0x54, 0x61, 0x62, 0x6c, 0x65, 0x4d, 0x65, 0x74, 0x61, 0x49, 0x6e, 0x66,
	0x6f, 0x12, 0x28, 0x0a, 0x10, 0x70, 0x72, 0x69, 0x6d, 0x61, 0x72, 0x79, 0x5f, 0x6b, 0x65, 0x79,
	0x5f, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0e, 0x70, 0x72, 0x69,
	0x6d, 0x61, 0x72, 0x79, 0x4b, 0x65, 0x79, 0x4e, 0x61, 0x6d, 0x65, 0x12, 0x3d, 0x0a, 0x10, 0x70,
	0x72, 0x69, 0x6d, 0x61, 0x72, 0x79, 0x5f, 0x6b, 0x65, 0x79, 0x5f, 0x74, 0x79, 0x70, 0x65, 0x18,
	0x02, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x13, 0x2e, 0x76, 0x65, 0x61, 0x72, 0x63, 0x68, 0x70, 0x62,
	0x2e, 0x46, 0x69, 0x65, 0x6c, 0x64, 0x54, 0x79, 0x70, 0x65, 0x52, 0x0e, 0x70, 0x72, 0x69, 0x6d,
	0x61, 0x72, 0x79, 0x4b, 0x65, 0x79, 0x54, 0x79, 0x70, 0x65, 0x12, 0x25, 0x0a, 0x0e, 0x70, 0x61,
	0x72, 0x74, 0x69, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x5f, 0x6e, 0x75, 0x6d, 0x18, 0x03, 0x20, 0x01,
	0x28, 0x05, 0x52, 0x0d, 0x70, 0x61, 0x72, 0x74, 0x69, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x4e, 0x75,
	0x6d, 0x12, 0x21, 0x0a, 0x0c, 0x72, 0x65, 0x70, 0x6c, 0x69, 0x63, 0x61, 0x73, 0x5f, 0x6e, 0x75,
	0x6d, 0x18, 0x04, 0x20, 0x01, 0x28, 0x05, 0x52, 0x0b, 0x72, 0x65, 0x70, 0x6c, 0x69, 0x63, 0x61,
	0x73, 0x4e, 0x75, 0x6d, 0x12, 0x3f, 0x0a, 0x0f, 0x66, 0x69, 0x65, 0x6c, 0x64, 0x5f, 0x6d, 0x65,
	0x74, 0x61, 0x5f, 0x69, 0x6e, 0x66, 0x6f, 0x18, 0x05, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x17, 0x2e,
	0x76, 0x65, 0x61, 0x72, 0x63, 0x68, 0x70, 0x62, 0x2e, 0x46, 0x69, 0x65, 0x6c, 0x64, 0x4d, 0x65,
	0x74, 0x61, 0x49, 0x6e, 0x66, 0x6f, 0x52, 0x0d, 0x66, 0x69, 0x65, 0x6c, 0x64, 0x4d, 0x65, 0x74,
	0x61, 0x49, 0x6e, 0x66, 0x6f, 0x22, 0x5c, 0x0a, 0x05, 0x54, 0x61, 0x62, 0x6c, 0x65, 0x12, 0x12,
	0x0a, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x6e, 0x61,
	0x6d, 0x65, 0x12, 0x3f, 0x0a, 0x0f, 0x74, 0x61, 0x62, 0x6c, 0x65, 0x5f, 0x6d, 0x65, 0x74, 0x61,
	0x5f, 0x69, 0x6e, 0x66, 0x6f, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x17, 0x2e, 0x76, 0x65,
	0x61, 0x72, 0x63, 0x68, 0x70, 0x62, 0x2e, 0x54, 0x61, 0x62, 0x6c, 0x65, 0x4d, 0x65, 0x74, 0x61,
	0x49, 0x6e, 0x66, 0x6f, 0x52, 0x0d, 0x74, 0x61, 0x62, 0x6c, 0x65, 0x4d, 0x65, 0x74, 0x61, 0x49,
	0x6e, 0x66, 0x6f, 0x22, 0xd8, 0x01, 0x0a, 0x02, 0x44, 0x42, 0x12, 0x12, 0x0a, 0x04, 0x6e, 0x61,
	0x6d, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x12, 0x27,
	0x0a, 0x06, 0x74, 0x61, 0x62, 0x6c, 0x65, 0x73, 0x18, 0x02, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x0f,
	0x2e, 0x76, 0x65, 0x61, 0x72, 0x63, 0x68, 0x70, 0x62, 0x2e, 0x54, 0x61, 0x62, 0x6c, 0x65, 0x52,
	0x06, 0x74, 0x61, 0x62, 0x6c, 0x65, 0x73, 0x12, 0x50, 0x0a, 0x12, 0x75, 0x73, 0x65, 0x72, 0x5f,
	0x70, 0x61, 0x73, 0x73, 0x77, 0x6f, 0x72, 0x64, 0x5f, 0x70, 0x61, 0x69, 0x72, 0x18, 0x03, 0x20,
	0x03, 0x28, 0x0b, 0x32, 0x22, 0x2e, 0x76, 0x65, 0x61, 0x72, 0x63, 0x68, 0x70, 0x62, 0x2e, 0x44,
	0x42, 0x2e, 0x55, 0x73, 0x65, 0x72, 0x50, 0x61, 0x73, 0x73, 0x77, 0x6f, 0x72, 0x64, 0x50, 0x61,
	0x69, 0x72, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x52, 0x10, 0x75, 0x73, 0x65, 0x72, 0x50, 0x61, 0x73,
	0x73, 0x77, 0x6f, 0x72, 0x64, 0x50, 0x61, 0x69, 0x72, 0x1a, 0x43, 0x0a, 0x15, 0x55, 0x73, 0x65,
	0x72, 0x50, 0x61, 0x73, 0x73, 0x77, 0x6f, 0x72, 0x64, 0x50, 0x61, 0x69, 0x72, 0x45, 0x6e, 0x74,
	0x72, 0x79, 0x12, 0x10, 0x0a, 0x03, 0x6b, 0x65, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x03, 0x6b, 0x65, 0x79, 0x12, 0x14, 0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x18, 0x02, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x3a, 0x02, 0x38, 0x01, 0x2a, 0x72,
	0x0a, 0x09, 0x46, 0x69, 0x65, 0x6c, 0x64, 0x54, 0x79, 0x70, 0x65, 0x12, 0x07, 0x0a, 0x03, 0x49,
	0x4e, 0x54, 0x10, 0x00, 0x12, 0x08, 0x0a, 0x04, 0x4c, 0x4f, 0x4e, 0x47, 0x10, 0x01, 0x12, 0x09,
	0x0a, 0x05, 0x46, 0x4c, 0x4f, 0x41, 0x54, 0x10, 0x02, 0x12, 0x0a, 0x0a, 0x06, 0x44, 0x4f, 0x55,
	0x42, 0x4c, 0x45, 0x10, 0x03, 0x12, 0x0a, 0x0a, 0x06, 0x53, 0x54, 0x52, 0x49, 0x4e, 0x47, 0x10,
	0x04, 0x12, 0x0a, 0x0a, 0x06, 0x56, 0x45, 0x43, 0x54, 0x4f, 0x52, 0x10, 0x05, 0x12, 0x08, 0x0a,
	0x04, 0x42, 0x4f, 0x4f, 0x4c, 0x10, 0x06, 0x12, 0x08, 0x0a, 0x04, 0x44, 0x41, 0x54, 0x45, 0x10,
	0x07, 0x12, 0x0f, 0x0a, 0x0b, 0x53, 0x54, 0x52, 0x49, 0x4e, 0x47, 0x41, 0x52, 0x52, 0x41, 0x59,
	0x10, 0x08, 0x2a, 0x22, 0x0a, 0x0b, 0x46, 0x69, 0x65, 0x6c, 0x64, 0x4f, 0x70, 0x74, 0x69, 0x6f,
	0x6e, 0x12, 0x08, 0x0a, 0x04, 0x4e, 0x75, 0x6c, 0x6c, 0x10, 0x00, 0x12, 0x09, 0x0a, 0x05, 0x49,
	0x6e, 0x64, 0x65, 0x78, 0x10, 0x01, 0x42, 0x0e, 0x48, 0x01, 0x5a, 0x0a, 0x2e, 0x2f, 0x76, 0x65,
	0x61, 0x72, 0x63, 0x68, 0x70, 0x62, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_data_model_proto_rawDescOnce sync.Once
	file_data_model_proto_rawDescData = file_data_model_proto_rawDesc
)

func file_data_model_proto_rawDescGZIP() []byte {
	file_data_model_proto_rawDescOnce.Do(func() {
		file_data_model_proto_rawDescData = protoimpl.X.CompressGZIP(file_data_model_proto_rawDescData)
	})
	return file_data_model_proto_rawDescData
}

var file_data_model_proto_enumTypes = make([]protoimpl.EnumInfo, 4)
var file_data_model_proto_msgTypes = make([]protoimpl.MessageInfo, 9)
var file_data_model_proto_goTypes = []interface{}{
	(FieldType)(0),                // 0: vearchpb.FieldType
	(FieldOption)(0),              // 1: vearchpb.FieldOption
	(VectorMetaInfo_ValueType)(0), // 2: vearchpb.VectorMetaInfo.ValueType
	(VectorMetaInfo_StoreType)(0), // 3: vearchpb.VectorMetaInfo.StoreType
	(*Field)(nil),                 // 4: vearchpb.Field
	(*Document)(nil),              // 5: vearchpb.Document
	(*Item)(nil),                  // 6: vearchpb.Item
	(*VectorMetaInfo)(nil),        // 7: vearchpb.VectorMetaInfo
	(*FieldMetaInfo)(nil),         // 8: vearchpb.FieldMetaInfo
	(*TableMetaInfo)(nil),         // 9: vearchpb.TableMetaInfo
	(*Table)(nil),                 // 10: vearchpb.Table
	(*DB)(nil),                    // 11: vearchpb.DB
	nil,                           // 12: vearchpb.DB.UserPasswordPairEntry
	(*Error)(nil),                 // 13: vearchpb.Error
}
var file_data_model_proto_depIdxs = []int32{
	0,  // 0: vearchpb.Field.type:type_name -> vearchpb.FieldType
	1,  // 1: vearchpb.Field.option:type_name -> vearchpb.FieldOption
	4,  // 2: vearchpb.Document.fields:type_name -> vearchpb.Field
	13, // 3: vearchpb.Item.err:type_name -> vearchpb.Error
	5,  // 4: vearchpb.Item.doc:type_name -> vearchpb.Document
	2,  // 5: vearchpb.VectorMetaInfo.value_type:type_name -> vearchpb.VectorMetaInfo.ValueType
	3,  // 6: vearchpb.VectorMetaInfo.store_type:type_name -> vearchpb.VectorMetaInfo.StoreType
	0,  // 7: vearchpb.FieldMetaInfo.data_type:type_name -> vearchpb.FieldType
	7,  // 8: vearchpb.FieldMetaInfo.vector_meta_info:type_name -> vearchpb.VectorMetaInfo
	0,  // 9: vearchpb.TableMetaInfo.primary_key_type:type_name -> vearchpb.FieldType
	8,  // 10: vearchpb.TableMetaInfo.field_meta_info:type_name -> vearchpb.FieldMetaInfo
	9,  // 11: vearchpb.Table.table_meta_info:type_name -> vearchpb.TableMetaInfo
	10, // 12: vearchpb.DB.tables:type_name -> vearchpb.Table
	12, // 13: vearchpb.DB.user_password_pair:type_name -> vearchpb.DB.UserPasswordPairEntry
	14, // [14:14] is the sub-list for method output_type
	14, // [14:14] is the sub-list for method input_type
	14, // [14:14] is the sub-list for extension type_name
	14, // [14:14] is the sub-list for extension extendee
	0,  // [0:14] is the sub-list for field type_name
}

func init() { file_data_model_proto_init() }
func file_data_model_proto_init() {
	if File_data_model_proto != nil {
		return
	}
	file_errors_proto_init()
	if !protoimpl.UnsafeEnabled {
		file_data_model_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Field); i {
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
		file_data_model_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Document); i {
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
		file_data_model_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Item); i {
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
		file_data_model_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*VectorMetaInfo); i {
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
		file_data_model_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*FieldMetaInfo); i {
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
		file_data_model_proto_msgTypes[5].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*TableMetaInfo); i {
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
		file_data_model_proto_msgTypes[6].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Table); i {
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
		file_data_model_proto_msgTypes[7].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*DB); i {
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
			RawDescriptor: file_data_model_proto_rawDesc,
			NumEnums:      4,
			NumMessages:   9,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_data_model_proto_goTypes,
		DependencyIndexes: file_data_model_proto_depIdxs,
		EnumInfos:         file_data_model_proto_enumTypes,
		MessageInfos:      file_data_model_proto_msgTypes,
	}.Build()
	File_data_model_proto = out.File
	file_data_model_proto_rawDesc = nil
	file_data_model_proto_goTypes = nil
	file_data_model_proto_depIdxs = nil
}
