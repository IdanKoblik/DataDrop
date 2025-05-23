package internals

import (
	"crypto/md5"
	"echo/fileproto"
	"encoding/binary"
	"fmt"
	"io"

	"google.golang.org/protobuf/proto"
)

const maxRetries = 5

type Chunk struct {
	Index int
	Data  []byte
}

func SendPacket(stream io.Writer, msg proto.Message) error {
	data, err := proto.Marshal(msg)
	if err != nil {
		return err
	}

	if err := binary.Write(stream, binary.LittleEndian, uint32(len(data))); err != nil {
		return err
	}

	if _, err := stream.Write(data); err != nil {
		return err
	}

	return nil
}

func ReceivePacket(stream io.Reader, msg proto.Message) error {
	var msgLen uint32
	if err := binary.Read(stream, binary.LittleEndian, &msgLen); err != nil {
		return err
	}

	if msgLen > 10*1024*1024 {
		return fmt.Errorf("message too large: %d bytes", msgLen)
	}

	data := make([]byte, msgLen)
	if _, err := io.ReadFull(stream, data); err != nil {
		return err
	}

	if err := proto.Unmarshal(data, msg); err != nil {
		return err
	}

	return nil
}

func CreateFileChunk(version uint32, filename string, chunkIndex, totalChunks uint32, data []byte) *fileproto.FileChunk {
	hash := md5.Sum(data)
	checksum := fmt.Sprintf("%x", hash)

	return &fileproto.FileChunk{
		Version:     version,
		Filename:    filename,
		ChunkIndex:  chunkIndex,
		TotalChunks: totalChunks,
		Data:        data,
		Checksum:    checksum,
	}
}

func ValidateChecksum(chunk *fileproto.FileChunk) bool {
	hash := md5.Sum(chunk.Data)
	expectedChecksum := fmt.Sprintf("%x", hash)
	return expectedChecksum == chunk.Checksum
}
