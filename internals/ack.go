package internals

import (
	"echo/fileproto"
	"sync"

	"github.com/quic-go/quic-go"
	"google.golang.org/protobuf/proto"
)

type AckManager struct {
	mu   sync.Mutex
	acks map[uint32]chan struct{}
}

func NewAckManager() *AckManager {
	return &AckManager{acks: make(map[uint32]chan struct{})}
}

func (am *AckManager) Register(index uint32) chan struct{} {
	am.mu.Lock()
	defer am.mu.Unlock()
	ch := make(chan struct{}, 1)
	am.acks[index] = ch
	return ch
}

func (am *AckManager) Notify(index uint32) {
	am.mu.Lock()
	defer am.mu.Unlock()
	if ch, ok := am.acks[index]; ok {
		ch <- struct{}{}
		close(ch)
		delete(am.acks, index)
	}
}

func (am *AckManager) Listen(stream quic.Stream) {
	buffer := make([]byte, 128)
	for {
		n, err := stream.Read(buffer)
		if err != nil {
			continue
		}

		var ack fileproto.FileAck
		if err := proto.Unmarshal(buffer[:n], &ack); err == nil {
			am.Notify(ack.ChunkIndex)
		}
	}
}
