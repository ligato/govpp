// Code generated by GoVPP's binapi-generator. DO NOT EDIT.
// versions:
//  binapi-generator: v0.4.0-dev
//  VPP:              20.09-rc0~194-g52e9aaf0b~b1447
// source: /usr/share/vpp/api/core/vxlan.api.json

// Package vxlan contains generated bindings for API file vxlan.api.
//
// Contents:
//   8 messages
//
package vxlan

import (
	api "git.fd.io/govpp.git/api"
	interface_types "git.fd.io/govpp.git/binapi/interface_types"
	ip_types "git.fd.io/govpp.git/binapi/ip_types"
	codec "git.fd.io/govpp.git/codec"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the GoVPP api package it is being compiled against.
// A compilation error at this line likely means your copy of the
// GoVPP api package needs to be updated.
const _ = api.GoVppAPIPackageIsVersion2

const (
	// APIFile is the name of this module.
	APIFile = "vxlan"
	// APIVersion is the API version of this module.
	APIVersion = "2.0.0"
	// VersionCrc is the CRC of this module.
	VersionCrc = 0xf11ad29f
)

// SwInterfaceSetVxlanBypass defines message 'sw_interface_set_vxlan_bypass'.
type SwInterfaceSetVxlanBypass struct {
	SwIfIndex interface_types.InterfaceIndex `binapi:"interface_index,name=sw_if_index" json:"sw_if_index,omitempty"`
	IsIPv6    bool                           `binapi:"bool,name=is_ipv6" json:"is_ipv6,omitempty"`
	Enable    bool                           `binapi:"bool,name=enable,default=true" json:"enable,omitempty"`
}

func (m *SwInterfaceSetVxlanBypass) Reset()               { *m = SwInterfaceSetVxlanBypass{} }
func (*SwInterfaceSetVxlanBypass) GetMessageName() string { return "sw_interface_set_vxlan_bypass" }
func (*SwInterfaceSetVxlanBypass) GetCrcString() string   { return "65247409" }
func (*SwInterfaceSetVxlanBypass) GetMessageType() api.MessageType {
	return api.RequestMessage
}

func (m *SwInterfaceSetVxlanBypass) Size() int {
	if m == nil {
		return 0
	}
	var size int
	size += 4 // m.SwIfIndex
	size += 1 // m.IsIPv6
	size += 1 // m.Enable
	return size
}
func (m *SwInterfaceSetVxlanBypass) Marshal(b []byte) ([]byte, error) {
	var buf *codec.Buffer
	if b == nil {
		buf = codec.NewBuffer(make([]byte, m.Size()))
	} else {
		buf = codec.NewBuffer(b)
	}
	buf.EncodeUint32(uint32(m.SwIfIndex))
	buf.EncodeBool(m.IsIPv6)
	buf.EncodeBool(m.Enable)
	return buf.Bytes(), nil
}
func (m *SwInterfaceSetVxlanBypass) Unmarshal(b []byte) error {
	buf := codec.NewBuffer(b)
	m.SwIfIndex = interface_types.InterfaceIndex(buf.DecodeUint32())
	m.IsIPv6 = buf.DecodeBool()
	m.Enable = buf.DecodeBool()
	return nil
}

// SwInterfaceSetVxlanBypassReply defines message 'sw_interface_set_vxlan_bypass_reply'.
type SwInterfaceSetVxlanBypassReply struct {
	Retval int32 `binapi:"i32,name=retval" json:"retval,omitempty"`
}

func (m *SwInterfaceSetVxlanBypassReply) Reset() { *m = SwInterfaceSetVxlanBypassReply{} }
func (*SwInterfaceSetVxlanBypassReply) GetMessageName() string {
	return "sw_interface_set_vxlan_bypass_reply"
}
func (*SwInterfaceSetVxlanBypassReply) GetCrcString() string { return "e8d4e804" }
func (*SwInterfaceSetVxlanBypassReply) GetMessageType() api.MessageType {
	return api.ReplyMessage
}

func (m *SwInterfaceSetVxlanBypassReply) Size() int {
	if m == nil {
		return 0
	}
	var size int
	size += 4 // m.Retval
	return size
}
func (m *SwInterfaceSetVxlanBypassReply) Marshal(b []byte) ([]byte, error) {
	var buf *codec.Buffer
	if b == nil {
		buf = codec.NewBuffer(make([]byte, m.Size()))
	} else {
		buf = codec.NewBuffer(b)
	}
	buf.EncodeUint32(uint32(m.Retval))
	return buf.Bytes(), nil
}
func (m *SwInterfaceSetVxlanBypassReply) Unmarshal(b []byte) error {
	buf := codec.NewBuffer(b)
	m.Retval = int32(buf.DecodeUint32())
	return nil
}

// VxlanAddDelTunnel defines message 'vxlan_add_del_tunnel'.
type VxlanAddDelTunnel struct {
	IsAdd          bool                           `binapi:"bool,name=is_add,default=true" json:"is_add,omitempty"`
	Instance       uint32                         `binapi:"u32,name=instance" json:"instance,omitempty"`
	SrcAddress     ip_types.Address               `binapi:"address,name=src_address" json:"src_address,omitempty"`
	DstAddress     ip_types.Address               `binapi:"address,name=dst_address" json:"dst_address,omitempty"`
	McastSwIfIndex interface_types.InterfaceIndex `binapi:"interface_index,name=mcast_sw_if_index" json:"mcast_sw_if_index,omitempty"`
	EncapVrfID     uint32                         `binapi:"u32,name=encap_vrf_id" json:"encap_vrf_id,omitempty"`
	DecapNextIndex uint32                         `binapi:"u32,name=decap_next_index" json:"decap_next_index,omitempty"`
	Vni            uint32                         `binapi:"u32,name=vni" json:"vni,omitempty"`
}

func (m *VxlanAddDelTunnel) Reset()               { *m = VxlanAddDelTunnel{} }
func (*VxlanAddDelTunnel) GetMessageName() string { return "vxlan_add_del_tunnel" }
func (*VxlanAddDelTunnel) GetCrcString() string   { return "a35dc8f5" }
func (*VxlanAddDelTunnel) GetMessageType() api.MessageType {
	return api.RequestMessage
}

func (m *VxlanAddDelTunnel) Size() int {
	if m == nil {
		return 0
	}
	var size int
	size += 1      // m.IsAdd
	size += 4      // m.Instance
	size += 1      // m.SrcAddress.Af
	size += 1 * 16 // m.SrcAddress.Un
	size += 1      // m.DstAddress.Af
	size += 1 * 16 // m.DstAddress.Un
	size += 4      // m.McastSwIfIndex
	size += 4      // m.EncapVrfID
	size += 4      // m.DecapNextIndex
	size += 4      // m.Vni
	return size
}
func (m *VxlanAddDelTunnel) Marshal(b []byte) ([]byte, error) {
	var buf *codec.Buffer
	if b == nil {
		buf = codec.NewBuffer(make([]byte, m.Size()))
	} else {
		buf = codec.NewBuffer(b)
	}
	buf.EncodeBool(m.IsAdd)
	buf.EncodeUint32(uint32(m.Instance))
	buf.EncodeUint8(uint8(m.SrcAddress.Af))
	buf.EncodeBytes(m.SrcAddress.Un.XXX_UnionData[:], 0)
	buf.EncodeUint8(uint8(m.DstAddress.Af))
	buf.EncodeBytes(m.DstAddress.Un.XXX_UnionData[:], 0)
	buf.EncodeUint32(uint32(m.McastSwIfIndex))
	buf.EncodeUint32(uint32(m.EncapVrfID))
	buf.EncodeUint32(uint32(m.DecapNextIndex))
	buf.EncodeUint32(uint32(m.Vni))
	return buf.Bytes(), nil
}
func (m *VxlanAddDelTunnel) Unmarshal(b []byte) error {
	buf := codec.NewBuffer(b)
	m.IsAdd = buf.DecodeBool()
	m.Instance = buf.DecodeUint32()
	m.SrcAddress.Af = ip_types.AddressFamily(buf.DecodeUint8())
	copy(m.SrcAddress.Un.XXX_UnionData[:], buf.DecodeBytes(16))
	m.DstAddress.Af = ip_types.AddressFamily(buf.DecodeUint8())
	copy(m.DstAddress.Un.XXX_UnionData[:], buf.DecodeBytes(16))
	m.McastSwIfIndex = interface_types.InterfaceIndex(buf.DecodeUint32())
	m.EncapVrfID = buf.DecodeUint32()
	m.DecapNextIndex = buf.DecodeUint32()
	m.Vni = buf.DecodeUint32()
	return nil
}

// VxlanAddDelTunnelReply defines message 'vxlan_add_del_tunnel_reply'.
type VxlanAddDelTunnelReply struct {
	Retval    int32                          `binapi:"i32,name=retval" json:"retval,omitempty"`
	SwIfIndex interface_types.InterfaceIndex `binapi:"interface_index,name=sw_if_index" json:"sw_if_index,omitempty"`
}

func (m *VxlanAddDelTunnelReply) Reset()               { *m = VxlanAddDelTunnelReply{} }
func (*VxlanAddDelTunnelReply) GetMessageName() string { return "vxlan_add_del_tunnel_reply" }
func (*VxlanAddDelTunnelReply) GetCrcString() string   { return "5383d31f" }
func (*VxlanAddDelTunnelReply) GetMessageType() api.MessageType {
	return api.ReplyMessage
}

func (m *VxlanAddDelTunnelReply) Size() int {
	if m == nil {
		return 0
	}
	var size int
	size += 4 // m.Retval
	size += 4 // m.SwIfIndex
	return size
}
func (m *VxlanAddDelTunnelReply) Marshal(b []byte) ([]byte, error) {
	var buf *codec.Buffer
	if b == nil {
		buf = codec.NewBuffer(make([]byte, m.Size()))
	} else {
		buf = codec.NewBuffer(b)
	}
	buf.EncodeUint32(uint32(m.Retval))
	buf.EncodeUint32(uint32(m.SwIfIndex))
	return buf.Bytes(), nil
}
func (m *VxlanAddDelTunnelReply) Unmarshal(b []byte) error {
	buf := codec.NewBuffer(b)
	m.Retval = int32(buf.DecodeUint32())
	m.SwIfIndex = interface_types.InterfaceIndex(buf.DecodeUint32())
	return nil
}

// VxlanOffloadRx defines message 'vxlan_offload_rx'.
type VxlanOffloadRx struct {
	HwIfIndex interface_types.InterfaceIndex `binapi:"interface_index,name=hw_if_index" json:"hw_if_index,omitempty"`
	SwIfIndex interface_types.InterfaceIndex `binapi:"interface_index,name=sw_if_index" json:"sw_if_index,omitempty"`
	Enable    bool                           `binapi:"bool,name=enable,default=true" json:"enable,omitempty"`
}

func (m *VxlanOffloadRx) Reset()               { *m = VxlanOffloadRx{} }
func (*VxlanOffloadRx) GetMessageName() string { return "vxlan_offload_rx" }
func (*VxlanOffloadRx) GetCrcString() string   { return "89a1564b" }
func (*VxlanOffloadRx) GetMessageType() api.MessageType {
	return api.RequestMessage
}

func (m *VxlanOffloadRx) Size() int {
	if m == nil {
		return 0
	}
	var size int
	size += 4 // m.HwIfIndex
	size += 4 // m.SwIfIndex
	size += 1 // m.Enable
	return size
}
func (m *VxlanOffloadRx) Marshal(b []byte) ([]byte, error) {
	var buf *codec.Buffer
	if b == nil {
		buf = codec.NewBuffer(make([]byte, m.Size()))
	} else {
		buf = codec.NewBuffer(b)
	}
	buf.EncodeUint32(uint32(m.HwIfIndex))
	buf.EncodeUint32(uint32(m.SwIfIndex))
	buf.EncodeBool(m.Enable)
	return buf.Bytes(), nil
}
func (m *VxlanOffloadRx) Unmarshal(b []byte) error {
	buf := codec.NewBuffer(b)
	m.HwIfIndex = interface_types.InterfaceIndex(buf.DecodeUint32())
	m.SwIfIndex = interface_types.InterfaceIndex(buf.DecodeUint32())
	m.Enable = buf.DecodeBool()
	return nil
}

// VxlanOffloadRxReply defines message 'vxlan_offload_rx_reply'.
type VxlanOffloadRxReply struct {
	Retval int32 `binapi:"i32,name=retval" json:"retval,omitempty"`
}

func (m *VxlanOffloadRxReply) Reset()               { *m = VxlanOffloadRxReply{} }
func (*VxlanOffloadRxReply) GetMessageName() string { return "vxlan_offload_rx_reply" }
func (*VxlanOffloadRxReply) GetCrcString() string   { return "e8d4e804" }
func (*VxlanOffloadRxReply) GetMessageType() api.MessageType {
	return api.ReplyMessage
}

func (m *VxlanOffloadRxReply) Size() int {
	if m == nil {
		return 0
	}
	var size int
	size += 4 // m.Retval
	return size
}
func (m *VxlanOffloadRxReply) Marshal(b []byte) ([]byte, error) {
	var buf *codec.Buffer
	if b == nil {
		buf = codec.NewBuffer(make([]byte, m.Size()))
	} else {
		buf = codec.NewBuffer(b)
	}
	buf.EncodeUint32(uint32(m.Retval))
	return buf.Bytes(), nil
}
func (m *VxlanOffloadRxReply) Unmarshal(b []byte) error {
	buf := codec.NewBuffer(b)
	m.Retval = int32(buf.DecodeUint32())
	return nil
}

// VxlanTunnelDetails defines message 'vxlan_tunnel_details'.
type VxlanTunnelDetails struct {
	SwIfIndex      interface_types.InterfaceIndex `binapi:"interface_index,name=sw_if_index" json:"sw_if_index,omitempty"`
	Instance       uint32                         `binapi:"u32,name=instance" json:"instance,omitempty"`
	SrcAddress     ip_types.Address               `binapi:"address,name=src_address" json:"src_address,omitempty"`
	DstAddress     ip_types.Address               `binapi:"address,name=dst_address" json:"dst_address,omitempty"`
	McastSwIfIndex interface_types.InterfaceIndex `binapi:"interface_index,name=mcast_sw_if_index" json:"mcast_sw_if_index,omitempty"`
	EncapVrfID     uint32                         `binapi:"u32,name=encap_vrf_id" json:"encap_vrf_id,omitempty"`
	DecapNextIndex uint32                         `binapi:"u32,name=decap_next_index" json:"decap_next_index,omitempty"`
	Vni            uint32                         `binapi:"u32,name=vni" json:"vni,omitempty"`
}

func (m *VxlanTunnelDetails) Reset()               { *m = VxlanTunnelDetails{} }
func (*VxlanTunnelDetails) GetMessageName() string { return "vxlan_tunnel_details" }
func (*VxlanTunnelDetails) GetCrcString() string   { return "e782f70f" }
func (*VxlanTunnelDetails) GetMessageType() api.MessageType {
	return api.ReplyMessage
}

func (m *VxlanTunnelDetails) Size() int {
	if m == nil {
		return 0
	}
	var size int
	size += 4      // m.SwIfIndex
	size += 4      // m.Instance
	size += 1      // m.SrcAddress.Af
	size += 1 * 16 // m.SrcAddress.Un
	size += 1      // m.DstAddress.Af
	size += 1 * 16 // m.DstAddress.Un
	size += 4      // m.McastSwIfIndex
	size += 4      // m.EncapVrfID
	size += 4      // m.DecapNextIndex
	size += 4      // m.Vni
	return size
}
func (m *VxlanTunnelDetails) Marshal(b []byte) ([]byte, error) {
	var buf *codec.Buffer
	if b == nil {
		buf = codec.NewBuffer(make([]byte, m.Size()))
	} else {
		buf = codec.NewBuffer(b)
	}
	buf.EncodeUint32(uint32(m.SwIfIndex))
	buf.EncodeUint32(uint32(m.Instance))
	buf.EncodeUint8(uint8(m.SrcAddress.Af))
	buf.EncodeBytes(m.SrcAddress.Un.XXX_UnionData[:], 0)
	buf.EncodeUint8(uint8(m.DstAddress.Af))
	buf.EncodeBytes(m.DstAddress.Un.XXX_UnionData[:], 0)
	buf.EncodeUint32(uint32(m.McastSwIfIndex))
	buf.EncodeUint32(uint32(m.EncapVrfID))
	buf.EncodeUint32(uint32(m.DecapNextIndex))
	buf.EncodeUint32(uint32(m.Vni))
	return buf.Bytes(), nil
}
func (m *VxlanTunnelDetails) Unmarshal(b []byte) error {
	buf := codec.NewBuffer(b)
	m.SwIfIndex = interface_types.InterfaceIndex(buf.DecodeUint32())
	m.Instance = buf.DecodeUint32()
	m.SrcAddress.Af = ip_types.AddressFamily(buf.DecodeUint8())
	copy(m.SrcAddress.Un.XXX_UnionData[:], buf.DecodeBytes(16))
	m.DstAddress.Af = ip_types.AddressFamily(buf.DecodeUint8())
	copy(m.DstAddress.Un.XXX_UnionData[:], buf.DecodeBytes(16))
	m.McastSwIfIndex = interface_types.InterfaceIndex(buf.DecodeUint32())
	m.EncapVrfID = buf.DecodeUint32()
	m.DecapNextIndex = buf.DecodeUint32()
	m.Vni = buf.DecodeUint32()
	return nil
}

// VxlanTunnelDump defines message 'vxlan_tunnel_dump'.
type VxlanTunnelDump struct {
	SwIfIndex interface_types.InterfaceIndex `binapi:"interface_index,name=sw_if_index" json:"sw_if_index,omitempty"`
}

func (m *VxlanTunnelDump) Reset()               { *m = VxlanTunnelDump{} }
func (*VxlanTunnelDump) GetMessageName() string { return "vxlan_tunnel_dump" }
func (*VxlanTunnelDump) GetCrcString() string   { return "f9e6675e" }
func (*VxlanTunnelDump) GetMessageType() api.MessageType {
	return api.RequestMessage
}

func (m *VxlanTunnelDump) Size() int {
	if m == nil {
		return 0
	}
	var size int
	size += 4 // m.SwIfIndex
	return size
}
func (m *VxlanTunnelDump) Marshal(b []byte) ([]byte, error) {
	var buf *codec.Buffer
	if b == nil {
		buf = codec.NewBuffer(make([]byte, m.Size()))
	} else {
		buf = codec.NewBuffer(b)
	}
	buf.EncodeUint32(uint32(m.SwIfIndex))
	return buf.Bytes(), nil
}
func (m *VxlanTunnelDump) Unmarshal(b []byte) error {
	buf := codec.NewBuffer(b)
	m.SwIfIndex = interface_types.InterfaceIndex(buf.DecodeUint32())
	return nil
}

func init() { file_vxlan_binapi_init() }
func file_vxlan_binapi_init() {
	api.RegisterMessage((*SwInterfaceSetVxlanBypass)(nil), "sw_interface_set_vxlan_bypass_65247409")
	api.RegisterMessage((*SwInterfaceSetVxlanBypassReply)(nil), "sw_interface_set_vxlan_bypass_reply_e8d4e804")
	api.RegisterMessage((*VxlanAddDelTunnel)(nil), "vxlan_add_del_tunnel_a35dc8f5")
	api.RegisterMessage((*VxlanAddDelTunnelReply)(nil), "vxlan_add_del_tunnel_reply_5383d31f")
	api.RegisterMessage((*VxlanOffloadRx)(nil), "vxlan_offload_rx_89a1564b")
	api.RegisterMessage((*VxlanOffloadRxReply)(nil), "vxlan_offload_rx_reply_e8d4e804")
	api.RegisterMessage((*VxlanTunnelDetails)(nil), "vxlan_tunnel_details_e782f70f")
	api.RegisterMessage((*VxlanTunnelDump)(nil), "vxlan_tunnel_dump_f9e6675e")
}

// Messages returns list of all messages in this module.
func AllMessages() []api.Message {
	return []api.Message{
		(*SwInterfaceSetVxlanBypass)(nil),
		(*SwInterfaceSetVxlanBypassReply)(nil),
		(*VxlanAddDelTunnel)(nil),
		(*VxlanAddDelTunnelReply)(nil),
		(*VxlanOffloadRx)(nil),
		(*VxlanOffloadRxReply)(nil),
		(*VxlanTunnelDetails)(nil),
		(*VxlanTunnelDump)(nil),
	}
}
