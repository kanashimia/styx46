package translator

import (
    "fmt"
    "net"
    "os/exec"
    "sync"
	"time"
	"math/rand"
)

type Entry struct {
    map4 net.IP
    map6 net.IP
    updatedAt time.Time
    createdAt time.Time
}

type Translator struct {
    mu      sync.RWMutex
    pool	*IPNet
    entries []*Entry
}

// stringify Entry
func (e *Entry) String() string {
    return fmt.Sprintf("Map4: %s, Map6: %s, UpdatedAt: %s CreatedAt: %s", e.map4, e.map6, e.updatedAt,e.createdAt)
}

// new Translator
func New(cidr string) (*Translator, error) {
    ip, network, err := net.ParseCIDR(cidr)
    if err != nil {
        return nil, fmt.Errorf("invalid CIDR: %w", err)
    }
    if ip.Equal(network.IP) == false {
        return nil, fmt.Errorf("CIDR must be a network address, not a host address")
    }

    return &Table{pool:network}, nil
}

// Lookup translation entry and return the corresponding Map4 entry
func (t *Translator) Lookup(map6 net.IP) (net.IP, error) {
	//Lock mutex for safety
    t.mu.Lock()
    defer t.mu.Unlock()

	//Search list for this map6 entry
    for _, e := range t.entries {
        if e.map6.Equal(map6) {
			//Already mapped, return the existing entry
			fmt.Printf("Lookup for [%s] found entry [%s]",map6,e)
            e.updatedAt = time.Now()
            return e.map4, nil
        }
    }
	fmt.Printf("Lookup for [%s] not found, allocating...",map6)

	//Find a new entry to allocate
    for ip := cloneIP(t.pool.IP); t.pool.Contains(ip); {
        inUse := false
        for _, e := range t.entries {
            if e.Map4.Equal(ip) {
                inUse = true
                break
            }
        }
        if !inUse {
			//Found a new spot for the entry
            e := &Entry{
                nap4:      cloneIP(ip),
                map6:      map6.To16(),
                updatedAt: time.Now(),
				createdAt: time.Now(),
            }
			fmt.Printf("Lookup for [%s] created entry [%s]",map6,e)
			//add to the entries list
            t.entries = append(t.entries, e)
			//TODO update external translator
            return e.map4, nil
        }
        for i := len(ip) - 1; i >= 0; i-- {
            ip[i]++
            if ip[i] != 0 {
                break
            }
        }
    }

    return nil, fmt.Errorf("pool exhausted")
}

func cloneIP(ip net.IP) net.IP {
    return append(net.IP{}, ip...)
}