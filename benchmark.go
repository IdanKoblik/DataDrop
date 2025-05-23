package main

import (
	"fmt"
	"time"
)

type BenchmarkStats struct {
	TotalTime       time.Duration
	TransferSpeed   float64
	PacketLoss      float64
	PacketsSent     int
	PacketsReceived int
	TotalBytes      int64
	ChunkTimings    []time.Duration
	MemoryUsage     uint64
	CpuUsage        float64
}

func (b *BenchmarkStats) PrintStats(mode bool) {
	if !mode {
		return
	}

	fmt.Printf("\nBenchmark Results:\n")
	fmt.Printf("  Total Time: %v\n", b.TotalTime)
	fmt.Printf("  Transfer Speed: %.2f bytes/sec\n", b.TransferSpeed)
	fmt.Printf("  Packet Loss: %.2f%%\n", b.PacketLoss)
	fmt.Printf("  Packets Sent: %d\n", b.PacketsSent)
	fmt.Printf("  Packets Received: %d\n", b.PacketsReceived)
	fmt.Printf("  Total Data Transferred: %d bytes\n", b.TotalBytes)
	fmt.Printf("  Average Chunk Time: %v\n", avg(b.ChunkTimings))
	fmt.Printf("  Memory Usage: %d bytes\n", b.MemoryUsage)
	fmt.Printf("  CPU Usage: %.2f%%\n", b.CpuUsage)
}

func avg(durations []time.Duration) time.Duration {
	var total time.Duration
	for _, dur := range durations {
		total += dur
	}

	return total / time.Duration(len(durations))
}
