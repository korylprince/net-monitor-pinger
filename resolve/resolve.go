package resolve

import (
	"fmt"
	"log"
	"net"
)

type host struct {
	Hostname string
	IPs      []net.IP
	Error    error
	callback chan *host
}

type Service struct {
	in chan *host

	poolIn  chan chan *host
	poolOut chan chan *host
}

func (s *Service) pooler() {
	for {
		s.poolOut <- <-s.poolIn
	}
}

func (s *Service) resolver() {
	for host := range s.in {
		ips, err := net.LookupIP(host.Hostname)
		if err != nil {
			host.Error = fmt.Errorf("Unable to lookup host: %v", err)
			host.callback <- host
			continue
		}

		host.IPs = make([]net.IP, 0)
		for _, ip := range ips {
			if ipv4 := ip.To4(); ipv4 != nil {
				host.IPs = append(host.IPs, ipv4)
			}
		}
		host.callback <- host
	}
}

func NewService(workers, buffer int) *Service {
	s := &Service{
		in:      make(chan *host),
		poolIn:  make(chan chan *host, buffer),
		poolOut: make(chan chan *host, buffer),
	}
	log.Println("ResolverService: Starting", workers, "workers")

	for i := 0; i < buffer; i++ {
		s.poolIn <- make(chan *host)
	}
	go s.pooler()

	for i := 0; i < workers; i++ {
		go s.resolver()
	}

	return s
}

func (s *Service) Resolve(hostname string) ([]net.IP, error) {
	callback := <-s.poolOut
	s.in <- &host{Hostname: hostname, callback: callback}
	host := <-callback
	s.poolIn <- callback
	return host.IPs, host.Error
}
