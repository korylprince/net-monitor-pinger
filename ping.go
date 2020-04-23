package main

import (
	"fmt"
	"log"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/korylprince/go-icmpv4/v2/echo"
)

const ICMPEchoRequestIdentifier uint16 = 0x3039

type Ping struct {
	*Device
	IP       net.IP
	Sequence uint16
	SentTime time.Time
	RecvTime *time.Time
}

type PingService struct {
	sequence chan uint16

	devices  chan *Device
	requests chan *Ping

	packets chan *echo.IPPacket
	replies chan *pong

	pending   map[uint16]*Ping
	pendingMu *sync.RWMutex

	listener func(p *Ping)

	errors chan error
}

func (p *PingService) sequencer() {
	var s uint16 = 0
	for {
		p.sequence <- s
		s++
	}
}

func (p *PingService) nextSequence() uint16 {
	return <-(p.sequence)
}

func (p *PingService) requester() {
	for d := range p.devices {
		d.mu.RLock()
		ips := make([]net.IP, len(d.ips))
		copy(ips, d.ips)
		d.mu.RUnlock()
		for _, ip := range ips {
			seq := p.nextSequence()
			t := time.Now()
			p.pendingMu.Lock()
			p.pending[seq] = &Ping{Device: d, IP: ip, Sequence: seq, SentTime: t}
			p.pendingMu.Unlock()
			err := echo.Send(nil, &net.IPAddr{IP: ip}, ICMPEchoRequestIdentifier, seq)
			if err != nil {
				log.Printf("PingService: Unable to send ping request to %v: %v", ip, err)
				p.pendingMu.Lock()
				delete(p.pending, seq)
				p.pendingMu.Unlock()
			}
		}
	}
}

type pong struct {
	IP       net.IP
	Sequence uint16
	RecvTime time.Time
}

func (p *PingService) receiver() {
	for pk := range p.packets {
		recv := time.Now()
		if pk.Identifier() != ICMPEchoRequestIdentifier {
			continue
		}
		p.pendingMu.Lock()
		if req, ok := p.pending[pk.Sequence()]; ok {
			if req.IP.Equal(pk.RemoteAddr.IP) {
				req.RecvTime = &recv
			} else {
				log.Printf("PingService: Mismatched IP: Original IP %s, Received IP: %s, Sequence: %d\n", req.IP.String(), pk.RemoteAddr.IP.String(), req.Sequence)
			}
		} else {
			log.Printf("PingService: Unknown Sequence: IP %s, Sequence: %d\n", pk.RemoteAddr.IP.String(), pk.Sequence())
		}
		p.pendingMu.Unlock()
	}
}

func (p *PingService) scavenger(timeout time.Duration) {
	for {
		time.Sleep(timeout / 2)
		p.pendingMu.Lock()
		done := make([]uint16, 0)

		for seq, req := range p.pending {
			if req.RecvTime != nil || time.Now().After(req.SentTime.Add(timeout)) {
				done = append(done, seq)
			}
		}

		if len(done) > 0 {
			for _, seq := range done {
				if p.listener != nil {
					go p.listener(p.pending[seq])
				}
				delete(p.pending, seq)
			}
		}
		p.pendingMu.Unlock()
	}
}

func (p *PingService) errorLogger() {
	for err := range p.errors {
		log.Println("PingService: Unknown ICMP Error:", err)
	}
}

func NewPingService(workers, buffer int, timeout time.Duration) (*PingService, error) {
	p := &PingService{
		sequence:  make(chan uint16),
		devices:   make(chan *Device),
		requests:  make(chan *Ping, buffer),
		packets:   make(chan *echo.IPPacket, buffer),
		replies:   make(chan *pong, buffer),
		pending:   make(map[uint16]*Ping),
		pendingMu: new(sync.RWMutex),
		errors:    make(chan error),
	}

	ips, err := echo.ListenerAll(p.packets, p.errors, nil)
	if err != nil {
		return nil, fmt.Errorf("Unable to start listeners: %v", err)
	}

	ipStrs := make([]string, 0)
	for _, ip := range ips {
		ipStrs = append(ipStrs, ip.String())
	}

	log.Println("PingService: Listening on:", strings.Join(ipStrs, ", "))

	log.Println("PingService: Starting", workers, "workers")
	go p.sequencer()
	for i := 0; i < workers; i++ {
		go p.requester()
	}
	go p.receiver()
	go p.scavenger(timeout)
	go p.errorLogger()

	return p, nil
}

func (p *PingService) SetListener(f func(p *Ping)) {
	p.listener = f
}

func (p *PingService) Ping(d *Device) {
	p.devices <- d
}
