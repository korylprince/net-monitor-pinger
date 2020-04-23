package main

import (
	"log"
	"net"
)

//ResolverService is a service to resolve hostnames to IP addresses
type ResolverService struct {
	in chan *Device
}

//NewResolverService returns a new ResolverService with the given number of workers
func NewResolverService(workers int) *ResolverService {
	r := &ResolverService{in: make(chan *Device)}
	log.Println("ResolverService: Starting", workers, "workers")
	for i := 0; i < workers; i++ {
		go r.resolver()
	}
	return r
}

func (r *ResolverService) resolver() {
	for d := range r.in {
		ips, err := net.LookupIP(d.Hostname)
		if err != nil {
			log.Printf("ResolverService: Unable to lookup host %s: %v\n", d.Hostname, err)
			continue
		}

		d.mu.Lock()
		d.ips = make([]net.IP, 0)
		for _, ip := range ips {
			if ipv4 := ip.To4(); ipv4 != nil {
				d.ips = append(d.ips, ipv4)
			}
		}
		d.mu.Unlock()
	}
}

//Resolve resolves the IP Addresses for the given device
func (r *ResolverService) Resolve(d *Device) {
	r.in <- d
}
