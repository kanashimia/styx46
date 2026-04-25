package main

import (
    "log"
    "flag"
	"context"
    "github.com/apalrd/styx46/config"
    "github.com/apalrd/styx46/resolver"
    "github.com/apalrd/styx46/server"
    "github.com/apalrd/styx46/translator"
)

func main() {
    configPath := flag.String("conf", "styx.yaml", "path to config file")
    flag.Parse()

    cfg, err := config.Load(*configPath)
    if err != nil {
        log.Fatalf("failed to load config: %v", err)
    }

    //new translator w/ tayga config
    xlate, err := translator.New(cfg.Legacy, cfg.BinaryPath)
    if err != nil {
        log.Fatalf("failed to start translator: %v", err)
    }

    //new resolver w/ upstreams
    res := resolver.New(cfg.Upstreams)

    //new server w/ config, resolver, and translator
    srv := server.New(cfg, res, xlate)

    //start the translator
    if err := xlate.Start(context.Background()); err != nil {
        log.Fatalf("translator error: %v", err)
    }

    //run the server
    if err := srv.ListenAndServe(); err != nil {
        log.Fatalf("server error: %v", err)
    }
}