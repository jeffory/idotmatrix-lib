package device

import (
	"testing"

	"github.com/jeffory/idotmatrix-lib/command"
	"github.com/jeffory/idotmatrix-lib/protocol"
)

type mockConnection struct {
	packets [][]byte
	chunks  [][]byte
}

func (m *mockConnection) WritePacket(data []byte) error {
	cp := make([]byte, len(data))
	copy(cp, data)
	m.packets = append(m.packets, cp)
	return nil
}

func (m *mockConnection) WriteChunked(chunks [][]byte) error {
	for _, c := range chunks {
		cp := make([]byte, len(c))
		copy(cp, c)
		m.chunks = append(m.chunks, cp)
	}
	return nil
}

func (m *mockConnection) Close() error { return nil }

func TestDeviceSendSimpleCommand(t *testing.T) {
	mock := &mockConnection{}
	dev := NewWithWriter(mock, protocol.Display32x32)

	err := dev.Send(command.PowerOn())
	if err != nil {
		t.Fatalf("Send() error: %v", err)
	}

	if len(mock.packets) != 1 {
		t.Errorf("got %d packets, want 1", len(mock.packets))
	}
	if len(mock.chunks) != 0 {
		t.Errorf("got %d chunk calls, want 0", len(mock.chunks))
	}
}

func TestDeviceSendChunkedCommand(t *testing.T) {
	mock := &mockConnection{}
	dev := NewWithWriter(mock, protocol.Display32x32)

	cmd := command.NewText("A",
		command.WithDisplaySize(protocol.Display32x32),
	)

	err := dev.Send(cmd)
	if err != nil {
		t.Fatalf("Send() error: %v", err)
	}

	if len(mock.chunks) == 0 {
		t.Error("expected chunked write for text command")
	}
}
