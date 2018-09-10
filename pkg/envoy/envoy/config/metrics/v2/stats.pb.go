// Code generated by protoc-gen-go. DO NOT EDIT.
// source: envoy/config/metrics/v2/stats.proto

package v2

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"
import core "github.com/cilium/cilium/pkg/envoy/envoy/api/v2/core"
import _struct "github.com/golang/protobuf/ptypes/struct"
import wrappers "github.com/golang/protobuf/ptypes/wrappers"
import _ "github.com/lyft/protoc-gen-validate/validate"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion2 // please upgrade the proto package

// Configuration for pluggable stats sinks.
type StatsSink struct {
	// The name of the stats sink to instantiate. The name must match a supported
	// stats sink. The built-in stats sinks are:
	//
	// * :ref:`envoy.statsd <envoy_api_msg_config.metrics.v2.StatsdSink>`
	// * :ref:`envoy.dog_statsd <envoy_api_msg_config.metrics.v2.DogStatsdSink>`
	// * :ref:`envoy.metrics_service <envoy_api_msg_config.metrics.v2.MetricsServiceConfig>`
	// * :ref:`envoy.stat_sinks.hystrix <envoy_api_msg_config.metrics.v2.HystrixSink>`
	//
	// Sinks optionally support tagged/multiple dimensional metrics.
	Name string `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	// Stats sink specific configuration which depends on the sink being
	// instantiated. See :ref:`StatsdSink <envoy_api_msg_config.metrics.v2.StatsdSink>` for an
	// example.
	Config               *_struct.Struct `protobuf:"bytes,2,opt,name=config,proto3" json:"config,omitempty"`
	XXX_NoUnkeyedLiteral struct{}        `json:"-"`
	XXX_unrecognized     []byte          `json:"-"`
	XXX_sizecache        int32           `json:"-"`
}

func (m *StatsSink) Reset()         { *m = StatsSink{} }
func (m *StatsSink) String() string { return proto.CompactTextString(m) }
func (*StatsSink) ProtoMessage()    {}
func (*StatsSink) Descriptor() ([]byte, []int) {
	return fileDescriptor_stats_feb95eff0d73590e, []int{0}
}
func (m *StatsSink) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_StatsSink.Unmarshal(m, b)
}
func (m *StatsSink) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_StatsSink.Marshal(b, m, deterministic)
}
func (dst *StatsSink) XXX_Merge(src proto.Message) {
	xxx_messageInfo_StatsSink.Merge(dst, src)
}
func (m *StatsSink) XXX_Size() int {
	return xxx_messageInfo_StatsSink.Size(m)
}
func (m *StatsSink) XXX_DiscardUnknown() {
	xxx_messageInfo_StatsSink.DiscardUnknown(m)
}

var xxx_messageInfo_StatsSink proto.InternalMessageInfo

func (m *StatsSink) GetName() string {
	if m != nil {
		return m.Name
	}
	return ""
}

func (m *StatsSink) GetConfig() *_struct.Struct {
	if m != nil {
		return m.Config
	}
	return nil
}

// Statistics configuration such as tagging.
type StatsConfig struct {
	// Each stat name is iteratively processed through these tag specifiers.
	// When a tag is matched, the first capture group is removed from the name so
	// later :ref:`TagSpecifiers <envoy_api_msg_config.metrics.v2.TagSpecifier>` cannot match that
	// same portion of the match.
	StatsTags []*TagSpecifier `protobuf:"bytes,1,rep,name=stats_tags,json=statsTags,proto3" json:"stats_tags,omitempty"`
	// Use all default tag regexes specified in Envoy. These can be combined with
	// custom tags specified in :ref:`stats_tags
	// <envoy_api_field_config.metrics.v2.StatsConfig.stats_tags>`. They will be processed before
	// the custom tags.
	//
	// .. note::
	//
	//   If any default tags are specified twice, the config will be considered
	//   invalid.
	//
	// See `well_known_names.h
	// <https://github.com/envoyproxy/envoy/blob/master/source/common/config/well_known_names.h>`_
	// for a list of the default tags in Envoy.
	//
	// If not provided, the value is assumed to be true.
	UseAllDefaultTags    *wrappers.BoolValue `protobuf:"bytes,2,opt,name=use_all_default_tags,json=useAllDefaultTags,proto3" json:"use_all_default_tags,omitempty"`
	XXX_NoUnkeyedLiteral struct{}            `json:"-"`
	XXX_unrecognized     []byte              `json:"-"`
	XXX_sizecache        int32               `json:"-"`
}

func (m *StatsConfig) Reset()         { *m = StatsConfig{} }
func (m *StatsConfig) String() string { return proto.CompactTextString(m) }
func (*StatsConfig) ProtoMessage()    {}
func (*StatsConfig) Descriptor() ([]byte, []int) {
	return fileDescriptor_stats_feb95eff0d73590e, []int{1}
}
func (m *StatsConfig) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_StatsConfig.Unmarshal(m, b)
}
func (m *StatsConfig) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_StatsConfig.Marshal(b, m, deterministic)
}
func (dst *StatsConfig) XXX_Merge(src proto.Message) {
	xxx_messageInfo_StatsConfig.Merge(dst, src)
}
func (m *StatsConfig) XXX_Size() int {
	return xxx_messageInfo_StatsConfig.Size(m)
}
func (m *StatsConfig) XXX_DiscardUnknown() {
	xxx_messageInfo_StatsConfig.DiscardUnknown(m)
}

var xxx_messageInfo_StatsConfig proto.InternalMessageInfo

func (m *StatsConfig) GetStatsTags() []*TagSpecifier {
	if m != nil {
		return m.StatsTags
	}
	return nil
}

func (m *StatsConfig) GetUseAllDefaultTags() *wrappers.BoolValue {
	if m != nil {
		return m.UseAllDefaultTags
	}
	return nil
}

// Designates a tag name and value pair. The value may be either a fixed value
// or a regex providing the value via capture groups. The specified tag will be
// unconditionally set if a fixed value, otherwise it will only be set if one
// or more capture groups in the regex match.
type TagSpecifier struct {
	// Attaches an identifier to the tag values to identify the tag being in the
	// sink. Envoy has a set of default names and regexes to extract dynamic
	// portions of existing stats, which can be found in `well_known_names.h
	// <https://github.com/envoyproxy/envoy/blob/master/source/common/config/well_known_names.h>`_
	// in the Envoy repository. If a :ref:`tag_name
	// <envoy_api_field_config.metrics.v2.TagSpecifier.tag_name>` is provided in the config and
	// neither :ref:`regex <envoy_api_field_config.metrics.v2.TagSpecifier.regex>` or
	// :ref:`fixed_value <envoy_api_field_config.metrics.v2.TagSpecifier.fixed_value>` were specified,
	// Envoy will attempt to find that name in its set of defaults and use the accompanying regex.
	//
	// .. note::
	//
	//   It is invalid to specify the same tag name twice in a config.
	TagName string `protobuf:"bytes,1,opt,name=tag_name,json=tagName,proto3" json:"tag_name,omitempty"`
	// Types that are valid to be assigned to TagValue:
	//	*TagSpecifier_Regex
	//	*TagSpecifier_FixedValue
	TagValue             isTagSpecifier_TagValue `protobuf_oneof:"tag_value"`
	XXX_NoUnkeyedLiteral struct{}                `json:"-"`
	XXX_unrecognized     []byte                  `json:"-"`
	XXX_sizecache        int32                   `json:"-"`
}

func (m *TagSpecifier) Reset()         { *m = TagSpecifier{} }
func (m *TagSpecifier) String() string { return proto.CompactTextString(m) }
func (*TagSpecifier) ProtoMessage()    {}
func (*TagSpecifier) Descriptor() ([]byte, []int) {
	return fileDescriptor_stats_feb95eff0d73590e, []int{2}
}
func (m *TagSpecifier) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_TagSpecifier.Unmarshal(m, b)
}
func (m *TagSpecifier) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_TagSpecifier.Marshal(b, m, deterministic)
}
func (dst *TagSpecifier) XXX_Merge(src proto.Message) {
	xxx_messageInfo_TagSpecifier.Merge(dst, src)
}
func (m *TagSpecifier) XXX_Size() int {
	return xxx_messageInfo_TagSpecifier.Size(m)
}
func (m *TagSpecifier) XXX_DiscardUnknown() {
	xxx_messageInfo_TagSpecifier.DiscardUnknown(m)
}

var xxx_messageInfo_TagSpecifier proto.InternalMessageInfo

func (m *TagSpecifier) GetTagName() string {
	if m != nil {
		return m.TagName
	}
	return ""
}

type isTagSpecifier_TagValue interface {
	isTagSpecifier_TagValue()
}

type TagSpecifier_Regex struct {
	Regex string `protobuf:"bytes,2,opt,name=regex,proto3,oneof"`
}

type TagSpecifier_FixedValue struct {
	FixedValue string `protobuf:"bytes,3,opt,name=fixed_value,json=fixedValue,proto3,oneof"`
}

func (*TagSpecifier_Regex) isTagSpecifier_TagValue() {}

func (*TagSpecifier_FixedValue) isTagSpecifier_TagValue() {}

func (m *TagSpecifier) GetTagValue() isTagSpecifier_TagValue {
	if m != nil {
		return m.TagValue
	}
	return nil
}

func (m *TagSpecifier) GetRegex() string {
	if x, ok := m.GetTagValue().(*TagSpecifier_Regex); ok {
		return x.Regex
	}
	return ""
}

func (m *TagSpecifier) GetFixedValue() string {
	if x, ok := m.GetTagValue().(*TagSpecifier_FixedValue); ok {
		return x.FixedValue
	}
	return ""
}

// XXX_OneofFuncs is for the internal use of the proto package.
func (*TagSpecifier) XXX_OneofFuncs() (func(msg proto.Message, b *proto.Buffer) error, func(msg proto.Message, tag, wire int, b *proto.Buffer) (bool, error), func(msg proto.Message) (n int), []interface{}) {
	return _TagSpecifier_OneofMarshaler, _TagSpecifier_OneofUnmarshaler, _TagSpecifier_OneofSizer, []interface{}{
		(*TagSpecifier_Regex)(nil),
		(*TagSpecifier_FixedValue)(nil),
	}
}

func _TagSpecifier_OneofMarshaler(msg proto.Message, b *proto.Buffer) error {
	m := msg.(*TagSpecifier)
	// tag_value
	switch x := m.TagValue.(type) {
	case *TagSpecifier_Regex:
		b.EncodeVarint(2<<3 | proto.WireBytes)
		b.EncodeStringBytes(x.Regex)
	case *TagSpecifier_FixedValue:
		b.EncodeVarint(3<<3 | proto.WireBytes)
		b.EncodeStringBytes(x.FixedValue)
	case nil:
	default:
		return fmt.Errorf("TagSpecifier.TagValue has unexpected type %T", x)
	}
	return nil
}

func _TagSpecifier_OneofUnmarshaler(msg proto.Message, tag, wire int, b *proto.Buffer) (bool, error) {
	m := msg.(*TagSpecifier)
	switch tag {
	case 2: // tag_value.regex
		if wire != proto.WireBytes {
			return true, proto.ErrInternalBadWireType
		}
		x, err := b.DecodeStringBytes()
		m.TagValue = &TagSpecifier_Regex{x}
		return true, err
	case 3: // tag_value.fixed_value
		if wire != proto.WireBytes {
			return true, proto.ErrInternalBadWireType
		}
		x, err := b.DecodeStringBytes()
		m.TagValue = &TagSpecifier_FixedValue{x}
		return true, err
	default:
		return false, nil
	}
}

func _TagSpecifier_OneofSizer(msg proto.Message) (n int) {
	m := msg.(*TagSpecifier)
	// tag_value
	switch x := m.TagValue.(type) {
	case *TagSpecifier_Regex:
		n += 1 // tag and wire
		n += proto.SizeVarint(uint64(len(x.Regex)))
		n += len(x.Regex)
	case *TagSpecifier_FixedValue:
		n += 1 // tag and wire
		n += proto.SizeVarint(uint64(len(x.FixedValue)))
		n += len(x.FixedValue)
	case nil:
	default:
		panic(fmt.Sprintf("proto: unexpected type %T in oneof", x))
	}
	return n
}

// Stats configuration proto schema for built-in *envoy.statsd* sink. This sink does not support
// tagged metrics.
type StatsdSink struct {
	// Types that are valid to be assigned to StatsdSpecifier:
	//	*StatsdSink_Address
	//	*StatsdSink_TcpClusterName
	StatsdSpecifier isStatsdSink_StatsdSpecifier `protobuf_oneof:"statsd_specifier"`
	// Optional custom prefix for StatsdSink. If
	// specified, this will override the default prefix.
	// For example:
	//
	// .. code-block:: json
	//
	//   {
	//     "prefix" : "envoy-prod"
	//   }
	//
	// will change emitted stats to
	//
	// .. code-block:: cpp
	//
	//   envoy-prod.test_counter:1|c
	//   envoy-prod.test_timer:5|ms
	//
	// Note that the default prefix, "envoy", will be used if a prefix is not
	// specified.
	//
	// Stats with default prefix:
	//
	// .. code-block:: cpp
	//
	//   envoy.test_counter:1|c
	//   envoy.test_timer:5|ms
	Prefix               string   `protobuf:"bytes,3,opt,name=prefix,proto3" json:"prefix,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *StatsdSink) Reset()         { *m = StatsdSink{} }
func (m *StatsdSink) String() string { return proto.CompactTextString(m) }
func (*StatsdSink) ProtoMessage()    {}
func (*StatsdSink) Descriptor() ([]byte, []int) {
	return fileDescriptor_stats_feb95eff0d73590e, []int{3}
}
func (m *StatsdSink) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_StatsdSink.Unmarshal(m, b)
}
func (m *StatsdSink) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_StatsdSink.Marshal(b, m, deterministic)
}
func (dst *StatsdSink) XXX_Merge(src proto.Message) {
	xxx_messageInfo_StatsdSink.Merge(dst, src)
}
func (m *StatsdSink) XXX_Size() int {
	return xxx_messageInfo_StatsdSink.Size(m)
}
func (m *StatsdSink) XXX_DiscardUnknown() {
	xxx_messageInfo_StatsdSink.DiscardUnknown(m)
}

var xxx_messageInfo_StatsdSink proto.InternalMessageInfo

type isStatsdSink_StatsdSpecifier interface {
	isStatsdSink_StatsdSpecifier()
}

type StatsdSink_Address struct {
	Address *core.Address `protobuf:"bytes,1,opt,name=address,proto3,oneof"`
}

type StatsdSink_TcpClusterName struct {
	TcpClusterName string `protobuf:"bytes,2,opt,name=tcp_cluster_name,json=tcpClusterName,proto3,oneof"`
}

func (*StatsdSink_Address) isStatsdSink_StatsdSpecifier() {}

func (*StatsdSink_TcpClusterName) isStatsdSink_StatsdSpecifier() {}

func (m *StatsdSink) GetStatsdSpecifier() isStatsdSink_StatsdSpecifier {
	if m != nil {
		return m.StatsdSpecifier
	}
	return nil
}

func (m *StatsdSink) GetAddress() *core.Address {
	if x, ok := m.GetStatsdSpecifier().(*StatsdSink_Address); ok {
		return x.Address
	}
	return nil
}

func (m *StatsdSink) GetTcpClusterName() string {
	if x, ok := m.GetStatsdSpecifier().(*StatsdSink_TcpClusterName); ok {
		return x.TcpClusterName
	}
	return ""
}

func (m *StatsdSink) GetPrefix() string {
	if m != nil {
		return m.Prefix
	}
	return ""
}

// XXX_OneofFuncs is for the internal use of the proto package.
func (*StatsdSink) XXX_OneofFuncs() (func(msg proto.Message, b *proto.Buffer) error, func(msg proto.Message, tag, wire int, b *proto.Buffer) (bool, error), func(msg proto.Message) (n int), []interface{}) {
	return _StatsdSink_OneofMarshaler, _StatsdSink_OneofUnmarshaler, _StatsdSink_OneofSizer, []interface{}{
		(*StatsdSink_Address)(nil),
		(*StatsdSink_TcpClusterName)(nil),
	}
}

func _StatsdSink_OneofMarshaler(msg proto.Message, b *proto.Buffer) error {
	m := msg.(*StatsdSink)
	// statsd_specifier
	switch x := m.StatsdSpecifier.(type) {
	case *StatsdSink_Address:
		b.EncodeVarint(1<<3 | proto.WireBytes)
		if err := b.EncodeMessage(x.Address); err != nil {
			return err
		}
	case *StatsdSink_TcpClusterName:
		b.EncodeVarint(2<<3 | proto.WireBytes)
		b.EncodeStringBytes(x.TcpClusterName)
	case nil:
	default:
		return fmt.Errorf("StatsdSink.StatsdSpecifier has unexpected type %T", x)
	}
	return nil
}

func _StatsdSink_OneofUnmarshaler(msg proto.Message, tag, wire int, b *proto.Buffer) (bool, error) {
	m := msg.(*StatsdSink)
	switch tag {
	case 1: // statsd_specifier.address
		if wire != proto.WireBytes {
			return true, proto.ErrInternalBadWireType
		}
		msg := new(core.Address)
		err := b.DecodeMessage(msg)
		m.StatsdSpecifier = &StatsdSink_Address{msg}
		return true, err
	case 2: // statsd_specifier.tcp_cluster_name
		if wire != proto.WireBytes {
			return true, proto.ErrInternalBadWireType
		}
		x, err := b.DecodeStringBytes()
		m.StatsdSpecifier = &StatsdSink_TcpClusterName{x}
		return true, err
	default:
		return false, nil
	}
}

func _StatsdSink_OneofSizer(msg proto.Message) (n int) {
	m := msg.(*StatsdSink)
	// statsd_specifier
	switch x := m.StatsdSpecifier.(type) {
	case *StatsdSink_Address:
		s := proto.Size(x.Address)
		n += 1 // tag and wire
		n += proto.SizeVarint(uint64(s))
		n += s
	case *StatsdSink_TcpClusterName:
		n += 1 // tag and wire
		n += proto.SizeVarint(uint64(len(x.TcpClusterName)))
		n += len(x.TcpClusterName)
	case nil:
	default:
		panic(fmt.Sprintf("proto: unexpected type %T in oneof", x))
	}
	return n
}

// Stats configuration proto schema for built-in *envoy.dog_statsd* sink.
// The sink emits stats with `DogStatsD <https://docs.datadoghq.com/guides/dogstatsd/>`_
// compatible tags. Tags are configurable via :ref:`StatsConfig
// <envoy_api_msg_config.metrics.v2.StatsConfig>`.
// [#comment:next free field: 3]
type DogStatsdSink struct {
	// Types that are valid to be assigned to DogStatsdSpecifier:
	//	*DogStatsdSink_Address
	DogStatsdSpecifier   isDogStatsdSink_DogStatsdSpecifier `protobuf_oneof:"dog_statsd_specifier"`
	XXX_NoUnkeyedLiteral struct{}                           `json:"-"`
	XXX_unrecognized     []byte                             `json:"-"`
	XXX_sizecache        int32                              `json:"-"`
}

func (m *DogStatsdSink) Reset()         { *m = DogStatsdSink{} }
func (m *DogStatsdSink) String() string { return proto.CompactTextString(m) }
func (*DogStatsdSink) ProtoMessage()    {}
func (*DogStatsdSink) Descriptor() ([]byte, []int) {
	return fileDescriptor_stats_feb95eff0d73590e, []int{4}
}
func (m *DogStatsdSink) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_DogStatsdSink.Unmarshal(m, b)
}
func (m *DogStatsdSink) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_DogStatsdSink.Marshal(b, m, deterministic)
}
func (dst *DogStatsdSink) XXX_Merge(src proto.Message) {
	xxx_messageInfo_DogStatsdSink.Merge(dst, src)
}
func (m *DogStatsdSink) XXX_Size() int {
	return xxx_messageInfo_DogStatsdSink.Size(m)
}
func (m *DogStatsdSink) XXX_DiscardUnknown() {
	xxx_messageInfo_DogStatsdSink.DiscardUnknown(m)
}

var xxx_messageInfo_DogStatsdSink proto.InternalMessageInfo

type isDogStatsdSink_DogStatsdSpecifier interface {
	isDogStatsdSink_DogStatsdSpecifier()
}

type DogStatsdSink_Address struct {
	Address *core.Address `protobuf:"bytes,1,opt,name=address,proto3,oneof"`
}

func (*DogStatsdSink_Address) isDogStatsdSink_DogStatsdSpecifier() {}

func (m *DogStatsdSink) GetDogStatsdSpecifier() isDogStatsdSink_DogStatsdSpecifier {
	if m != nil {
		return m.DogStatsdSpecifier
	}
	return nil
}

func (m *DogStatsdSink) GetAddress() *core.Address {
	if x, ok := m.GetDogStatsdSpecifier().(*DogStatsdSink_Address); ok {
		return x.Address
	}
	return nil
}

// XXX_OneofFuncs is for the internal use of the proto package.
func (*DogStatsdSink) XXX_OneofFuncs() (func(msg proto.Message, b *proto.Buffer) error, func(msg proto.Message, tag, wire int, b *proto.Buffer) (bool, error), func(msg proto.Message) (n int), []interface{}) {
	return _DogStatsdSink_OneofMarshaler, _DogStatsdSink_OneofUnmarshaler, _DogStatsdSink_OneofSizer, []interface{}{
		(*DogStatsdSink_Address)(nil),
	}
}

func _DogStatsdSink_OneofMarshaler(msg proto.Message, b *proto.Buffer) error {
	m := msg.(*DogStatsdSink)
	// dog_statsd_specifier
	switch x := m.DogStatsdSpecifier.(type) {
	case *DogStatsdSink_Address:
		b.EncodeVarint(1<<3 | proto.WireBytes)
		if err := b.EncodeMessage(x.Address); err != nil {
			return err
		}
	case nil:
	default:
		return fmt.Errorf("DogStatsdSink.DogStatsdSpecifier has unexpected type %T", x)
	}
	return nil
}

func _DogStatsdSink_OneofUnmarshaler(msg proto.Message, tag, wire int, b *proto.Buffer) (bool, error) {
	m := msg.(*DogStatsdSink)
	switch tag {
	case 1: // dog_statsd_specifier.address
		if wire != proto.WireBytes {
			return true, proto.ErrInternalBadWireType
		}
		msg := new(core.Address)
		err := b.DecodeMessage(msg)
		m.DogStatsdSpecifier = &DogStatsdSink_Address{msg}
		return true, err
	default:
		return false, nil
	}
}

func _DogStatsdSink_OneofSizer(msg proto.Message) (n int) {
	m := msg.(*DogStatsdSink)
	// dog_statsd_specifier
	switch x := m.DogStatsdSpecifier.(type) {
	case *DogStatsdSink_Address:
		s := proto.Size(x.Address)
		n += 1 // tag and wire
		n += proto.SizeVarint(uint64(s))
		n += s
	case nil:
	default:
		panic(fmt.Sprintf("proto: unexpected type %T in oneof", x))
	}
	return n
}

// Stats configuration proto schema for built-in *envoy.stat_sinks.hystrix* sink.
// The sink emits stats in `text/event-stream
// <https://developer.mozilla.org/en-US/docs/Web/API/Server-sent_events/Using_server-sent_events>`_
// formatted stream for use by `Hystrix dashboard
// <https://github.com/Netflix-Skunkworks/hystrix-dashboard/wiki>`_.
//
// Note that only a single HystrixSink should be configured.
//
// Streaming is started through an admin endpoint :http:get:`/hystrix_event_stream`.
type HystrixSink struct {
	// The number of buckets the rolling statistical window is divided into.
	//
	// Each time the sink is flushed, all relevant Envoy statistics are sampled and
	// added to the rolling window (removing the oldest samples in the window
	// in the process). The sink then outputs the aggregate statistics across the
	// current rolling window to the event stream(s).
	//
	// rolling_window(ms) = stats_flush_interval(ms) * num_of_buckets
	//
	// More detailed explanation can be found in `Hystix wiki
	// <https://github.com/Netflix/Hystrix/wiki/Metrics-and-Monitoring#hystrixrollingnumber>`_.
	NumBuckets           int64    `protobuf:"varint,1,opt,name=num_buckets,json=numBuckets,proto3" json:"num_buckets,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *HystrixSink) Reset()         { *m = HystrixSink{} }
func (m *HystrixSink) String() string { return proto.CompactTextString(m) }
func (*HystrixSink) ProtoMessage()    {}
func (*HystrixSink) Descriptor() ([]byte, []int) {
	return fileDescriptor_stats_feb95eff0d73590e, []int{5}
}
func (m *HystrixSink) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_HystrixSink.Unmarshal(m, b)
}
func (m *HystrixSink) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_HystrixSink.Marshal(b, m, deterministic)
}
func (dst *HystrixSink) XXX_Merge(src proto.Message) {
	xxx_messageInfo_HystrixSink.Merge(dst, src)
}
func (m *HystrixSink) XXX_Size() int {
	return xxx_messageInfo_HystrixSink.Size(m)
}
func (m *HystrixSink) XXX_DiscardUnknown() {
	xxx_messageInfo_HystrixSink.DiscardUnknown(m)
}

var xxx_messageInfo_HystrixSink proto.InternalMessageInfo

func (m *HystrixSink) GetNumBuckets() int64 {
	if m != nil {
		return m.NumBuckets
	}
	return 0
}

func init() {
	proto.RegisterType((*StatsSink)(nil), "envoy.config.metrics.v2.StatsSink")
	proto.RegisterType((*StatsConfig)(nil), "envoy.config.metrics.v2.StatsConfig")
	proto.RegisterType((*TagSpecifier)(nil), "envoy.config.metrics.v2.TagSpecifier")
	proto.RegisterType((*StatsdSink)(nil), "envoy.config.metrics.v2.StatsdSink")
	proto.RegisterType((*DogStatsdSink)(nil), "envoy.config.metrics.v2.DogStatsdSink")
	proto.RegisterType((*HystrixSink)(nil), "envoy.config.metrics.v2.HystrixSink")
}

func init() {
	proto.RegisterFile("envoy/config/metrics/v2/stats.proto", fileDescriptor_stats_feb95eff0d73590e)
}

var fileDescriptor_stats_feb95eff0d73590e = []byte{
	// 516 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xa4, 0x52, 0x41, 0x8b, 0xd3, 0x40,
	0x14, 0xde, 0x34, 0x6d, 0xb7, 0x7d, 0x51, 0xa9, 0x61, 0xb1, 0xdd, 0xa2, 0x6e, 0x8d, 0x08, 0xc5,
	0xc3, 0x04, 0x22, 0x78, 0xdf, 0x6c, 0x0f, 0x45, 0x41, 0x24, 0x5d, 0x3c, 0x78, 0x09, 0xd3, 0x64,
	0x32, 0x0c, 0x3b, 0xcd, 0x84, 0x99, 0x49, 0xec, 0x1e, 0x04, 0x7f, 0x8a, 0xe0, 0x9f, 0x10, 0x4f,
	0xfb, 0x77, 0xf6, 0x5f, 0x48, 0x66, 0x52, 0x59, 0x94, 0x3d, 0x79, 0x9b, 0x37, 0xef, 0xfb, 0xde,
	0xf7, 0xbd, 0x6f, 0x06, 0x5e, 0x92, 0xb2, 0x11, 0xd7, 0x61, 0x26, 0xca, 0x82, 0xd1, 0x70, 0x47,
	0xb4, 0x64, 0x99, 0x0a, 0x9b, 0x28, 0x54, 0x1a, 0x6b, 0x85, 0x2a, 0x29, 0xb4, 0xf0, 0xa7, 0x06,
	0x84, 0x2c, 0x08, 0x75, 0x20, 0xd4, 0x44, 0xf3, 0x33, 0xcb, 0xc6, 0x15, 0x6b, 0x29, 0x99, 0x90,
	0x24, 0xc4, 0x79, 0x2e, 0x89, 0xea, 0x98, 0xf3, 0xa7, 0x54, 0x08, 0xca, 0x49, 0x68, 0xaa, 0x6d,
	0x5d, 0x84, 0x4a, 0xcb, 0x3a, 0xd3, 0x5d, 0xf7, 0xf9, 0xdf, 0xdd, 0x2f, 0x12, 0x57, 0x15, 0x91,
	0x07, 0xf6, 0xb4, 0xc1, 0x9c, 0xe5, 0x58, 0x93, 0xf0, 0x70, 0xb0, 0x8d, 0xe0, 0x23, 0x8c, 0x37,
	0xad, 0xbf, 0x0d, 0x2b, 0xaf, 0x7c, 0x1f, 0xfa, 0x25, 0xde, 0x91, 0x99, 0xb3, 0x70, 0x96, 0xe3,
	0xc4, 0x9c, 0xfd, 0x10, 0x86, 0xd6, 0xed, 0xac, 0xb7, 0x70, 0x96, 0x5e, 0x34, 0x45, 0x56, 0x0a,
	0x1d, 0xa4, 0xd0, 0xc6, 0x18, 0x49, 0x3a, 0x58, 0xf0, 0xdd, 0x01, 0xcf, 0x8c, 0xbc, 0x30, 0xb5,
	0xbf, 0x02, 0x30, 0x09, 0xa4, 0x1a, 0x53, 0x35, 0x73, 0x16, 0xee, 0xd2, 0x8b, 0x5e, 0xa1, 0x7b,
	0x72, 0x40, 0x97, 0x98, 0x6e, 0x2a, 0x92, 0xb1, 0x82, 0x11, 0x99, 0x8c, 0x0d, 0xf1, 0x12, 0x53,
	0xe5, 0xbf, 0x87, 0x93, 0x5a, 0x91, 0x14, 0x73, 0x9e, 0xe6, 0xa4, 0xc0, 0x35, 0xd7, 0x76, 0x9e,
	0x35, 0x35, 0xff, 0xc7, 0x54, 0x2c, 0x04, 0xff, 0x84, 0x79, 0x4d, 0x92, 0xc7, 0xb5, 0x22, 0xe7,
	0x9c, 0xaf, 0x2c, 0xab, 0x1d, 0x16, 0x7c, 0x85, 0x07, 0x77, 0x75, 0xfc, 0x53, 0x18, 0x69, 0x4c,
	0xd3, 0x3b, 0xbb, 0x1f, 0x6b, 0x4c, 0x3f, 0xb4, 0xeb, 0x07, 0x30, 0x90, 0x84, 0x92, 0xbd, 0x11,
	0x1a, 0xc7, 0xf0, 0xeb, 0xf6, 0xc6, 0x1d, 0x48, 0x77, 0xf9, 0x6d, 0xb4, 0x3e, 0x4a, 0x6c, 0xcb,
	0x7f, 0x01, 0x5e, 0xc1, 0xf6, 0x24, 0x4f, 0x9b, 0x56, 0x70, 0xe6, 0xb6, 0xc8, 0xf5, 0x51, 0x02,
	0xe6, 0xd2, 0x98, 0x88, 0x3d, 0x18, 0xb7, 0x0a, 0x06, 0x10, 0xfc, 0x70, 0x00, 0x4c, 0x42, 0xb9,
	0x49, 0xfd, 0x2d, 0x1c, 0x77, 0x4f, 0x6d, 0xc4, 0xdb, 0x6d, 0x6c, 0x3a, 0xb8, 0x62, 0x6d, 0x24,
	0xed, 0x67, 0x40, 0xe7, 0x16, 0xb1, 0x3e, 0x4a, 0x0e, 0x60, 0xff, 0x35, 0x4c, 0x74, 0x56, 0xa5,
	0x19, 0xaf, 0x95, 0x26, 0xd2, 0xba, 0xef, 0x75, 0xda, 0x8f, 0x74, 0x56, 0x5d, 0xd8, 0x86, 0x59,
	0xe3, 0x09, 0x0c, 0x2b, 0x49, 0x0a, 0xb6, 0xb7, 0xee, 0x92, 0xae, 0x8a, 0x4f, 0x61, 0x62, 0x32,
	0xce, 0x53, 0xf5, 0x27, 0x8d, 0xc1, 0xcf, 0xdb, 0x1b, 0xd7, 0x09, 0x38, 0x3c, 0x5c, 0x09, 0xfa,
	0xff, 0x3e, 0xe3, 0x67, 0x70, 0x92, 0x0b, 0x9a, 0xde, 0xa3, 0xf3, 0xae, 0x3f, 0xea, 0x4d, 0xdc,
	0x00, 0x81, 0xb7, 0xbe, 0x56, 0x5a, 0xb2, 0xbd, 0xd1, 0x3a, 0x03, 0xaf, 0xac, 0x77, 0xe9, 0xb6,
	0xce, 0xae, 0x88, 0xb6, 0x7a, 0x6e, 0x02, 0x65, 0xbd, 0x8b, 0xed, 0x4d, 0xdc, 0xff, 0xdc, 0x6b,
	0xa2, 0xed, 0xd0, 0xbc, 0xf7, 0x9b, 0xdf, 0x01, 0x00, 0x00, 0xff, 0xff, 0x2c, 0x9e, 0x10, 0xee,
	0x7c, 0x03, 0x00, 0x00,
}
