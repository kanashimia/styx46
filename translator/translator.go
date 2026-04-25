package translator

import (
    "fmt"
    "net"
	"os"
    "os/exec"
	"syscall"
    "sync"
	"time"
	"context"
	"path/filepath"
//	"github.com/apalrd/styx46/config"
)

type Entry struct {
    map4 net.IP
    map6 net.IP
    updatedAt time.Time
    createdAt time.Time
}

type Translator struct {
    mu      sync.RWMutex
    pool	*net.IPNet
    entries []*Entry
	binaryPath string
	configPath string
    cmd        *exec.Cmd
}

// stringify Entry
func (e *Entry) String() string {
    return fmt.Sprintf("Map4: %s, Map6: %s, UpdatedAt: %s CreatedAt: %s", e.map4, e.map6, e.updatedAt,e.createdAt)
}

// new Translator
func New(cidr string, binaryPath string) (*Translator, error) {
    ip, network, err := net.ParseCIDR(cidr)
    if err != nil {
        return nil, fmt.Errorf("invalid CIDR: %w", err)
    }
    if ip.Equal(network.IP) == false {
        return nil, fmt.Errorf("CIDR must be a network address, not a host address")
    }

	//Determine the config directory
	dir, err := configPath()
	if err != nil {
        return nil, fmt.Errorf("failed to get config path: %w", err)
    }
	fmt.Printf("Using %s as working directory\n",dir)

    return &Translator{
		pool:network,
		binaryPath:binaryPath,
		configPath:dir,
	}, nil
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
			fmt.Printf("Lookup for [%s] found entry [%s]\n",map6,e)
            e.updatedAt = time.Now()
            return e.map4, nil
        }
    }
	fmt.Printf("Lookup for [%s] not found, allocating...\n",map6)

	//Find a new entry to allocate
    for ip := cloneIP(t.pool.IP); t.pool.Contains(ip); {
        inUse := false
        for _, e := range t.entries {
            if e.map4.Equal(ip) {
                inUse = true
                break
            }
        }
        if !inUse {
			//Found a new spot for the entry
            e := &Entry{
                map4:      cloneIP(ip),
                map6:      map6.To16(),
                updatedAt: time.Now(),
				createdAt: time.Now(),
            }
			fmt.Printf("Lookup for [%s] created entry [%s]\n",map6,e)
			//add to the entries list
            t.entries = append(t.entries, e)
			//update Tayga mapfile
			if err := t.writeMapfile(); err != nil {
				return nil, fmt.Errorf("failed to write mapfile: %w\n", err)
			}
			//reload Mapfile
			if t.cmd != nil {
				if err := t.cmd.Process.Signal(syscall.SIGHUP); err != nil {
					return nil, fmt.Errorf("failed to signal tayga: %w", err)
				}
			}
			//return newly created entry
            return e.map4, nil
        }
		//Increment IP to search
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

func (t *Translator) Start(ctx context.Context) error {
	//Lock mutex for safety
    t.mu.Lock()
    defer t.mu.Unlock()

	//Write tayga.conf in this directory
    if err := t.writeConfig(); err != nil {
        return fmt.Errorf("failed to write config: %w", err)
    }

	//Write styx.map in this directory
    if err := t.writeMapfile(); err != nil {
        return fmt.Errorf("failed to write mapfile: %w", err)
    }

	//Create command context
    t.cmd = exec.CommandContext(ctx, t.binaryPath, "-c", filepath.Join(t.configPath, "tayga.conf"),"-d")
    t.cmd.Stdout = os.Stdout
    t.cmd.Stderr = os.Stderr

	//start Tayga
    fmt.Printf("starting tayga from: %q\n", t.binaryPath)
    if err := t.cmd.Start(); err != nil {
        return fmt.Errorf("failed to start tayga: %w", err)
    }

    // reap the child, regardless of how it exits.
    done := make(chan struct{})
    go func() {
        defer close(done)
        t.cmd.Wait() // blocks until process exits for any reason
		fmt.Printf("Tayga exited on its own\n")

    }()

    //  handle graceful shutdown on context cancellation.
    go func() {
        select {
        case <-ctx.Done():
            t.cmd.Process.Signal(syscall.SIGTERM)
            <-done // wait for the reaper goroutine to finish
        case <-done:
            // process already exited on its own, nothing to do
        }
    }()

    return nil
}

//get working directory from systemd, or current dir
func configPath() (string,error) {
    if dir, ok := os.LookupEnv("STATE_DIRECTORY"); ok {
        return dir,nil
    }
    return os.Getwd()
}

//write tayga.conf
func (t *Translator) writeConfig() error {
    f, err := os.CreateTemp(t.configPath, ".tayga.conf.tmp")
    if err != nil {
        return fmt.Errorf("failed to create temp config: %w", err)
    }
    tmpPath := f.Name()

	//This config is entirely static - maybe someday we will update
	//to take this from the config file
	fmt.Fprintf(f,"tun-device styx46\n")
	fmt.Fprintf(f,"ipv4-addr 192.0.0.8\n")
	fmt.Fprintf(f,"ipv6-addr 2001:99a:390:2800::470\n")
	//TODO get pref64
	fmt.Fprintf(f,"prefix %s\n","64:ff9b::/96")
	fmt.Fprintf(f,"wkpf-strict no\n")
	fmt.Fprintf(f,"data-dir %s\n",t.configPath)
	fmt.Fprintf(f,"udp-cksum-mode fwd\n")
	fmt.Fprintf(f,"log drop reject icmp self dyn\n")
	fmt.Fprintf(f,"map-file styx.map\n")
	fmt.Fprintf(f,"tun-up yes\n")
	//todo fix these
	fmt.Fprintf(f,"tun-route 0.0.0.0/0\n")
	//todo we also need to do routing for this
	fmt.Fprintf(f,"tun-route 2001:99a:390:2800::460/123\n")
    fmt.Fprintf(f,"map 192.168.55.0/28 2001:99a:390:2800::460/124\n")

    if err := f.Close(); err != nil {
        os.Remove(tmpPath)
        return err
    }

    if err := os.Rename(tmpPath, filepath.Join(t.configPath, "tayga.conf")); err != nil {
        os.Remove(tmpPath)
        return fmt.Errorf("failed to replace config: %w", err)
    }

    return nil
}

//write out the mapfile for Tayga
func (t *Translator) writeMapfile() error {
    f, err := os.CreateTemp(t.configPath, ".styx-*.tmp")
    if err != nil {
        return fmt.Errorf("failed to create temp mapfile: %w\n", err)
    }
    tmpPath := f.Name()

	//print every item in the entry table
    for _, e := range t.entries {
		fmt.Printf("Writing mapfile line for [%s]\n",e)
        fmt.Fprintf(f, "map %s %s\n", e.map4, e.map6)
    }

    if err := f.Close(); err != nil {
        os.Remove(tmpPath)
        return err
    }

    if err := os.Rename(tmpPath,  filepath.Join(t.configPath,"styx.map")); err != nil {
        os.Remove(tmpPath)
        return fmt.Errorf("failed to replace mapfile: %w", err)
    }

    return nil
}
