package config

import (
    "os"
    "gopkg.in/yaml.v3"
)

type Config struct {
    Listen    	string   `yaml:"listen"`     // clients connect to us e.g. "0.0.0.0:53"
    Upstreams 	[]string `yaml:"upstreams"`  // we connect upstream e.g. ["[2620:fe::fe]:53", "9.9.9.9:53"]
	Legacy		string	 `yaml:"legacy"`	 // legacy network space reserved for this, i.e. "10.0.0.0/8"
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
    return &cfg, nil
}