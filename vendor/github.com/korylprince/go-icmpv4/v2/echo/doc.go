/*
Package echo provides a thin wrapper over github.com/korylprince/go-icmpv4/v2 solely for ICMPv4 Echo Request/Reply packets.
This package makes it easy to add ICMPv4 pings to your program. This example will ping every IP on a subnet and print responses as they come in:

	package main

	import (
		"fmt"
		"net"
		"strconv"
		"time"

		"github.com/korylprince/go-icmpv4/v2/echo"
	)

	func printer(in <-chan *echo.IPPacket) {
		for {
			fmt.Println("Response from:", (<-in).RemoteAddr.String())
		}
	}

	func errPrinter(in <-chan error) {
		for {
			fmt.Printf("%#v\n", <-in)
		}
	}

	func main() {
		//set up channels
		packets := make(chan *echo.IPPacket)
		go printer(packets)
		errors := make(chan error)
		go errPrinter(errors)
		done := make(chan struct{})

		//start listener
		intList, err := echo.ListenerAll(packets, errors, done)
		if err != nil {
			panic(err)
		}
		for _, intfc := range intList {
			fmt.Println("Listening on:", intfc)
		}

		//send pings to all IPs on subnet
		for i := 1; i < 255; i++ {
			raddr, err := net.ResolveIPAddr("ip4", "192.168.100."+strconv.Itoa(i))
			if err != nil {
				panic(err)
			}
			err = echo.Send(nil, raddr, 0x1234, 1)
			if err != nil {
				panic(err)
			}
		}

		//wait to receive replies
		time.Sleep(5 * time.Second)

		//shut down listener
		close(done)
	}

*/
package echo
