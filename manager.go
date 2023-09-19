package main

import (
	"fmt"
	"log"
	"net"
	"sync"
	"time"

	"github.com/korylprince/net-monitor-pinger/ping"
	"github.com/korylprince/net-monitor-pinger/resolve"
)

type Device struct {
	ID       string `json:"id"`
	Hostname string `json:"hostname"`

	ips []net.IP
	mu  *sync.RWMutex
}

type Manager struct {
	r *resolve.Service
	p *ping.Service
	g *GraphQLService

	devices map[string]*Device
	devMu   *sync.RWMutex

	buf   []*ping.Ping
	bufMu *sync.Mutex
}

func (m *Manager) resolve(dev *Device) {
	ips, err := m.r.Resolve(dev.Hostname)
	if err != nil {
		log.Printf("ResolverService: Unable to resolve %s: %v\n", dev.Hostname, err)
		return
	}
	dev.mu.Lock()
	dev.ips = ips
	dev.mu.Unlock()
}

func (m *Manager) syncer(devices []*Device) {
	m.devMu.Lock()
	for _, dNew := range devices {
		if dOld, ok := m.devices[dNew.ID]; ok && dNew.Hostname != dOld.Hostname {
			dOld.mu.Lock()
			dOld.Hostname = dNew.Hostname
			dOld.ips = make([]net.IP, 0)
			dOld.mu.Unlock()
			m.resolve(dOld)
		} else {
			m.devices[dNew.ID] = &Device{
				ID:       dNew.ID,
				Hostname: dNew.Hostname,
				ips:      make([]net.IP, 0),
				mu:       new(sync.RWMutex),
			}
			m.resolve(m.devices[dNew.ID])
		}
	}

outer:
	for _, dOld := range m.devices {
		for _, dNew := range devices {
			if dOld.ID == dNew.ID {
				continue outer
			}
		}
		delete(m.devices, dOld.ID)
	}

	m.devMu.Unlock()

	log.Println("Manager: synced", len(devices), "Devices")
}

func (m *Manager) resolver(interval time.Duration) {
	for {
		time.Sleep(interval)
		m.devMu.RLock()
		for _, d := range m.devices {
			m.resolve(d)
		}
		m.devMu.RUnlock()
	}
}

func (m *Manager) pinger(interval time.Duration) {
	for {
		time.Sleep(interval)
		m.devMu.RLock()
		for _, d := range m.devices {
			if len(d.ips) == 0 {
				continue
			}
			for _, ip := range d.ips {
				go func(i net.IP) {
					p, err := m.p.Ping(i)
					if err != nil {
						log.Printf("Unable to ping %v: %v\n", i, err)
						return
					}
					m.buffer(p)
				}(ip)
			}
		}
		m.devMu.RUnlock()
	}
}

func (m *Manager) buffer(p *ping.Ping) {
	m.bufMu.Lock()
	m.buf = append(m.buf, p)
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
		buf := make([]*ping.Ping, len(m.buf))
		copy(buf, m.buf)
		m.buf = make([]*ping.Ping, 0)
		m.bufMu.Unlock()

		_ = buf
		//FIXME: do something with pings
		// go func(b []*ping.Ping) {
		// 	if err := m.g.InsertPings(b); err != nil {
		// 		log.Println("Manager: Failed to insert Pings:", err)
		// 	}
		// }(buf)
	}
}

func (m *Manager) purger(interval, olderThan time.Duration) {
	for {
		//FIXME: purge pings
		// log.Println("Manager: Purging Pings older than:", olderThan)
		// if err := m.g.PurgePings(time.Now().Add(-olderThan)); err != nil {
		// 	log.Println("Manager: Unable to purge Pings:", err)
		// }
		time.Sleep(interval)
	}
}

func NewManager(c *config) (*Manager, error) {
	r := resolve.NewService(c.DNSWorkers)

	p, err := ping.NewService(c.PingWorkers, c.PingBufferSize, time.Millisecond*time.Duration(c.PingTimeout))
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
		buf:     make([]*ping.Ping, 0),
		bufMu:   new(sync.Mutex),
	}

	if err = g.SubscribeDevices(m.syncer); err != nil {
		return nil, fmt.Errorf("Unable to Subscribe to Devices: %v", err)
	}

	go m.pinger(time.Second * time.Duration(c.PingInterval))
	go m.writer(time.Second * time.Duration(c.PingInterval))
	go m.purger(time.Minute*time.Duration(c.PurgeInterval), time.Minute*time.Duration(c.PurgeOlderThan))
	go m.resolver(time.Minute * time.Duration(c.DNSLookupInterval))

	log.Println("Manager: Successfully started")

	return m, nil
}
