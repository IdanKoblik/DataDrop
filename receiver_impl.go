package main

import (
	"context"
	"echo/fileproto"
	"echo/internals"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

func Receive(localAddr string, benchmark bool) error {
	qm, err := InitServer(localAddr)
	if err != nil {
		return err
	}

	defer qm.Close()

	conn, err := qm.AcceptConnection(context.Background())
	if err != nil {
		return err
	}

	defer conn.CloseWithError(0, "transfer complete")

	stream, err := conn.AcceptStream(context.Background())
	if err != nil {
		return err
	}

	defer stream.Close()

	start := time.Now()
	stats := &BenchmarkStats{}

	var outputFile *os.File
	var expectedFilename string
	var totalChunks uint32
	chunks := make(map[uint32][]byte)

	for {
		var chunk fileproto.FileChunk
		if err := internals.ReceivePacket(stream, &chunk); err != nil {
			return err
		}

		chunkStart := time.Now()
		if chunk.Version != VERSION {
			return fmt.Errorf("Version mismatch, expected: %d, actual: %d", VERSION, chunk.Version)
		}

		if chunk.ChunkIndex == 0 && chunk.Checksum == "eof" {
			totalChunks = chunk.TotalChunks
			fmt.Printf("Received EOF marker. Expected %d chunks total.\n", totalChunks)
			break
		}

		if !internals.ValidateChecksum(&chunk) {
			return fmt.Errorf("checksum validation failed for chunk %d", chunk.ChunkIndex)
		}

		if outputFile == nil {
			expectedFilename = chunk.Filename
			homeDir, err := os.UserHomeDir()
			if err != nil {
				return err
			}

			filePath := filepath.Join(homeDir, expectedFilename)
			outputFile, err = os.Create(filePath)
			if err != nil {
				return fmt.Errorf("failed to create file: %w", err)
			}
			defer outputFile.Close()

			fmt.Printf("Receiving file: %s\n", expectedFilename)
		}

		if chunk.Filename != expectedFilename {
			return fmt.Errorf("filename mismatch: expected %s, got %s",
				expectedFilename, chunk.Filename)
		}

		// Store chunk data
		chunks[chunk.ChunkIndex] = chunk.Data
		stats.ChunkTimings = append(stats.ChunkTimings, time.Since(chunkStart))
		stats.PacketsReceived++
		stats.TotalBytes += int64(len(chunk.Data))

		// Progress indicator
		if len(chunks)%100 == 0 {
			fmt.Printf("Received %d chunks...\n", len(chunks))
		}
	}

	fmt.Println("Writing chunks to file...")
	for i := uint32(1); i <= totalChunks; i++ {
		data, exists := chunks[i]
		if !exists {
			return fmt.Errorf("missing chunk at index %d", i)
		}

		if _, err := outputFile.Write(data); err != nil {
			return fmt.Errorf("failed to write chunk %d: %w", i, err)
		}
	}

	duration := time.Since(start)
	stats.TotalTime = duration
	stats.CpuUsage = getCpuUsage()
	stats.MemoryUsage = GetMemoryUsage()
	stats.PrintStats(benchmark)

	homeDir, _ := os.UserHomeDir()
	filePath := filepath.Join(homeDir, expectedFilename)
	fmt.Printf("File received successfully: %s (%d bytes) in %.2fs\n",
		filePath, stats.TotalBytes, duration.Seconds())

	return nil
}
