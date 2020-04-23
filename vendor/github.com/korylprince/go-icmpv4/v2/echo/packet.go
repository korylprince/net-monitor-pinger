package echo

import (
	"net"

	"github.com/korylprince/go-icmpv4/v2"
)

//Packet represents an ICMPv4 Echo Request/Reply
type Packet struct {
	*icmpv4.Packet
}

//Identifier gets the ICMPv4 Echo Request/Reply identifier
func (e *Packet) Identifier() uint16 {
	return e.HeaderOptions.Uint16(0)
}

//SetIdentifier sets the ICMPv4 Echo Request/Reply identifier
func (e *Packet) SetIdentifier(i uint16) {
	e.HeaderOptions.SetUint16(0, i)
}

//Sequence gets the ICMPv4 Echo Request/Reply sequence
func (e *Packet) Sequence() uint16 {
	return e.HeaderOptions.Uint16(1)
}

//SetSequence sets the ICMPv4 Echo Request/Reply sequence
func (e *Packet) SetSequence(i uint16) {
	e.HeaderOptions.SetUint16(1, i)
}

//NewEchoRequest creates a new ICMPv4 Echo Request with the given identifier and sequence
func NewEchoRequest(identifier, sequence uint16) *Packet {
	p := Packet{Packet: &icmpv4.Packet{Type: 8, Code: 0}}
	p.SetIdentifier(identifier)
	p.SetSequence(sequence)
	return &p
}

//IPPacket is a wrapper for Packet with IP information
type IPPacket struct {
	*Packet
	RemoteAddr *net.IPAddr
	LocalAddr  *net.IPAddr
}
