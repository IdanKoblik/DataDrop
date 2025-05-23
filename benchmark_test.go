package main

import (
	"bytes"
	"os"
	"testing"
	"time"
)

func TestPrintStats(t *testing.T) {
	stats := &BenchmarkStats{
		TotalTime:       2 * time.Second,
		TransferSpeed:   1024.5,
		PacketLoss:      1.5,
		PacketsSent:     100,
		PacketsReceived: 98,
		TotalBytes:      1048576,
		ChunkTimings:    []time.Duration{time.Millisecond * 10, time.Millisecond * 20},
		MemoryUsage:     2048000,
		CpuUsage:        55.5,
	}

	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	stats.PrintStats(true)

	w.Close()
	var buf bytes.Buffer
	_, err := buf.ReadFrom(r)
	if err != nil {
		t.Fatalf("Failed to read from buffer: %v", err)
	}
	os.Stdout = old

	output := buf.String()
	if !containsAll(output,
		"Benchmark Results:",
		"Total Time: 2s",
		"Transfer Speed: 1024.50 bytes/sec",
		"Packet Loss: 1.50%",
		"Packets Sent: 100",
		"Packets Received: 98",
		"Total Data Transferred: 1048576 bytes",
		"Average Chunk Time: 15ms",
		"Memory Usage: 2048000 bytes",
		"CPU Usage: 55.50%",
	) {
		t.Errorf("Output did not contain expected benchmark results:\n%s", output)
	}

	r, w, _ = os.Pipe()
	os.Stdout = w

	stats.PrintStats(false)

	w.Close()
	buf.Reset()
	_, err = buf.ReadFrom(r)
	if err != nil {
		t.Fatalf("Failed to read from buffer: %v", err)
	}
	os.Stdout = old

	if buf.String() != "" {
		t.Errorf("Expected no output when mode is false, got:\n%s", buf.String())
	}
}

func containsAll(output string, substrings ...string) bool {
	for _, substr := range substrings {
		if !bytes.Contains([]byte(output), []byte(substr)) {
			return false
		}
	}
	return true
}
