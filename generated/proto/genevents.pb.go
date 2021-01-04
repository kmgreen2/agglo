// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.25.0
// 	protoc        v3.6.1
// source: genevents.proto

package api

import (
	proto "github.com/golang/protobuf/proto"
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

// This is a compile-time assertion that a sufficiently up-to-date version
// of the legacy proto package is being used.
const _ = proto.ProtoPackageIsVersion4

type RandomStringValue struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	MaxLen   int32  `protobuf:"varint,1,opt,name=maxLen,proto3" json:"maxLen,omitempty"`
	MinLen   int32  `protobuf:"varint,2,opt,name=minLen,proto3" json:"minLen,omitempty"`
	Alphabet string `protobuf:"bytes,3,opt,name=alphabet,proto3" json:"alphabet,omitempty"`
	MaxRef   int32  `protobuf:"varint,4,opt,name=maxRef,proto3" json:"maxRef,omitempty"`
}

func (x *RandomStringValue) Reset() {
	*x = RandomStringValue{}
	if protoimpl.UnsafeEnabled {
		mi := &file_genevents_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *RandomStringValue) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*RandomStringValue) ProtoMessage() {}

func (x *RandomStringValue) ProtoReflect() protoreflect.Message {
	mi := &file_genevents_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use RandomStringValue.ProtoReflect.Descriptor instead.
func (*RandomStringValue) Descriptor() ([]byte, []int) {
	return file_genevents_proto_rawDescGZIP(), []int{0}
}

func (x *RandomStringValue) GetMaxLen() int32 {
	if x != nil {
		return x.MaxLen
	}
	return 0
}

func (x *RandomStringValue) GetMinLen() int32 {
	if x != nil {
		return x.MinLen
	}
	return 0
}

func (x *RandomStringValue) GetAlphabet() string {
	if x != nil {
		return x.Alphabet
	}
	return ""
}

func (x *RandomStringValue) GetMaxRef() int32 {
	if x != nil {
		return x.MaxRef
	}
	return 0
}

type VocabStringValue struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Vocab  []string `protobuf:"bytes,1,rep,name=vocab,proto3" json:"vocab,omitempty"`
	MaxRef int32    `protobuf:"varint,4,opt,name=maxRef,proto3" json:"maxRef,omitempty"`
}

func (x *VocabStringValue) Reset() {
	*x = VocabStringValue{}
	if protoimpl.UnsafeEnabled {
		mi := &file_genevents_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *VocabStringValue) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*VocabStringValue) ProtoMessage() {}

func (x *VocabStringValue) ProtoReflect() protoreflect.Message {
	mi := &file_genevents_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use VocabStringValue.ProtoReflect.Descriptor instead.
func (*VocabStringValue) Descriptor() ([]byte, []int) {
	return file_genevents_proto_rawDescGZIP(), []int{1}
}

func (x *VocabStringValue) GetVocab() []string {
	if x != nil {
		return x.Vocab
	}
	return nil
}

func (x *VocabStringValue) GetMaxRef() int32 {
	if x != nil {
		return x.MaxRef
	}
	return 0
}

type FixedStringValue struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Value string `protobuf:"bytes,1,opt,name=value,proto3" json:"value,omitempty"`
}

func (x *FixedStringValue) Reset() {
	*x = FixedStringValue{}
	if protoimpl.UnsafeEnabled {
		mi := &file_genevents_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *FixedStringValue) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*FixedStringValue) ProtoMessage() {}

func (x *FixedStringValue) ProtoReflect() protoreflect.Message {
	mi := &file_genevents_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use FixedStringValue.ProtoReflect.Descriptor instead.
func (*FixedStringValue) Descriptor() ([]byte, []int) {
	return file_genevents_proto_rawDescGZIP(), []int{2}
}

func (x *FixedStringValue) GetValue() string {
	if x != nil {
		return x.Value
	}
	return ""
}

type RandomNumericValue struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Min    float64 `protobuf:"fixed64,1,opt,name=min,proto3" json:"min,omitempty"`
	Max    float64 `protobuf:"fixed64,2,opt,name=max,proto3" json:"max,omitempty"`
	MaxRef int32   `protobuf:"varint,4,opt,name=maxRef,proto3" json:"maxRef,omitempty"`
}

func (x *RandomNumericValue) Reset() {
	*x = RandomNumericValue{}
	if protoimpl.UnsafeEnabled {
		mi := &file_genevents_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *RandomNumericValue) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*RandomNumericValue) ProtoMessage() {}

func (x *RandomNumericValue) ProtoReflect() protoreflect.Message {
	mi := &file_genevents_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use RandomNumericValue.ProtoReflect.Descriptor instead.
func (*RandomNumericValue) Descriptor() ([]byte, []int) {
	return file_genevents_proto_rawDescGZIP(), []int{3}
}

func (x *RandomNumericValue) GetMin() float64 {
	if x != nil {
		return x.Min
	}
	return 0
}

func (x *RandomNumericValue) GetMax() float64 {
	if x != nil {
		return x.Max
	}
	return 0
}

func (x *RandomNumericValue) GetMaxRef() int32 {
	if x != nil {
		return x.MaxRef
	}
	return 0
}

type NumericSetValue struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Values []float64 `protobuf:"fixed64,1,rep,packed,name=values,proto3" json:"values,omitempty"`
}

func (x *NumericSetValue) Reset() {
	*x = NumericSetValue{}
	if protoimpl.UnsafeEnabled {
		mi := &file_genevents_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *NumericSetValue) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*NumericSetValue) ProtoMessage() {}

func (x *NumericSetValue) ProtoReflect() protoreflect.Message {
	mi := &file_genevents_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use NumericSetValue.ProtoReflect.Descriptor instead.
func (*NumericSetValue) Descriptor() ([]byte, []int) {
	return file_genevents_proto_rawDescGZIP(), []int{4}
}

func (x *NumericSetValue) GetValues() []float64 {
	if x != nil {
		return x.Values
	}
	return nil
}

type FixedNumericValue struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Value float64 `protobuf:"fixed64,1,opt,name=value,proto3" json:"value,omitempty"`
}

func (x *FixedNumericValue) Reset() {
	*x = FixedNumericValue{}
	if protoimpl.UnsafeEnabled {
		mi := &file_genevents_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *FixedNumericValue) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*FixedNumericValue) ProtoMessage() {}

func (x *FixedNumericValue) ProtoReflect() protoreflect.Message {
	mi := &file_genevents_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use FixedNumericValue.ProtoReflect.Descriptor instead.
func (*FixedNumericValue) Descriptor() ([]byte, []int) {
	return file_genevents_proto_rawDescGZIP(), []int{5}
}

func (x *FixedNumericValue) GetValue() float64 {
	if x != nil {
		return x.Value
	}
	return 0
}

type BooleanValue struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *BooleanValue) Reset() {
	*x = BooleanValue{}
	if protoimpl.UnsafeEnabled {
		mi := &file_genevents_proto_msgTypes[6]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *BooleanValue) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*BooleanValue) ProtoMessage() {}

func (x *BooleanValue) ProtoReflect() protoreflect.Message {
	mi := &file_genevents_proto_msgTypes[6]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use BooleanValue.ProtoReflect.Descriptor instead.
func (*BooleanValue) Descriptor() ([]byte, []int) {
	return file_genevents_proto_rawDescGZIP(), []int{6}
}

type DictValue struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Kvs map[string]*Value `protobuf:"bytes,1,rep,name=kvs,proto3" json:"kvs,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
}

func (x *DictValue) Reset() {
	*x = DictValue{}
	if protoimpl.UnsafeEnabled {
		mi := &file_genevents_proto_msgTypes[7]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *DictValue) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*DictValue) ProtoMessage() {}

func (x *DictValue) ProtoReflect() protoreflect.Message {
	mi := &file_genevents_proto_msgTypes[7]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use DictValue.ProtoReflect.Descriptor instead.
func (*DictValue) Descriptor() ([]byte, []int) {
	return file_genevents_proto_rawDescGZIP(), []int{7}
}

func (x *DictValue) GetKvs() map[string]*Value {
	if x != nil {
		return x.Kvs
	}
	return nil
}

type ListValue struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	MaxLen int32  `protobuf:"varint,1,opt,name=maxLen,proto3" json:"maxLen,omitempty"`
	MinLen int32  `protobuf:"varint,2,opt,name=minLen,proto3" json:"minLen,omitempty"`
	Value  *Value `protobuf:"bytes,3,opt,name=value,proto3" json:"value,omitempty"`
}

func (x *ListValue) Reset() {
	*x = ListValue{}
	if protoimpl.UnsafeEnabled {
		mi := &file_genevents_proto_msgTypes[8]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ListValue) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ListValue) ProtoMessage() {}

func (x *ListValue) ProtoReflect() protoreflect.Message {
	mi := &file_genevents_proto_msgTypes[8]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ListValue.ProtoReflect.Descriptor instead.
func (*ListValue) Descriptor() ([]byte, []int) {
	return file_genevents_proto_rawDescGZIP(), []int{8}
}

func (x *ListValue) GetMaxLen() int32 {
	if x != nil {
		return x.MaxLen
	}
	return 0
}

func (x *ListValue) GetMinLen() int32 {
	if x != nil {
		return x.MinLen
	}
	return 0
}

func (x *ListValue) GetValue() *Value {
	if x != nil {
		return x.Value
	}
	return nil
}

type ReferenceValue struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	SchemaName string `protobuf:"bytes,1,opt,name=schemaName,proto3" json:"schemaName,omitempty"`
	Path       string `protobuf:"bytes,2,opt,name=path,proto3" json:"path,omitempty"`
}

func (x *ReferenceValue) Reset() {
	*x = ReferenceValue{}
	if protoimpl.UnsafeEnabled {
		mi := &file_genevents_proto_msgTypes[9]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ReferenceValue) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ReferenceValue) ProtoMessage() {}

func (x *ReferenceValue) ProtoReflect() protoreflect.Message {
	mi := &file_genevents_proto_msgTypes[9]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ReferenceValue.ProtoReflect.Descriptor instead.
func (*ReferenceValue) Descriptor() ([]byte, []int) {
	return file_genevents_proto_rawDescGZIP(), []int{9}
}

func (x *ReferenceValue) GetSchemaName() string {
	if x != nil {
		return x.SchemaName
	}
	return ""
}

func (x *ReferenceValue) GetPath() string {
	if x != nil {
		return x.Path
	}
	return ""
}

type Value struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Types that are assignable to Values:
	//	*Value_RandomString
	//	*Value_VocabString
	//	*Value_FixedString
	//	*Value_RandomNumeric
	//	*Value_NumericSet
	//	*Value_FixedNumeric
	//	*Value_Boolean
	//	*Value_Dict
	//	*Value_List
	//	*Value_Reference
	Values isValue_Values `protobuf_oneof:"values"`
}

func (x *Value) Reset() {
	*x = Value{}
	if protoimpl.UnsafeEnabled {
		mi := &file_genevents_proto_msgTypes[10]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Value) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Value) ProtoMessage() {}

func (x *Value) ProtoReflect() protoreflect.Message {
	mi := &file_genevents_proto_msgTypes[10]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Value.ProtoReflect.Descriptor instead.
func (*Value) Descriptor() ([]byte, []int) {
	return file_genevents_proto_rawDescGZIP(), []int{10}
}

func (m *Value) GetValues() isValue_Values {
	if m != nil {
		return m.Values
	}
	return nil
}

func (x *Value) GetRandomString() *RandomStringValue {
	if x, ok := x.GetValues().(*Value_RandomString); ok {
		return x.RandomString
	}
	return nil
}

func (x *Value) GetVocabString() *VocabStringValue {
	if x, ok := x.GetValues().(*Value_VocabString); ok {
		return x.VocabString
	}
	return nil
}

func (x *Value) GetFixedString() *FixedStringValue {
	if x, ok := x.GetValues().(*Value_FixedString); ok {
		return x.FixedString
	}
	return nil
}

func (x *Value) GetRandomNumeric() *RandomNumericValue {
	if x, ok := x.GetValues().(*Value_RandomNumeric); ok {
		return x.RandomNumeric
	}
	return nil
}

func (x *Value) GetNumericSet() *NumericSetValue {
	if x, ok := x.GetValues().(*Value_NumericSet); ok {
		return x.NumericSet
	}
	return nil
}

func (x *Value) GetFixedNumeric() *FixedNumericValue {
	if x, ok := x.GetValues().(*Value_FixedNumeric); ok {
		return x.FixedNumeric
	}
	return nil
}

func (x *Value) GetBoolean() *BooleanValue {
	if x, ok := x.GetValues().(*Value_Boolean); ok {
		return x.Boolean
	}
	return nil
}

func (x *Value) GetDict() *DictValue {
	if x, ok := x.GetValues().(*Value_Dict); ok {
		return x.Dict
	}
	return nil
}

func (x *Value) GetList() *ListValue {
	if x, ok := x.GetValues().(*Value_List); ok {
		return x.List
	}
	return nil
}

func (x *Value) GetReference() *ReferenceValue {
	if x, ok := x.GetValues().(*Value_Reference); ok {
		return x.Reference
	}
	return nil
}

type isValue_Values interface {
	isValue_Values()
}

type Value_RandomString struct {
	RandomString *RandomStringValue `protobuf:"bytes,1,opt,name=randomString,proto3,oneof"`
}

type Value_VocabString struct {
	VocabString *VocabStringValue `protobuf:"bytes,2,opt,name=vocabString,proto3,oneof"`
}

type Value_FixedString struct {
	FixedString *FixedStringValue `protobuf:"bytes,3,opt,name=fixedString,proto3,oneof"`
}

type Value_RandomNumeric struct {
	RandomNumeric *RandomNumericValue `protobuf:"bytes,4,opt,name=randomNumeric,proto3,oneof"`
}

type Value_NumericSet struct {
	NumericSet *NumericSetValue `protobuf:"bytes,5,opt,name=numericSet,proto3,oneof"`
}

type Value_FixedNumeric struct {
	FixedNumeric *FixedNumericValue `protobuf:"bytes,6,opt,name=fixedNumeric,proto3,oneof"`
}

type Value_Boolean struct {
	Boolean *BooleanValue `protobuf:"bytes,7,opt,name=boolean,proto3,oneof"`
}

type Value_Dict struct {
	Dict *DictValue `protobuf:"bytes,8,opt,name=dict,proto3,oneof"`
}

type Value_List struct {
	List *ListValue `protobuf:"bytes,9,opt,name=list,proto3,oneof"`
}

type Value_Reference struct {
	Reference *ReferenceValue `protobuf:"bytes,10,opt,name=reference,proto3,oneof"`
}

func (*Value_RandomString) isValue_Values() {}

func (*Value_VocabString) isValue_Values() {}

func (*Value_FixedString) isValue_Values() {}

func (*Value_RandomNumeric) isValue_Values() {}

func (*Value_NumericSet) isValue_Values() {}

func (*Value_FixedNumeric) isValue_Values() {}

func (*Value_Boolean) isValue_Values() {}

func (*Value_Dict) isValue_Values() {}

func (*Value_List) isValue_Values() {}

func (*Value_Reference) isValue_Values() {}

type Schema struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Name string     `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	Root *DictValue `protobuf:"bytes,2,opt,name=root,proto3" json:"root,omitempty"`
}

func (x *Schema) Reset() {
	*x = Schema{}
	if protoimpl.UnsafeEnabled {
		mi := &file_genevents_proto_msgTypes[11]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Schema) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Schema) ProtoMessage() {}

func (x *Schema) ProtoReflect() protoreflect.Message {
	mi := &file_genevents_proto_msgTypes[11]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Schema.ProtoReflect.Descriptor instead.
func (*Schema) Descriptor() ([]byte, []int) {
	return file_genevents_proto_rawDescGZIP(), []int{11}
}

func (x *Schema) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *Schema) GetRoot() *DictValue {
	if x != nil {
		return x.Root
	}
	return nil
}

type Schemas struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Schemas            []*Schema `protobuf:"bytes,1,rep,name=schemas,proto3" json:"schemas,omitempty"`
	SchemaDistribution []float64 `protobuf:"fixed64,2,rep,packed,name=schemaDistribution,proto3" json:"schemaDistribution,omitempty"`
}

func (x *Schemas) Reset() {
	*x = Schemas{}
	if protoimpl.UnsafeEnabled {
		mi := &file_genevents_proto_msgTypes[12]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Schemas) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Schemas) ProtoMessage() {}

func (x *Schemas) ProtoReflect() protoreflect.Message {
	mi := &file_genevents_proto_msgTypes[12]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Schemas.ProtoReflect.Descriptor instead.
func (*Schemas) Descriptor() ([]byte, []int) {
	return file_genevents_proto_rawDescGZIP(), []int{12}
}

func (x *Schemas) GetSchemas() []*Schema {
	if x != nil {
		return x.Schemas
	}
	return nil
}

func (x *Schemas) GetSchemaDistribution() []float64 {
	if x != nil {
		return x.SchemaDistribution
	}
	return nil
}

var File_genevents_proto protoreflect.FileDescriptor

var file_genevents_proto_rawDesc = []byte{
	0x0a, 0x0f, 0x67, 0x65, 0x6e, 0x65, 0x76, 0x65, 0x6e, 0x74, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x12, 0x09, 0x67, 0x65, 0x6e, 0x65, 0x76, 0x65, 0x6e, 0x74, 0x73, 0x22, 0x77, 0x0a, 0x11,
	0x52, 0x61, 0x6e, 0x64, 0x6f, 0x6d, 0x53, 0x74, 0x72, 0x69, 0x6e, 0x67, 0x56, 0x61, 0x6c, 0x75,
	0x65, 0x12, 0x16, 0x0a, 0x06, 0x6d, 0x61, 0x78, 0x4c, 0x65, 0x6e, 0x18, 0x01, 0x20, 0x01, 0x28,
	0x05, 0x52, 0x06, 0x6d, 0x61, 0x78, 0x4c, 0x65, 0x6e, 0x12, 0x16, 0x0a, 0x06, 0x6d, 0x69, 0x6e,
	0x4c, 0x65, 0x6e, 0x18, 0x02, 0x20, 0x01, 0x28, 0x05, 0x52, 0x06, 0x6d, 0x69, 0x6e, 0x4c, 0x65,
	0x6e, 0x12, 0x1a, 0x0a, 0x08, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x62, 0x65, 0x74, 0x18, 0x03, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x08, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x62, 0x65, 0x74, 0x12, 0x16, 0x0a,
	0x06, 0x6d, 0x61, 0x78, 0x52, 0x65, 0x66, 0x18, 0x04, 0x20, 0x01, 0x28, 0x05, 0x52, 0x06, 0x6d,
	0x61, 0x78, 0x52, 0x65, 0x66, 0x22, 0x40, 0x0a, 0x10, 0x56, 0x6f, 0x63, 0x61, 0x62, 0x53, 0x74,
	0x72, 0x69, 0x6e, 0x67, 0x56, 0x61, 0x6c, 0x75, 0x65, 0x12, 0x14, 0x0a, 0x05, 0x76, 0x6f, 0x63,
	0x61, 0x62, 0x18, 0x01, 0x20, 0x03, 0x28, 0x09, 0x52, 0x05, 0x76, 0x6f, 0x63, 0x61, 0x62, 0x12,
	0x16, 0x0a, 0x06, 0x6d, 0x61, 0x78, 0x52, 0x65, 0x66, 0x18, 0x04, 0x20, 0x01, 0x28, 0x05, 0x52,
	0x06, 0x6d, 0x61, 0x78, 0x52, 0x65, 0x66, 0x22, 0x28, 0x0a, 0x10, 0x46, 0x69, 0x78, 0x65, 0x64,
	0x53, 0x74, 0x72, 0x69, 0x6e, 0x67, 0x56, 0x61, 0x6c, 0x75, 0x65, 0x12, 0x14, 0x0a, 0x05, 0x76,
	0x61, 0x6c, 0x75, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x76, 0x61, 0x6c, 0x75,
	0x65, 0x22, 0x50, 0x0a, 0x12, 0x52, 0x61, 0x6e, 0x64, 0x6f, 0x6d, 0x4e, 0x75, 0x6d, 0x65, 0x72,
	0x69, 0x63, 0x56, 0x61, 0x6c, 0x75, 0x65, 0x12, 0x10, 0x0a, 0x03, 0x6d, 0x69, 0x6e, 0x18, 0x01,
	0x20, 0x01, 0x28, 0x01, 0x52, 0x03, 0x6d, 0x69, 0x6e, 0x12, 0x10, 0x0a, 0x03, 0x6d, 0x61, 0x78,
	0x18, 0x02, 0x20, 0x01, 0x28, 0x01, 0x52, 0x03, 0x6d, 0x61, 0x78, 0x12, 0x16, 0x0a, 0x06, 0x6d,
	0x61, 0x78, 0x52, 0x65, 0x66, 0x18, 0x04, 0x20, 0x01, 0x28, 0x05, 0x52, 0x06, 0x6d, 0x61, 0x78,
	0x52, 0x65, 0x66, 0x22, 0x29, 0x0a, 0x0f, 0x4e, 0x75, 0x6d, 0x65, 0x72, 0x69, 0x63, 0x53, 0x65,
	0x74, 0x56, 0x61, 0x6c, 0x75, 0x65, 0x12, 0x16, 0x0a, 0x06, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x73,
	0x18, 0x01, 0x20, 0x03, 0x28, 0x01, 0x52, 0x06, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x73, 0x22, 0x29,
	0x0a, 0x11, 0x46, 0x69, 0x78, 0x65, 0x64, 0x4e, 0x75, 0x6d, 0x65, 0x72, 0x69, 0x63, 0x56, 0x61,
	0x6c, 0x75, 0x65, 0x12, 0x14, 0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x18, 0x01, 0x20, 0x01,
	0x28, 0x01, 0x52, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x22, 0x0e, 0x0a, 0x0c, 0x42, 0x6f, 0x6f,
	0x6c, 0x65, 0x61, 0x6e, 0x56, 0x61, 0x6c, 0x75, 0x65, 0x22, 0x86, 0x01, 0x0a, 0x09, 0x44, 0x69,
	0x63, 0x74, 0x56, 0x61, 0x6c, 0x75, 0x65, 0x12, 0x2f, 0x0a, 0x03, 0x6b, 0x76, 0x73, 0x18, 0x01,
	0x20, 0x03, 0x28, 0x0b, 0x32, 0x1d, 0x2e, 0x67, 0x65, 0x6e, 0x65, 0x76, 0x65, 0x6e, 0x74, 0x73,
	0x2e, 0x44, 0x69, 0x63, 0x74, 0x56, 0x61, 0x6c, 0x75, 0x65, 0x2e, 0x4b, 0x76, 0x73, 0x45, 0x6e,
	0x74, 0x72, 0x79, 0x52, 0x03, 0x6b, 0x76, 0x73, 0x1a, 0x48, 0x0a, 0x08, 0x4b, 0x76, 0x73, 0x45,
	0x6e, 0x74, 0x72, 0x79, 0x12, 0x10, 0x0a, 0x03, 0x6b, 0x65, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x03, 0x6b, 0x65, 0x79, 0x12, 0x26, 0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x18,
	0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x10, 0x2e, 0x67, 0x65, 0x6e, 0x65, 0x76, 0x65, 0x6e, 0x74,
	0x73, 0x2e, 0x56, 0x61, 0x6c, 0x75, 0x65, 0x52, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x3a, 0x02,
	0x38, 0x01, 0x22, 0x63, 0x0a, 0x09, 0x4c, 0x69, 0x73, 0x74, 0x56, 0x61, 0x6c, 0x75, 0x65, 0x12,
	0x16, 0x0a, 0x06, 0x6d, 0x61, 0x78, 0x4c, 0x65, 0x6e, 0x18, 0x01, 0x20, 0x01, 0x28, 0x05, 0x52,
	0x06, 0x6d, 0x61, 0x78, 0x4c, 0x65, 0x6e, 0x12, 0x16, 0x0a, 0x06, 0x6d, 0x69, 0x6e, 0x4c, 0x65,
	0x6e, 0x18, 0x02, 0x20, 0x01, 0x28, 0x05, 0x52, 0x06, 0x6d, 0x69, 0x6e, 0x4c, 0x65, 0x6e, 0x12,
	0x26, 0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x10,
	0x2e, 0x67, 0x65, 0x6e, 0x65, 0x76, 0x65, 0x6e, 0x74, 0x73, 0x2e, 0x56, 0x61, 0x6c, 0x75, 0x65,
	0x52, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x22, 0x44, 0x0a, 0x0e, 0x52, 0x65, 0x66, 0x65, 0x72,
	0x65, 0x6e, 0x63, 0x65, 0x56, 0x61, 0x6c, 0x75, 0x65, 0x12, 0x1e, 0x0a, 0x0a, 0x73, 0x63, 0x68,
	0x65, 0x6d, 0x61, 0x4e, 0x61, 0x6d, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0a, 0x73,
	0x63, 0x68, 0x65, 0x6d, 0x61, 0x4e, 0x61, 0x6d, 0x65, 0x12, 0x12, 0x0a, 0x04, 0x70, 0x61, 0x74,
	0x68, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x70, 0x61, 0x74, 0x68, 0x22, 0xe8, 0x04,
	0x0a, 0x05, 0x56, 0x61, 0x6c, 0x75, 0x65, 0x12, 0x42, 0x0a, 0x0c, 0x72, 0x61, 0x6e, 0x64, 0x6f,
	0x6d, 0x53, 0x74, 0x72, 0x69, 0x6e, 0x67, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1c, 0x2e,
	0x67, 0x65, 0x6e, 0x65, 0x76, 0x65, 0x6e, 0x74, 0x73, 0x2e, 0x52, 0x61, 0x6e, 0x64, 0x6f, 0x6d,
	0x53, 0x74, 0x72, 0x69, 0x6e, 0x67, 0x56, 0x61, 0x6c, 0x75, 0x65, 0x48, 0x00, 0x52, 0x0c, 0x72,
	0x61, 0x6e, 0x64, 0x6f, 0x6d, 0x53, 0x74, 0x72, 0x69, 0x6e, 0x67, 0x12, 0x3f, 0x0a, 0x0b, 0x76,
	0x6f, 0x63, 0x61, 0x62, 0x53, 0x74, 0x72, 0x69, 0x6e, 0x67, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b,
	0x32, 0x1b, 0x2e, 0x67, 0x65, 0x6e, 0x65, 0x76, 0x65, 0x6e, 0x74, 0x73, 0x2e, 0x56, 0x6f, 0x63,
	0x61, 0x62, 0x53, 0x74, 0x72, 0x69, 0x6e, 0x67, 0x56, 0x61, 0x6c, 0x75, 0x65, 0x48, 0x00, 0x52,
	0x0b, 0x76, 0x6f, 0x63, 0x61, 0x62, 0x53, 0x74, 0x72, 0x69, 0x6e, 0x67, 0x12, 0x3f, 0x0a, 0x0b,
	0x66, 0x69, 0x78, 0x65, 0x64, 0x53, 0x74, 0x72, 0x69, 0x6e, 0x67, 0x18, 0x03, 0x20, 0x01, 0x28,
	0x0b, 0x32, 0x1b, 0x2e, 0x67, 0x65, 0x6e, 0x65, 0x76, 0x65, 0x6e, 0x74, 0x73, 0x2e, 0x46, 0x69,
	0x78, 0x65, 0x64, 0x53, 0x74, 0x72, 0x69, 0x6e, 0x67, 0x56, 0x61, 0x6c, 0x75, 0x65, 0x48, 0x00,
	0x52, 0x0b, 0x66, 0x69, 0x78, 0x65, 0x64, 0x53, 0x74, 0x72, 0x69, 0x6e, 0x67, 0x12, 0x45, 0x0a,
	0x0d, 0x72, 0x61, 0x6e, 0x64, 0x6f, 0x6d, 0x4e, 0x75, 0x6d, 0x65, 0x72, 0x69, 0x63, 0x18, 0x04,
	0x20, 0x01, 0x28, 0x0b, 0x32, 0x1d, 0x2e, 0x67, 0x65, 0x6e, 0x65, 0x76, 0x65, 0x6e, 0x74, 0x73,
	0x2e, 0x52, 0x61, 0x6e, 0x64, 0x6f, 0x6d, 0x4e, 0x75, 0x6d, 0x65, 0x72, 0x69, 0x63, 0x56, 0x61,
	0x6c, 0x75, 0x65, 0x48, 0x00, 0x52, 0x0d, 0x72, 0x61, 0x6e, 0x64, 0x6f, 0x6d, 0x4e, 0x75, 0x6d,
	0x65, 0x72, 0x69, 0x63, 0x12, 0x3c, 0x0a, 0x0a, 0x6e, 0x75, 0x6d, 0x65, 0x72, 0x69, 0x63, 0x53,
	0x65, 0x74, 0x18, 0x05, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1a, 0x2e, 0x67, 0x65, 0x6e, 0x65, 0x76,
	0x65, 0x6e, 0x74, 0x73, 0x2e, 0x4e, 0x75, 0x6d, 0x65, 0x72, 0x69, 0x63, 0x53, 0x65, 0x74, 0x56,
	0x61, 0x6c, 0x75, 0x65, 0x48, 0x00, 0x52, 0x0a, 0x6e, 0x75, 0x6d, 0x65, 0x72, 0x69, 0x63, 0x53,
	0x65, 0x74, 0x12, 0x42, 0x0a, 0x0c, 0x66, 0x69, 0x78, 0x65, 0x64, 0x4e, 0x75, 0x6d, 0x65, 0x72,
	0x69, 0x63, 0x18, 0x06, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1c, 0x2e, 0x67, 0x65, 0x6e, 0x65, 0x76,
	0x65, 0x6e, 0x74, 0x73, 0x2e, 0x46, 0x69, 0x78, 0x65, 0x64, 0x4e, 0x75, 0x6d, 0x65, 0x72, 0x69,
	0x63, 0x56, 0x61, 0x6c, 0x75, 0x65, 0x48, 0x00, 0x52, 0x0c, 0x66, 0x69, 0x78, 0x65, 0x64, 0x4e,
	0x75, 0x6d, 0x65, 0x72, 0x69, 0x63, 0x12, 0x33, 0x0a, 0x07, 0x62, 0x6f, 0x6f, 0x6c, 0x65, 0x61,
	0x6e, 0x18, 0x07, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x17, 0x2e, 0x67, 0x65, 0x6e, 0x65, 0x76, 0x65,
	0x6e, 0x74, 0x73, 0x2e, 0x42, 0x6f, 0x6f, 0x6c, 0x65, 0x61, 0x6e, 0x56, 0x61, 0x6c, 0x75, 0x65,
	0x48, 0x00, 0x52, 0x07, 0x62, 0x6f, 0x6f, 0x6c, 0x65, 0x61, 0x6e, 0x12, 0x2a, 0x0a, 0x04, 0x64,
	0x69, 0x63, 0x74, 0x18, 0x08, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x14, 0x2e, 0x67, 0x65, 0x6e, 0x65,
	0x76, 0x65, 0x6e, 0x74, 0x73, 0x2e, 0x44, 0x69, 0x63, 0x74, 0x56, 0x61, 0x6c, 0x75, 0x65, 0x48,
	0x00, 0x52, 0x04, 0x64, 0x69, 0x63, 0x74, 0x12, 0x2a, 0x0a, 0x04, 0x6c, 0x69, 0x73, 0x74, 0x18,
	0x09, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x14, 0x2e, 0x67, 0x65, 0x6e, 0x65, 0x76, 0x65, 0x6e, 0x74,
	0x73, 0x2e, 0x4c, 0x69, 0x73, 0x74, 0x56, 0x61, 0x6c, 0x75, 0x65, 0x48, 0x00, 0x52, 0x04, 0x6c,
	0x69, 0x73, 0x74, 0x12, 0x39, 0x0a, 0x09, 0x72, 0x65, 0x66, 0x65, 0x72, 0x65, 0x6e, 0x63, 0x65,
	0x18, 0x0a, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x19, 0x2e, 0x67, 0x65, 0x6e, 0x65, 0x76, 0x65, 0x6e,
	0x74, 0x73, 0x2e, 0x52, 0x65, 0x66, 0x65, 0x72, 0x65, 0x6e, 0x63, 0x65, 0x56, 0x61, 0x6c, 0x75,
	0x65, 0x48, 0x00, 0x52, 0x09, 0x72, 0x65, 0x66, 0x65, 0x72, 0x65, 0x6e, 0x63, 0x65, 0x42, 0x08,
	0x0a, 0x06, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x73, 0x22, 0x46, 0x0a, 0x06, 0x53, 0x63, 0x68, 0x65,
	0x6d, 0x61, 0x12, 0x12, 0x0a, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x12, 0x28, 0x0a, 0x04, 0x72, 0x6f, 0x6f, 0x74, 0x18, 0x02,
	0x20, 0x01, 0x28, 0x0b, 0x32, 0x14, 0x2e, 0x67, 0x65, 0x6e, 0x65, 0x76, 0x65, 0x6e, 0x74, 0x73,
	0x2e, 0x44, 0x69, 0x63, 0x74, 0x56, 0x61, 0x6c, 0x75, 0x65, 0x52, 0x04, 0x72, 0x6f, 0x6f, 0x74,
	0x22, 0x66, 0x0a, 0x07, 0x53, 0x63, 0x68, 0x65, 0x6d, 0x61, 0x73, 0x12, 0x2b, 0x0a, 0x07, 0x73,
	0x63, 0x68, 0x65, 0x6d, 0x61, 0x73, 0x18, 0x01, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x11, 0x2e, 0x67,
	0x65, 0x6e, 0x65, 0x76, 0x65, 0x6e, 0x74, 0x73, 0x2e, 0x53, 0x63, 0x68, 0x65, 0x6d, 0x61, 0x52,
	0x07, 0x73, 0x63, 0x68, 0x65, 0x6d, 0x61, 0x73, 0x12, 0x2e, 0x0a, 0x12, 0x73, 0x63, 0x68, 0x65,
	0x6d, 0x61, 0x44, 0x69, 0x73, 0x74, 0x72, 0x69, 0x62, 0x75, 0x74, 0x69, 0x6f, 0x6e, 0x18, 0x02,
	0x20, 0x03, 0x28, 0x01, 0x52, 0x12, 0x73, 0x63, 0x68, 0x65, 0x6d, 0x61, 0x44, 0x69, 0x73, 0x74,
	0x72, 0x69, 0x62, 0x75, 0x74, 0x69, 0x6f, 0x6e, 0x42, 0x07, 0x5a, 0x05, 0x2e, 0x3b, 0x61, 0x70,
	0x69, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_genevents_proto_rawDescOnce sync.Once
	file_genevents_proto_rawDescData = file_genevents_proto_rawDesc
)

func file_genevents_proto_rawDescGZIP() []byte {
	file_genevents_proto_rawDescOnce.Do(func() {
		file_genevents_proto_rawDescData = protoimpl.X.CompressGZIP(file_genevents_proto_rawDescData)
	})
	return file_genevents_proto_rawDescData
}

var file_genevents_proto_msgTypes = make([]protoimpl.MessageInfo, 14)
var file_genevents_proto_goTypes = []interface{}{
	(*RandomStringValue)(nil),  // 0: genevents.RandomStringValue
	(*VocabStringValue)(nil),   // 1: genevents.VocabStringValue
	(*FixedStringValue)(nil),   // 2: genevents.FixedStringValue
	(*RandomNumericValue)(nil), // 3: genevents.RandomNumericValue
	(*NumericSetValue)(nil),    // 4: genevents.NumericSetValue
	(*FixedNumericValue)(nil),  // 5: genevents.FixedNumericValue
	(*BooleanValue)(nil),       // 6: genevents.BooleanValue
	(*DictValue)(nil),          // 7: genevents.DictValue
	(*ListValue)(nil),          // 8: genevents.ListValue
	(*ReferenceValue)(nil),     // 9: genevents.ReferenceValue
	(*Value)(nil),              // 10: genevents.Value
	(*Schema)(nil),             // 11: genevents.Schema
	(*Schemas)(nil),            // 12: genevents.Schemas
	nil,                        // 13: genevents.DictValue.KvsEntry
}
var file_genevents_proto_depIdxs = []int32{
	13, // 0: genevents.DictValue.kvs:type_name -> genevents.DictValue.KvsEntry
	10, // 1: genevents.ListValue.value:type_name -> genevents.Value
	0,  // 2: genevents.Value.randomString:type_name -> genevents.RandomStringValue
	1,  // 3: genevents.Value.vocabString:type_name -> genevents.VocabStringValue
	2,  // 4: genevents.Value.fixedString:type_name -> genevents.FixedStringValue
	3,  // 5: genevents.Value.randomNumeric:type_name -> genevents.RandomNumericValue
	4,  // 6: genevents.Value.numericSet:type_name -> genevents.NumericSetValue
	5,  // 7: genevents.Value.fixedNumeric:type_name -> genevents.FixedNumericValue
	6,  // 8: genevents.Value.boolean:type_name -> genevents.BooleanValue
	7,  // 9: genevents.Value.dict:type_name -> genevents.DictValue
	8,  // 10: genevents.Value.list:type_name -> genevents.ListValue
	9,  // 11: genevents.Value.reference:type_name -> genevents.ReferenceValue
	7,  // 12: genevents.Schema.root:type_name -> genevents.DictValue
	11, // 13: genevents.Schemas.schemas:type_name -> genevents.Schema
	10, // 14: genevents.DictValue.KvsEntry.value:type_name -> genevents.Value
	15, // [15:15] is the sub-list for method output_type
	15, // [15:15] is the sub-list for method input_type
	15, // [15:15] is the sub-list for extension type_name
	15, // [15:15] is the sub-list for extension extendee
	0,  // [0:15] is the sub-list for field type_name
}

func init() { file_genevents_proto_init() }
func file_genevents_proto_init() {
	if File_genevents_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_genevents_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*RandomStringValue); i {
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
		file_genevents_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*VocabStringValue); i {
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
		file_genevents_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*FixedStringValue); i {
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
		file_genevents_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*RandomNumericValue); i {
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
		file_genevents_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*NumericSetValue); i {
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
		file_genevents_proto_msgTypes[5].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*FixedNumericValue); i {
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
		file_genevents_proto_msgTypes[6].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*BooleanValue); i {
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
		file_genevents_proto_msgTypes[7].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*DictValue); i {
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
		file_genevents_proto_msgTypes[8].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ListValue); i {
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
		file_genevents_proto_msgTypes[9].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ReferenceValue); i {
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
		file_genevents_proto_msgTypes[10].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Value); i {
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
		file_genevents_proto_msgTypes[11].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Schema); i {
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
		file_genevents_proto_msgTypes[12].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Schemas); i {
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
	file_genevents_proto_msgTypes[10].OneofWrappers = []interface{}{
		(*Value_RandomString)(nil),
		(*Value_VocabString)(nil),
		(*Value_FixedString)(nil),
		(*Value_RandomNumeric)(nil),
		(*Value_NumericSet)(nil),
		(*Value_FixedNumeric)(nil),
		(*Value_Boolean)(nil),
		(*Value_Dict)(nil),
		(*Value_List)(nil),
		(*Value_Reference)(nil),
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_genevents_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   14,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_genevents_proto_goTypes,
		DependencyIndexes: file_genevents_proto_depIdxs,
		MessageInfos:      file_genevents_proto_msgTypes,
	}.Build()
	File_genevents_proto = out.File
	file_genevents_proto_rawDesc = nil
	file_genevents_proto_goTypes = nil
	file_genevents_proto_depIdxs = nil
}
