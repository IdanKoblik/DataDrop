package main

import (
	"fmt"
	"time"

	"github.com/showwin/speedtest-go/speedtest"
)

func MeasureUpload() (float64, float64, error) {
	serverList, err := speedtest.FetchServers()
	if err != nil {
		return -1, -1, err
	}

	targets, err := serverList.FindServer([]int{})
	if err != nil {
		return -1, -1, err
	}
	if len(targets) == 0 {
		return -1, -1, err
	}

	server := targets[0]

	err = server.UploadTest()
	if err != nil {
		return -1, -1, err
	}

	var rtt float64
	server.PingTest(func(latency time.Duration) { rtt = float64(latency.Milliseconds()) })
	upload := float64(server.ULSpeed) / 1_000_000 * 8
	fmt.Printf("Upload speed: %.2f Mbps\nRtt: %.2f ms\n", upload, rtt)
	return upload, rtt, nil
}
