package ping

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
	IP       net.IP
	Sequence uint16
	SentTime time.Time
	RecvTime *time.Time
	err      error
	callback chan *Ping
}

type Service struct {
	sequence chan uint16

	poolIn  chan chan *Ping
	poolOut chan chan *Ping

	requests chan *Ping

	packets chan *echo.IPPacket

	pending   map[uint16]*Ping
	pendingMu *sync.RWMutex

	errors chan error
}

func (s *Service) sequencer() {
	var seq uint16 = 0
	for {
		s.sequence <- seq
		seq++
	}
}

func (s *Service) nextSequence() uint16 {
	return <-(s.sequence)
}

func (s *Service) pooler() {
	for {
		s.poolOut <- <-s.poolIn
	}
}

func (s *Service) requester() {
	for req := range s.requests {
		seq := s.nextSequence()
		t := time.Now()
		req.Sequence = seq
		req.SentTime = t
		ip := req.IP
		s.pendingMu.Lock()
		s.pending[seq] = req
		s.pendingMu.Unlock()
		err := echo.Send(nil, &net.IPAddr{IP: ip}, ICMPEchoRequestIdentifier, seq)
		if err != nil {
			s.pendingMu.Lock()
			s.pending[seq].err = fmt.Errorf("Unable to send echo request: %v", err)
			s.pendingMu.Unlock()
		}
	}
}

func (s *Service) receiver() {
	for pk := range s.packets {
		recv := time.Now()
		if pk.Identifier() != ICMPEchoRequestIdentifier {
			continue
		}
		s.pendingMu.Lock()
		if req, ok := s.pending[pk.Sequence()]; ok {
			if req.IP.Equal(pk.RemoteAddr.IP) {
				req.RecvTime = &recv
			} else {
				log.Printf("PingService: Mismatched IP: Original IP %s, Received IP: %s, Sequence: %d\n", req.IP.String(), pk.RemoteAddr.IP.String(), req.Sequence)
			}
		} else {
			log.Printf("PingService: Unknown Sequence: IP %s, Sequence: %d\n", pk.RemoteAddr.IP.String(), pk.Sequence())
		}
		s.pendingMu.Unlock()
	}
}

//FIXME use a sync.Cond here
func (s *Service) scavenger(timeout time.Duration) {
	for {
		time.Sleep(timeout / 2)
		s.pendingMu.Lock()
		done := make([]uint16, 0)

		for seq, req := range s.pending {
			if req.RecvTime != nil || time.Now().After(req.SentTime.Add(timeout)) {
				done = append(done, seq)
			}
		}

		if len(done) > 0 {
			for _, seq := range done {
				s.pending[seq].callback <- s.pending[seq]
				delete(s.pending, seq)
			}
		}
		s.pendingMu.Unlock()
	}
}

func (s *Service) errorLogger() {
	for err := range s.errors {
		log.Println("PingService: Unknown ICMP Error:", err)
	}
}

func NewService(workers, buffer int, timeout time.Duration) (*Service, error) {
	s := &Service{
		sequence:  make(chan uint16),
		poolIn:    make(chan chan *Ping, buffer),
		poolOut:   make(chan chan *Ping, buffer),
		requests:  make(chan *Ping, buffer),
		packets:   make(chan *echo.IPPacket, buffer),
		pending:   make(map[uint16]*Ping),
		pendingMu: new(sync.RWMutex),
		errors:    make(chan error),
	}

	ips, err := echo.ListenerAll(s.packets, s.errors, nil)
	if err != nil {
		return nil, fmt.Errorf("Unable to start listeners: %v", err)
	}

	ipStrs := make([]string, 0)
	for _, ip := range ips {
		ipStrs = append(ipStrs, ip.String())
	}

	log.Println("PingService: Listening on:", strings.Join(ipStrs, ", "))
	log.Println("PingService: Starting", workers, "workers")

	go s.sequencer()

	for i := 0; i < buffer; i++ {
		s.poolIn <- make(chan *Ping)
	}
	go s.pooler()

	for i := 0; i < workers; i++ {
		go s.requester()
	}

	go s.receiver()
	go s.scavenger(timeout)
	go s.errorLogger()

	return s, nil
}

func (s *Service) Ping(ip net.IP) (*Ping, error) {
	callback := <-s.poolOut
	s.requests <- &Ping{IP: ip, callback: callback}
	ping := <-callback
	s.poolIn <- callback
	return ping, ping.err
}
