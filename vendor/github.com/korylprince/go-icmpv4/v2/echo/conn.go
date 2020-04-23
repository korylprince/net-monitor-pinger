package echo

import (
	"net"

	"github.com/korylprince/go-icmpv4/v2"
)

//Send sends an ICMPv4 Echo Request to raddr from laddr with the given identifier and sequence
func Send(laddr, raddr *net.IPAddr, identifier, sequence uint16) (err error) {
	p := NewEchoRequest(identifier, sequence)
	return icmpv4.Send(laddr, raddr, p.Marshal())
}

//convertAndFilter
func convertAndFilter(in <-chan *icmpv4.IPPacket, out chan<- *IPPacket) {
	for {
		p := <-in

		if p.Type == 0 && p.Code == 0 {
			out <- &IPPacket{
				Packet:     &Packet{Packet: p.Packet},
				LocalAddr:  p.LocalAddr,
				RemoteAddr: p.RemoteAddr,
			}
		}
	}
}

//Listener parses incoming ICMPv4 Echo Replys from an ICMPv4 net.IPConn and sends packets and errors back on channels.
//When done is closed, it returns an error (or nil) from conn.Close().
func Listener(conn *net.IPConn, packets chan<- *IPPacket, errors chan<- error, done <-chan struct{}) error {
	packetsInternal := make(chan *icmpv4.IPPacket)
	go convertAndFilter(packetsInternal, packets)
	return icmpv4.Listener(conn, packetsInternal, errors, done)
}

//ListenerAll creates a Listener for all IPv4 connections available. It returns a list of addresses that it's
//listening on or an error if it can't get that list.
func ListenerAll(packets chan<- *IPPacket, errors chan<- error, done <-chan struct{}) ([]*net.IPAddr, error) {
	packetsInternal := make(chan *icmpv4.IPPacket)
	go convertAndFilter(packetsInternal, packets)
	return icmpv4.ListenerAll(packetsInternal, errors, done)
}
