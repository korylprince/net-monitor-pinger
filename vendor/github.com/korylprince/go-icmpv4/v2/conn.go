package icmpv4

import (
	"fmt"
	"net"
)

//Dial creates an ICMPv4 *net.IPconn from laddr to raddr
func Dial(laddr, raddr *net.IPAddr) (*net.IPConn, error) {
	return net.DialIP("ip4:icmp", laddr, raddr)
}

//Send sends an ICMPv4 packet from laddr to raddr
func Send(laddr, raddr *net.IPAddr, packet []byte) (err error) {
	conn, err := Dial(laddr, raddr)
	if err != nil {
		return err
	}
	defer func() {
		e := conn.Close()
		if err != nil {
			err = e
		}
	}()
	_, err = conn.Write(packet)
	return err
}

//Listen listens for incoming ICMPv4 packets
func Listen(laddr *net.IPAddr) (*net.IPConn, error) {
	return net.ListenIP("ip4:icmp", laddr)
}

//Listener parses incoming ICMPv4 packets from an ICMPv4 net.IPConn and sends packets and errors back on channels.
//When done is closed, it returns an error (or nil) from conn.Close().
func Listener(conn *net.IPConn, packets chan<- *IPPacket, errors chan<- error, done <-chan struct{}) error {
	laddr, ok := (conn.LocalAddr()).(*net.IPAddr)
	if !ok {
		panic(fmt.Errorf("conn.LocalAddr() != *net.IPAddr"))
	}

	buf := make([]byte, 65535)
	for {
		select {
		case <-done:
			return conn.Close()
		default:
			length, raddr, err := conn.ReadFromIP(buf)
			if err != nil {
				errors <- err
				if length == 0 {
					continue
				}
			}

			packet, err := Parse(buf[:length])
			if err != nil {
				errors <- err
				continue
			}
			packets <- &IPPacket{
				Packet:     packet,
				LocalAddr:  laddr,
				RemoteAddr: raddr,
			}
		}
	}
}

//ListenerAll creates a Listener for all IPv4 connections available. It returns a list of addresses that it's
//listening on or an error if it can't get that list.
func ListenerAll(packets chan<- *IPPacket, errors chan<- error, done <-chan struct{}) ([]*net.IPAddr, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return nil, err
	}
	var laddrs []*net.IPAddr
	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && ipnet.IP.To4() != nil {
			laddr, err := net.ResolveIPAddr("ip4", ipnet.IP.String())
			if err != nil {
				//there's something wrong with go or this code that returning an error won't fix
				panic(err)
			}

			conn, err := Listen(laddr)
			if err != nil {
				continue
			}

			go func() {
				err := Listener(conn, packets, errors, done)
				if err != nil {
					errors <- err
				}
			}()

			laddrs = append(laddrs, laddr)
		}
	}
	return laddrs, nil
}
