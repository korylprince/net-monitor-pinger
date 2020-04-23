package main

import (
	"fmt"
	"log"
	"net"
	"sync"
	"time"
)

type Device struct {
	ID       string `json:"id"`
	Hostname string `json:"hostname"`

	ips []net.IP
	mu  *sync.RWMutex
}

type Manager struct {
	r *ResolverService
	p *PingService
	g *GraphQLService

	devices map[string]*Device
	devMu   *sync.RWMutex

	buf   []*Ping
	bufMu *sync.Mutex
}

func (m *Manager) syncer(devices []*Device) {
	m.devMu.Lock()
	for _, dNew := range devices {
		if dOld, ok := m.devices[dNew.ID]; ok && dNew.Hostname != dOld.Hostname {
			dOld.mu.Lock()
			dOld.Hostname = dNew.Hostname
			dOld.ips = make([]net.IP, 0)
			dOld.mu.Unlock()
			m.r.Resolve(dOld)
		} else {
			m.devices[dNew.ID] = &Device{
				ID:       dNew.ID,
				Hostname: dNew.Hostname,
				ips:      make([]net.IP, 0),
				mu:       new(sync.RWMutex),
			}
			m.r.Resolve(m.devices[dNew.ID])
		}
	}
	for _, dOld := range m.devices {
		if _, ok := m.devices[dOld.ID]; !ok {
			delete(m.devices, dOld.ID)
		}
	}
	m.devMu.Unlock()

	log.Println("Manager: synced", len(devices), "Devices")
}

func (m *Manager) resolver(interval time.Duration) {
	for {
		time.Sleep(interval)
		m.devMu.RLock()
		for _, d := range m.devices {
			m.r.Resolve(d)
		}
		m.devMu.RUnlock()
	}
}

func (m *Manager) pinger(interval time.Duration) {
	for {
		time.Sleep(interval)
		m.devMu.RLock()
		for _, d := range m.devices {
			if len(d.ips) > 0 {
				m.p.Ping(d)
			}
		}
		m.devMu.RUnlock()
	}
}

func (m *Manager) buffer(e *Ping) {
	m.bufMu.Lock()
	m.buf = append(m.buf, e)
	m.bufMu.Unlock()
}

func (m *Manager) writer(interval time.Duration) {
	for {
		time.Sleep(interval)

		m.bufMu.Lock()
		if len(m.buf) == 0 {
			m.bufMu.Unlock()
			continue
		}
		buf := make([]*Ping, len(m.buf))
		copy(buf, m.buf)
		m.buf = make([]*Ping, 0)
		m.bufMu.Unlock()

		go func(b []*Ping) {
			if err := m.g.InsertPings(b); err != nil {
				log.Println("Manager: Failed to insert Pings:", err)
			}
		}(buf)
	}
}

func (m *Manager) purger(interval, olderThan time.Duration) {
	for {
		log.Println("Manager: Purging Pings older than:", olderThan)
		if err := m.g.PurgePings(time.Now().Add(-olderThan)); err != nil {
			log.Println("Manager: Unable to purge Pings:", err)
		}
		time.Sleep(interval)
	}
}

func NewManager(c *config) (*Manager, error) {
	r := NewResolverService(c.DNSWorkers)

	p, err := NewPingService(c.PingWorkers, c.PingBufferSize, time.Millisecond*time.Duration(c.PingTimeout))
	if err != nil {
		return nil, fmt.Errorf("Unable to create PingService: %v", err)
	}

	g, err := NewGraphQLService(c.GraphQLEndpoint, c.GraphQLAPISecret)
	if err != nil {
		return nil, fmt.Errorf("Unable to create GraphQLService: %v", err)
	}

	m := &Manager{
		r: r, p: p, g: g,
		devices: make(map[string]*Device),
		devMu:   new(sync.RWMutex),
		buf:     make([]*Ping, 0),
		bufMu:   new(sync.Mutex),
	}

	if err = g.SubscribeDevices(m.syncer); err != nil {
		return nil, fmt.Errorf("Unable to Subscribe to Devices: %v", err)
	}

	p.SetListener(m.buffer)
	go m.pinger(time.Second * time.Duration(c.PingInterval))
	go m.writer(time.Second * time.Duration(c.PingInterval))
	go m.purger(time.Minute*time.Duration(c.PurgeInterval), time.Minute*time.Duration(c.PurgeOlderThan))
	go m.resolver(time.Minute * time.Duration(c.DNSLookupInterval))

	log.Println("Manager: Successfully started")

	return m, nil
}
