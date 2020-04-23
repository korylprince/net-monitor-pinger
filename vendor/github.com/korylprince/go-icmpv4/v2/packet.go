package icmpv4

import "net"

//ICMPv4HeaderLength is the length of the ICMPv4 header in bytes
const ICMPv4HeaderLength = 8

//HeaderOptions is a 4-byte part of the ICMPv4 header. The value(s) it represents depends on the Type and Code.
type HeaderOptions uint32

//Byte returns the nth byte of the HeaderOptions, indexed by 0. [0,3] is the only acceptable range.
func (h HeaderOptions) Byte(n uint8) byte {
	return byte(uint32(h) >> (8 * (3 - n)))
}

//SetByte sets the nth byte of the HeaderOptions to b, indexed by 0. [0,3] is the only acceptable range.
func (h *HeaderOptions) SetByte(n uint8, b byte) {
	var mask uint32 = ^(0xFF << (8 * (3 - n)))
	bShift := uint32(b) << (8 * (3 - n))
	*h = HeaderOptions((uint32(*h) & mask) | bShift)
}

//Uint16 returns the nth uint16 of the HeaderOptions, indexed by 0. [0,1] is the only acceptable range.
func (h HeaderOptions) Uint16(n uint8) uint16 {
	return uint16(uint32(h) >> (16 * (1 - n)))
}

//SetUint16 sets the nth uint16 of the HeaderOptions to i, indexed by 0. [0,1] is the only acceptable range.
func (h *HeaderOptions) SetUint16(n uint8, i uint16) {
	var mask uint32 = ^(0xFFFF << (16 * (1 - n)))
	iShift := uint32(i) << (16 * (1 - n))
	*h = HeaderOptions((uint32(*h) & mask) | iShift)
}

//Packet represents an ICMPv4 packet
type Packet struct {
	Type          uint8
	Code          uint8
	Checksum      uint16
	HeaderOptions HeaderOptions
	Body          []byte
}

//InvalidPacketError denotes and error parsing an ICMPv4 packet
type InvalidPacketError string

func (err InvalidPacketError) Error() string {
	return string(err)
}

//Parse parses a raw ICMPv4 packet and returns an Packet, or an error if one occurred
func Parse(b []byte) (*Packet, error) {
	if len(b) < ICMPv4HeaderLength {
		return nil, InvalidPacketError("Malformed headers")
	}
	p := Packet{
		Type:          b[0],
		Code:          b[1],
		Checksum:      uint16(b[2])<<8 | uint16(b[3]),
		HeaderOptions: HeaderOptions(uint32(b[4])<<24 | uint32(b[5])<<16 | uint32(b[6])<<8 | uint32(b[7])),
	}
	if len(b) > ICMPv4HeaderLength {
		p.Body = b[ICMPv4HeaderLength:]
	}
	if checksum(b) != 0 {
		return nil, InvalidPacketError("Invalid checksum")
	}
	return &p, nil
}

//Marshal creates a raw ICMPv4 packet from an Packet
func (p *Packet) Marshal() []byte {
	b := make([]byte, ICMPv4HeaderLength+len(p.Body))
	b[0] = p.Type
	b[1] = p.Code
	b[4] = byte(p.HeaderOptions >> 24)
	b[5] = byte(p.HeaderOptions >> 16)
	b[6] = byte(p.HeaderOptions >> 8)
	b[7] = byte(p.HeaderOptions)

	if len(p.Body) > 0 {
		copy(b[ICMPv4HeaderLength:], p.Body)
	}
	chksum := checksum(b)
	b[2] = byte(chksum >> 8)
	b[3] = byte(chksum)
	return b
}

//IPPacket is a wrapper for Packet with IP information
type IPPacket struct {
	*Packet
	LocalAddr  *net.IPAddr
	RemoteAddr *net.IPAddr
}
