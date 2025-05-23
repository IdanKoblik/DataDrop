package internals_test

import (
	"echo/internals"
	"testing"
	"time"
)

func TestAckManager_RegisterAndNotify(t *testing.T) {
	am := internals.NewAckManager()
	index := uint32(42)

	ch := am.Register(index)
	if ch == nil {
		t.Fatal("expected channel, got nil")
	}

	am.Notify(index)

	select {
	case _, ok := <-ch:
		if !ok {
			// Channel closed, good
		} else {
			t.Error("expected channel to be closed after notification")
		}
	case <-time.After(time.Second):
		t.Error("timeout waiting for notification")
	}
}

/*
func TestAckManager_Listen(t *testing.T) {
	am := internals.NewAckManager()
	index := uint32(7)
	ch := am.Register(index)

	addr := &net.UDPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 0}
	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		t.Fatalf("failed to start UDP listener: %v", err)
	}
	defer conn.Close()

	go am.Listen(conn)

	udpAddr := conn.LocalAddr().(*net.UDPAddr)
	ack := &fileproto.FileAck{
		ChunkIndex: index,
	}

	data, err := proto.Marshal(ack)
	if err != nil {
		t.Fatalf("failed to marshal proto: %v", err)
	}

	_, err = conn.WriteToUDP(data, udpAddr)
	if err != nil {
		t.Fatalf("failed to send udp packet: %v", err)
	}

	select {
	case _, ok := <-ch:
		if !ok {
			// Channel closed, good
		} else {
			t.Error("expected channel to be closed after notification")
		}
	case <-time.After(2 * time.Second):
		t.Error("timeout waiting for ack notification")
	}
}
*/
