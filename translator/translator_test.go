
// Test Cases for translator
package translator

import (
	"context"
	"net"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// mockBinary writes a simple shell script that acts as a stand-in for the
// tayga binary — it just sleeps until signalled so we can test
// Start/Stop behaviour without a real binary.
func mockBinary(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "mock-tayga")
	err := os.WriteFile(path, []byte("#!/bin/sh\ntrap exit HUP\nwhile true; do sleep 0.1; done\n"), 0755)
	if err != nil {
		t.Fatalf("failed to write mock binary: %v", err)
	}
	return path
}

func TestNew_ValidCIDR(t *testing.T) {
	tr, err := New("100.64.0.0/24", mockBinary(t))
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if tr == nil {
		t.Fatal("expected non-nil Translator")
	}
}

func TestNew_InvalidCIDR(t *testing.T) {
	_, err := New("not-a-cidr", mockBinary(t))
	if err == nil {
		t.Fatal("expected error for invalid CIDR, got nil")
	}
}

func TestLookup_AllocatesNewEntry(t *testing.T) {
	tr, err := New("100.64.0.0/24", mockBinary(t))
	if err != nil {
		t.Fatalf("failed to create translator: %v", err)
	}

	map6 := net.ParseIP("2001:db8::1")
	map4, err := tr.Lookup(map6)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if map4 == nil {
		t.Fatal("expected a map4 address, got nil")
	}
	if !tr.pool.Contains(map4) {
		t.Errorf("allocated address %s is outside pool", map4)
	}
}

func TestLookup_AllocatesExistingEntry(t *testing.T) {
	tr, err := New("100.64.0.0/24", mockBinary(t))
	if err != nil {
		t.Fatalf("failed to create translator: %v", err)
	}

	map6 := net.ParseIP("2001:db8::1")
	first, err := tr.Lookup(map6)
	if err != nil {
		t.Fatalf("first lookup failed: %v", err)
	}
	second, err := tr.Lookup(map6)
	if err != nil {
		t.Fatalf("second lookup failed: %v", err)
	}
	if !first.Equal(second) {
		t.Errorf("expected same map4 for same map6, got %s then %s", first, second)
	}
}

func TestLookup_AllocatesMultipleAddresses(t *testing.T) {
	tr, err := New("100.64.0.0/24", mockBinary(t))
	if err != nil {
		t.Fatalf("failed to create translator: %v", err)
	}

	map6a := net.ParseIP("2001:db8::1")
	map6b := net.ParseIP("2001:db8::2")

	map4a, err := tr.Lookup(map6a)
	if err != nil {
		t.Fatalf("lookup a failed: %v", err)
	}
	map4b, err := tr.Lookup(map6b)
	if err != nil {
		t.Fatalf("lookup b failed: %v", err)
	}
	if map4a.Equal(map4b) {
		t.Errorf("expected different map4 addresses, both got %s", map4a)
	}
}

func TestLookup_PoolExhausted(t *testing.T) {
	// /30 gives 4 addresses
	tr, err := New("100.64.0.0/30", mockBinary(t))
	if err != nil {
		t.Fatalf("failed to create translator: %v", err)
	}

	// Exhaust the pool
	i := 0
	for {
		ip := net.ParseIP("2001:db8::1") // reuse base, increment last octet
		ip[15] = byte(i + 1)
		_, err := tr.Lookup(ip)
		if err != nil {
			break
		}
		i++
		if i > 10 {
			t.Fatal("pool should have been exhausted by now")
		}
	}
}


func TestStart_LaunchesAndStopsProcess(t *testing.T) {
	tr, err := New("100.64.0.0/24", mockBinary(t))
	if err != nil {
		t.Fatalf("failed to create translator: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := tr.Start(ctx); err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	if tr.cmd == nil || tr.cmd.Process == nil {
		t.Fatal("expected process to be running after Start")
	}

	cancel()
	time.Sleep(100 * time.Millisecond)

	if tr.cmd.ProcessState == nil {
		t.Error("expected process to have exited after context cancel")
	}
}

func TestStart_WritesConfigFiles(t *testing.T) {
	tr, err := New("100.64.0.0/24", mockBinary(t))
	if err != nil {
		t.Fatalf("failed to create translator: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := tr.Start(ctx); err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	dir := tr.configPath
	for _, name := range []string{"tayga.conf", "styx.map"} {
		path := filepath.Join(dir, name)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("expected %s to exist after Start", name)
		}
	}
}