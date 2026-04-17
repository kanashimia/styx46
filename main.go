package main

import (
    "log"
    "flag"
    "github.com/apalrd/styx46/config"
    "github.com/apalrd/styx46/resolver"
    "github.com/apalrd/styx46/server"
)

func main() {
    configPath := flag.String("conf", "styx.yaml", "path to config file")
    flag.Parse()

    cfg, err := config.Load(*configPath)
    if err != nil {
        log.Fatalf("failed to load config: %v", err)
    }

    res := resolver.New(cfg.Upstreams)
    srv := server.New(cfg.Listen, res)

    if err := srv.ListenAndServe(); err != nil {
        log.Fatalf("server error: %v", err)
    }
}