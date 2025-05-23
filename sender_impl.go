package main

import (
	"context"
	"echo/fileproto"
	"echo/internals"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

func Send(filename, remoteAddr string, benchmark bool) error {
	start := time.Now()
	file, err := os.Open(filename)
	if err != nil {
		return err
	}

	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		return err
	}

	qm, err := InitClient(remoteAddr)
	if err != nil {
		return err
	}

	defer qm.Close()

	stream, err := qm.GetConnection().OpenStreamSync(context.Background())
	if err != nil {
		return err
	}

	stats := &BenchmarkStats{}
	const chunkSize = 64 * 1024
	totalChunks := uint32((fileInfo.Size() + chunkSize - 1) / chunkSize)
	baseFilename := filepath.Base(filename)

	fmt.Printf("Sending file: %s (%d bytes) in %d chunks\n", baseFilename, fileInfo.Size(), totalChunks)
	buffer := make([]byte, chunkSize)
	var chunkIndex uint32 = 1

	for {
		chunkStart := time.Now()
		n, err := file.Read(buffer)
		if err == io.EOF {
			break
		}

		if err != nil {
			return err
		}

		chunk := internals.CreateFileChunk(
			VERSION,
			baseFilename,
			chunkIndex,
			totalChunks,
			buffer[:n],
		)

		if err := internals.SendPacket(stream, chunk); err != nil {
			return err
		}

		stats.ChunkTimings = append(stats.ChunkTimings, time.Since(chunkStart))
		stats.PacketsSent++
		stats.TotalBytes += int64(n)

		chunkIndex++

		if chunkIndex%100 == 0 {
			progress := float64(chunkIndex) / float64(totalChunks) * 100
			fmt.Printf("Progress: %.1f%% (%d/%d chunks)\n", progress, chunkIndex-1, totalChunks)
		}
	}

	eofChunk := &fileproto.FileChunk{
		Version:     VERSION,
		Filename:    baseFilename,
		ChunkIndex:  0, // Special index for EOF
		TotalChunks: totalChunks,
		Data:        []byte{},
		Checksum:    "eof",
	}

	if err := internals.SendPacket(stream, eofChunk); err != nil {
		return fmt.Errorf("failed to send EOF marker: %w", err)
	}

	duration := time.Since(start)
	stats.TotalTime = duration
	stats.TransferSpeed = float64(stats.TotalBytes) / duration.Seconds()
	stats.MemoryUsage = GetMemoryUsage()
	stats.CpuUsage = getCpuUsage()
	stats.PrintStats(benchmark)

	fmt.Printf("File sent successfully: %s (%d bytes) in %.2fs\n",
		filename, stats.TotalBytes, duration.Seconds())

	return nil
}
