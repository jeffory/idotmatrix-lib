package command

import "github.com/jeffory/idotmatrix-lib/protocol"

type powerOnCmd struct{}

func PowerOn() Command                        { return &powerOnCmd{} }
func (*powerOnCmd) Chunked() bool             { return false }
func (*powerOnCmd) Encode() ([][]byte, error) { return [][]byte{protocol.EncodePowerOn()}, nil }

type powerOffCmd struct{}

func PowerOff() Command                        { return &powerOffCmd{} }
func (*powerOffCmd) Chunked() bool             { return false }
func (*powerOffCmd) Encode() ([][]byte, error) { return [][]byte{protocol.EncodePowerOff()}, nil }

type resetCmd struct{}

func Reset() Command                        { return &resetCmd{} }
func (*resetCmd) Chunked() bool             { return false }
func (*resetCmd) Encode() ([][]byte, error) {
	p1, p2 := protocol.EncodeReset()
	return [][]byte{p1, p2}, nil
}
