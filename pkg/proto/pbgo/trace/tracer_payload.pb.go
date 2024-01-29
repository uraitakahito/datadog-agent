// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.28.1
// 	protoc        v3.6.1
// source: datadog/trace/tracer_payload.proto

package trace

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

// TraceChunk represents a list of spans with the same trace ID. In other words, a chunk of a trace.
type TraceChunk struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// priority specifies sampling priority of the trace.
	// @gotags: json:"priority" msg:"priority"
	Priority int32 `protobuf:"varint,1,opt,name=priority,proto3" json:"priority" msg:"priority"`
	// origin specifies origin product ("lambda", "rum", etc.) of the trace.
	// @gotags: json:"origin" msg:"origin"
	Origin string `protobuf:"bytes,2,opt,name=origin,proto3" json:"origin" msg:"origin"`
	// spans specifies list of containing spans.
	// @gotags: json:"spans" msg:"spans"
	Spans []*Span `protobuf:"bytes,3,rep,name=spans,proto3" json:"spans" msg:"spans"`
	// tags specifies tags common in all `spans`.
	// @gotags: json:"tags" msg:"tags"
	Tags map[string]string `protobuf:"bytes,4,rep,name=tags,proto3" json:"tags" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3" msg:"tags"`
	// droppedTrace specifies whether the trace was dropped by samplers or not.
	// @gotags: json:"dropped_trace" msg:"dropped_trace"
	DroppedTrace bool `protobuf:"varint,5,opt,name=droppedTrace,proto3" json:"dropped_trace" msg:"dropped_trace"`
}

func (x *TraceChunk) Reset() {
	*x = TraceChunk{}
	if protoimpl.UnsafeEnabled {
		mi := &file_datadog_trace_tracer_payload_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *TraceChunk) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*TraceChunk) ProtoMessage() {}

func (x *TraceChunk) ProtoReflect() protoreflect.Message {
	mi := &file_datadog_trace_tracer_payload_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use TraceChunk.ProtoReflect.Descriptor instead.
func (*TraceChunk) Descriptor() ([]byte, []int) {
	return file_datadog_trace_tracer_payload_proto_rawDescGZIP(), []int{0}
}

func (x *TraceChunk) GetPriority() int32 {
	if x != nil {
		return x.Priority
	}
	return 0
}

func (x *TraceChunk) GetOrigin() string {
	if x != nil {
		return x.Origin
	}
	return ""
}

func (x *TraceChunk) GetSpans() []*Span {
	if x != nil {
		return x.Spans
	}
	return nil
}

func (x *TraceChunk) GetTags() map[string]string {
	if x != nil {
		return x.Tags
	}
	return nil
}

func (x *TraceChunk) GetDroppedTrace() bool {
	if x != nil {
		return x.DroppedTrace
	}
	return false
}

// TracerPayload represents a payload the trace agent receives from tracers.
type TracerPayload struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// containerID specifies the ID of the container where the tracer is running on.
	// @gotags: json:"container_id" msg:"container_id"
	ContainerID string `protobuf:"bytes,1,opt,name=containerID,proto3" json:"container_id" msg:"container_id"`
	// languageName specifies language of the tracer.
	// @gotags: json:"language_name" msg:"language_name"
	LanguageName string `protobuf:"bytes,2,opt,name=languageName,proto3" json:"language_name" msg:"language_name"`
	// languageVersion specifies language version of the tracer.
	// @gotags: json:"language_version" msg:"language_version"
	LanguageVersion string `protobuf:"bytes,3,opt,name=languageVersion,proto3" json:"language_version" msg:"language_version"`
	// tracerVersion specifies version of the tracer.
	// @gotags: json:"tracer_version" msg:"tracer_version"
	TracerVersion string `protobuf:"bytes,4,opt,name=tracerVersion,proto3" json:"tracer_version" msg:"tracer_version"`
	// runtimeID specifies V4 UUID representation of a tracer session.
	// @gotags: json:"runtime_id" msg:"runtime_id"
	RuntimeID string `protobuf:"bytes,5,opt,name=runtimeID,proto3" json:"runtime_id" msg:"runtime_id"`
	// chunks specifies list of containing trace chunks.
	// @gotags: json:"chunks" msg:"chunks"
	Chunks []*TraceChunk `protobuf:"bytes,6,rep,name=chunks,proto3" json:"chunks" msg:"chunks"`
	// tags specifies tags common in all `chunks`.
	// @gotags: json:"tags" msg:"tags"
	Tags map[string]string `protobuf:"bytes,7,rep,name=tags,proto3" json:"tags" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3" msg:"tags"`
	// env specifies `env` tag that set with the tracer.
	// @gotags: json:"env" msg:"env"
	Env string `protobuf:"bytes,8,opt,name=env,proto3" json:"env" msg:"env"`
	// hostname specifies hostname of where the tracer is running.
	// @gotags: json:"hostname" msg:"hostname"
	Hostname string `protobuf:"bytes,9,opt,name=hostname,proto3" json:"hostname" msg:"hostname"`
	// version specifies `version` tag that set with the tracer.
	// @gotags: json:"app_version" msg:"app_version"
	AppVersion string `protobuf:"bytes,10,opt,name=appVersion,proto3" json:"app_version" msg:"app_version"`
}

func (x *TracerPayload) Reset() {
	*x = TracerPayload{}
	if protoimpl.UnsafeEnabled {
		mi := &file_datadog_trace_tracer_payload_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *TracerPayload) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*TracerPayload) ProtoMessage() {}

func (x *TracerPayload) ProtoReflect() protoreflect.Message {
	mi := &file_datadog_trace_tracer_payload_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use TracerPayload.ProtoReflect.Descriptor instead.
func (*TracerPayload) Descriptor() ([]byte, []int) {
	return file_datadog_trace_tracer_payload_proto_rawDescGZIP(), []int{1}
}

func (x *TracerPayload) GetContainerID() string {
	if x != nil {
		return x.ContainerID
	}
	return ""
}

func (x *TracerPayload) GetLanguageName() string {
	if x != nil {
		return x.LanguageName
	}
	return ""
}

func (x *TracerPayload) GetLanguageVersion() string {
	if x != nil {
		return x.LanguageVersion
	}
	return ""
}

func (x *TracerPayload) GetTracerVersion() string {
	if x != nil {
		return x.TracerVersion
	}
	return ""
}

func (x *TracerPayload) GetRuntimeID() string {
	if x != nil {
		return x.RuntimeID
	}
	return ""
}

func (x *TracerPayload) GetChunks() []*TraceChunk {
	if x != nil {
		return x.Chunks
	}
	return nil
}

func (x *TracerPayload) GetTags() map[string]string {
	if x != nil {
		return x.Tags
	}
	return nil
}

func (x *TracerPayload) GetEnv() string {
	if x != nil {
		return x.Env
	}
	return ""
}

func (x *TracerPayload) GetHostname() string {
	if x != nil {
		return x.Hostname
	}
	return ""
}

func (x *TracerPayload) GetAppVersion() string {
	if x != nil {
		return x.AppVersion
	}
	return ""
}

var File_datadog_trace_tracer_payload_proto protoreflect.FileDescriptor

var file_datadog_trace_tracer_payload_proto_rawDesc = []byte{
	0x0a, 0x22, 0x64, 0x61, 0x74, 0x61, 0x64, 0x6f, 0x67, 0x2f, 0x74, 0x72, 0x61, 0x63, 0x65, 0x2f,
	0x74, 0x72, 0x61, 0x63, 0x65, 0x72, 0x5f, 0x70, 0x61, 0x79, 0x6c, 0x6f, 0x61, 0x64, 0x2e, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x12, 0x0d, 0x64, 0x61, 0x74, 0x61, 0x64, 0x6f, 0x67, 0x2e, 0x74, 0x72,
	0x61, 0x63, 0x65, 0x1a, 0x18, 0x64, 0x61, 0x74, 0x61, 0x64, 0x6f, 0x67, 0x2f, 0x74, 0x72, 0x61,
	0x63, 0x65, 0x2f, 0x73, 0x70, 0x61, 0x6e, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x81, 0x02,
	0x0a, 0x0a, 0x54, 0x72, 0x61, 0x63, 0x65, 0x43, 0x68, 0x75, 0x6e, 0x6b, 0x12, 0x1a, 0x0a, 0x08,
	0x70, 0x72, 0x69, 0x6f, 0x72, 0x69, 0x74, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x05, 0x52, 0x08,
	0x70, 0x72, 0x69, 0x6f, 0x72, 0x69, 0x74, 0x79, 0x12, 0x16, 0x0a, 0x06, 0x6f, 0x72, 0x69, 0x67,
	0x69, 0x6e, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x6f, 0x72, 0x69, 0x67, 0x69, 0x6e,
	0x12, 0x29, 0x0a, 0x05, 0x73, 0x70, 0x61, 0x6e, 0x73, 0x18, 0x03, 0x20, 0x03, 0x28, 0x0b, 0x32,
	0x13, 0x2e, 0x64, 0x61, 0x74, 0x61, 0x64, 0x6f, 0x67, 0x2e, 0x74, 0x72, 0x61, 0x63, 0x65, 0x2e,
	0x53, 0x70, 0x61, 0x6e, 0x52, 0x05, 0x73, 0x70, 0x61, 0x6e, 0x73, 0x12, 0x37, 0x0a, 0x04, 0x74,
	0x61, 0x67, 0x73, 0x18, 0x04, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x23, 0x2e, 0x64, 0x61, 0x74, 0x61,
	0x64, 0x6f, 0x67, 0x2e, 0x74, 0x72, 0x61, 0x63, 0x65, 0x2e, 0x54, 0x72, 0x61, 0x63, 0x65, 0x43,
	0x68, 0x75, 0x6e, 0x6b, 0x2e, 0x54, 0x61, 0x67, 0x73, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x52, 0x04,
	0x74, 0x61, 0x67, 0x73, 0x12, 0x22, 0x0a, 0x0c, 0x64, 0x72, 0x6f, 0x70, 0x70, 0x65, 0x64, 0x54,
	0x72, 0x61, 0x63, 0x65, 0x18, 0x05, 0x20, 0x01, 0x28, 0x08, 0x52, 0x0c, 0x64, 0x72, 0x6f, 0x70,
	0x70, 0x65, 0x64, 0x54, 0x72, 0x61, 0x63, 0x65, 0x1a, 0x37, 0x0a, 0x09, 0x54, 0x61, 0x67, 0x73,
	0x45, 0x6e, 0x74, 0x72, 0x79, 0x12, 0x10, 0x0a, 0x03, 0x6b, 0x65, 0x79, 0x18, 0x01, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x03, 0x6b, 0x65, 0x79, 0x12, 0x14, 0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65,
	0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x3a, 0x02, 0x38,
	0x01, 0x22, 0xb9, 0x03, 0x0a, 0x0d, 0x54, 0x72, 0x61, 0x63, 0x65, 0x72, 0x50, 0x61, 0x79, 0x6c,
	0x6f, 0x61, 0x64, 0x12, 0x20, 0x0a, 0x0b, 0x63, 0x6f, 0x6e, 0x74, 0x61, 0x69, 0x6e, 0x65, 0x72,
	0x49, 0x44, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0b, 0x63, 0x6f, 0x6e, 0x74, 0x61, 0x69,
	0x6e, 0x65, 0x72, 0x49, 0x44, 0x12, 0x22, 0x0a, 0x0c, 0x6c, 0x61, 0x6e, 0x67, 0x75, 0x61, 0x67,
	0x65, 0x4e, 0x61, 0x6d, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0c, 0x6c, 0x61, 0x6e,
	0x67, 0x75, 0x61, 0x67, 0x65, 0x4e, 0x61, 0x6d, 0x65, 0x12, 0x28, 0x0a, 0x0f, 0x6c, 0x61, 0x6e,
	0x67, 0x75, 0x61, 0x67, 0x65, 0x56, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e, 0x18, 0x03, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x0f, 0x6c, 0x61, 0x6e, 0x67, 0x75, 0x61, 0x67, 0x65, 0x56, 0x65, 0x72, 0x73,
	0x69, 0x6f, 0x6e, 0x12, 0x24, 0x0a, 0x0d, 0x74, 0x72, 0x61, 0x63, 0x65, 0x72, 0x56, 0x65, 0x72,
	0x73, 0x69, 0x6f, 0x6e, 0x18, 0x04, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0d, 0x74, 0x72, 0x61, 0x63,
	0x65, 0x72, 0x56, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e, 0x12, 0x1c, 0x0a, 0x09, 0x72, 0x75, 0x6e,
	0x74, 0x69, 0x6d, 0x65, 0x49, 0x44, 0x18, 0x05, 0x20, 0x01, 0x28, 0x09, 0x52, 0x09, 0x72, 0x75,
	0x6e, 0x74, 0x69, 0x6d, 0x65, 0x49, 0x44, 0x12, 0x31, 0x0a, 0x06, 0x63, 0x68, 0x75, 0x6e, 0x6b,
	0x73, 0x18, 0x06, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x19, 0x2e, 0x64, 0x61, 0x74, 0x61, 0x64, 0x6f,
	0x67, 0x2e, 0x74, 0x72, 0x61, 0x63, 0x65, 0x2e, 0x54, 0x72, 0x61, 0x63, 0x65, 0x43, 0x68, 0x75,
	0x6e, 0x6b, 0x52, 0x06, 0x63, 0x68, 0x75, 0x6e, 0x6b, 0x73, 0x12, 0x3a, 0x0a, 0x04, 0x74, 0x61,
	0x67, 0x73, 0x18, 0x07, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x26, 0x2e, 0x64, 0x61, 0x74, 0x61, 0x64,
	0x6f, 0x67, 0x2e, 0x74, 0x72, 0x61, 0x63, 0x65, 0x2e, 0x54, 0x72, 0x61, 0x63, 0x65, 0x72, 0x50,
	0x61, 0x79, 0x6c, 0x6f, 0x61, 0x64, 0x2e, 0x54, 0x61, 0x67, 0x73, 0x45, 0x6e, 0x74, 0x72, 0x79,
	0x52, 0x04, 0x74, 0x61, 0x67, 0x73, 0x12, 0x10, 0x0a, 0x03, 0x65, 0x6e, 0x76, 0x18, 0x08, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x03, 0x65, 0x6e, 0x76, 0x12, 0x1a, 0x0a, 0x08, 0x68, 0x6f, 0x73, 0x74,
	0x6e, 0x61, 0x6d, 0x65, 0x18, 0x09, 0x20, 0x01, 0x28, 0x09, 0x52, 0x08, 0x68, 0x6f, 0x73, 0x74,
	0x6e, 0x61, 0x6d, 0x65, 0x12, 0x1e, 0x0a, 0x0a, 0x61, 0x70, 0x70, 0x56, 0x65, 0x72, 0x73, 0x69,
	0x6f, 0x6e, 0x18, 0x0a, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0a, 0x61, 0x70, 0x70, 0x56, 0x65, 0x72,
	0x73, 0x69, 0x6f, 0x6e, 0x1a, 0x37, 0x0a, 0x09, 0x54, 0x61, 0x67, 0x73, 0x45, 0x6e, 0x74, 0x72,
	0x79, 0x12, 0x10, 0x0a, 0x03, 0x6b, 0x65, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03,
	0x6b, 0x65, 0x79, 0x12, 0x14, 0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x18, 0x02, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x3a, 0x02, 0x38, 0x01, 0x42, 0x16, 0x5a,
	0x14, 0x70, 0x6b, 0x67, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x70, 0x62, 0x67, 0x6f, 0x2f,
	0x74, 0x72, 0x61, 0x63, 0x65, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_datadog_trace_tracer_payload_proto_rawDescOnce sync.Once
	file_datadog_trace_tracer_payload_proto_rawDescData = file_datadog_trace_tracer_payload_proto_rawDesc
)

func file_datadog_trace_tracer_payload_proto_rawDescGZIP() []byte {
	file_datadog_trace_tracer_payload_proto_rawDescOnce.Do(func() {
		file_datadog_trace_tracer_payload_proto_rawDescData = protoimpl.X.CompressGZIP(file_datadog_trace_tracer_payload_proto_rawDescData)
	})
	return file_datadog_trace_tracer_payload_proto_rawDescData
}

var file_datadog_trace_tracer_payload_proto_msgTypes = make([]protoimpl.MessageInfo, 4)
var file_datadog_trace_tracer_payload_proto_goTypes = []interface{}{
	(*TraceChunk)(nil),    // 0: datadog.trace.TraceChunk
	(*TracerPayload)(nil), // 1: datadog.trace.TracerPayload
	nil,                   // 2: datadog.trace.TraceChunk.TagsEntry
	nil,                   // 3: datadog.trace.TracerPayload.TagsEntry
	(*Span)(nil),          // 4: datadog.trace.Span
}
var file_datadog_trace_tracer_payload_proto_depIdxs = []int32{
	4, // 0: datadog.trace.TraceChunk.spans:type_name -> datadog.trace.Span
	2, // 1: datadog.trace.TraceChunk.tags:type_name -> datadog.trace.TraceChunk.TagsEntry
	0, // 2: datadog.trace.TracerPayload.chunks:type_name -> datadog.trace.TraceChunk
	3, // 3: datadog.trace.TracerPayload.tags:type_name -> datadog.trace.TracerPayload.TagsEntry
	4, // [4:4] is the sub-list for method output_type
	4, // [4:4] is the sub-list for method input_type
	4, // [4:4] is the sub-list for extension type_name
	4, // [4:4] is the sub-list for extension extendee
	0, // [0:4] is the sub-list for field type_name
}

func init() { file_datadog_trace_tracer_payload_proto_init() }
func file_datadog_trace_tracer_payload_proto_init() {
	if File_datadog_trace_tracer_payload_proto != nil {
		return
	}
	file_datadog_trace_span_proto_init()
	if !protoimpl.UnsafeEnabled {
		file_datadog_trace_tracer_payload_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*TraceChunk); i {
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
		file_datadog_trace_tracer_payload_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*TracerPayload); i {
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
			RawDescriptor: file_datadog_trace_tracer_payload_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   4,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_datadog_trace_tracer_payload_proto_goTypes,
		DependencyIndexes: file_datadog_trace_tracer_payload_proto_depIdxs,
		MessageInfos:      file_datadog_trace_tracer_payload_proto_msgTypes,
	}.Build()
	File_datadog_trace_tracer_payload_proto = out.File
	file_datadog_trace_tracer_payload_proto_rawDesc = nil
	file_datadog_trace_tracer_payload_proto_goTypes = nil
	file_datadog_trace_tracer_payload_proto_depIdxs = nil
}
