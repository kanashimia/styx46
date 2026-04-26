package translator

import (
	"fmt"
	"log"
	"net"
	"os"
	"github.com/apalrd/styx46/config"
	"github.com/vishvananda/netlink"
)


// writeSysctl writes a value to a sysctl path.
func writeSysctl(path, value string) error {
	if err := os.WriteFile(path, []byte(value), 0644); err != nil {
		return fmt.Errorf("writing sysctl %s: %w", path, err)
	}
	return nil
}

// Router configures routing sysctls and proxy-ND entries via Netlink.
func Router(cfg *config.Config) error {
	// Enable routing
	if err := writeSysctl("/proc/sys/net/ipv6/conf/all/forwarding", "1"); err != nil {
		return err
	}	
	if err := writeSysctl("/proc/sys/net/ipv4/conf/all/forwarding", "1"); err != nil {
		return err
	}	

	//enable proxy NDP if configured
	if cfg.IfUp.ProxyND && len(cfg.IfUp.Name) > 0 {
		SysCtl := fmt.Sprintf("/proc/sys/net/ipv6/conf/%s/proxy_ndp",cfg.IfUp.Name)
		if err := writeSysctl(SysCtl, "1"); err != nil {
			return err
		}		
	}

	//write all the proxy ND entries
	if !cfg.IfUp.ProxyND {
		return nil
	}

	// get upstream if in netlink
	upLink, err := netlink.LinkByName(cfg.IfUp.Name)
	if err != nil {
		return fmt.Errorf("looking up interface %q: %w", cfg.IfUp.Name, err)
	}

	// For each downstream, add proxy-ND entries for every IP in the CIDR.
	for _, down := range cfg.IfDowns {
		if cfg.Debug {
			log.Printf("Adding Proxy-ND entries for %s (%s)", down.Name,down.Map6)
		}
		ip, ipNet, err := net.ParseCIDR(down.Map6)
		if err != nil {
			return fmt.Errorf("parsing CIDR %q: %w", down.Map6, err)
		}

		// Iterate every IP in the network (skip network address).
		for cur := cloneIP(ip.Mask(ipNet.Mask)); ipNet.Contains(cur); incrementIP(cur) {
			neigh := &netlink.Neigh{
				LinkIndex: upLink.Attrs().Index,
				IP:        cloneIP(cur),
				Flags:     netlink.NTF_PROXY,
				Family:    netlink.FAMILY_V6,
			}
			if err := netlink.NeighAdd(neigh); err != nil {
				return fmt.Errorf("adding proxy-ND for %s on %s: %w", cur, cfg.IfUp.Name, err)
			}
		}
	}

	return nil
}

// incrementIP increments an IP address in-place.
func incrementIP(ip net.IP) {
	for i := len(ip) - 1; i >= 0; i-- {
		ip[i]++
		if ip[i] != 0 {
			break
		}
	}
}