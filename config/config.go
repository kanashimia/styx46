package config

import (
    "os"
    "fmt"
    "net"
    "gopkg.in/yaml.v3"
)

type Config struct {
    Listen    	string   `yaml:"listen"`     // clients connect to us e.g. "0.0.0.0:53"
    Upstreams 	[]string `yaml:"upstreams"`  // we connect upstream e.g. ["[2620:fe::fe]:53", "9.9.9.9:53"]
	Legacy		string	 `yaml:"legacy"`	 // legacy network space reserved for this, i.e. "10.0.0.0/8"
	BinaryPath	string	 `yaml:"binary_path"` // Path or binary file name of the translator (default "tayga")
	PreferSynth	bool	 `yaml:"prefer_synth"` // Prefer to synthesize addresses if AAAA records are available
	Pref64		string	 `yaml:"pref64"` 	 // Pref64 for the PLAT
    Debug       bool     `yaml:"debug"`      // Debug-level query logging
    Pool        *net.IPNet                   // decoded from Legacy
}

func Load(path string) (*Config, error) {
    f, err := os.Open(path)
    if err != nil {
        return nil, err
    }
    defer f.Close()

    var cfg Config
    if err := yaml.NewDecoder(f).Decode(&cfg); err != nil {
        return nil, err
    }

    //validate ip net
    ip, network, err := net.ParseCIDR(cfg.Legacy)
    if err != nil {
        return nil, fmt.Errorf("invalid CIDR: %w", err)
    }
    if ip.Equal(network.IP) == false {
        return nil, fmt.Errorf("CIDR must be a network address, not a host address")
    }
    cfg.Pool = network
    return &cfg, nil
}