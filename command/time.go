package command

import (
	"time"

	"github.com/jeffory/idotmatrix-lib/protocol"
)

type syncTimeCmd struct{ t time.Time }

func SyncTime(t time.Time) Command              { return &syncTimeCmd{t: t} }
func (*syncTimeCmd) Chunked() bool              { return false }
func (c *syncTimeCmd) Encode() ([][]byte, error) { return [][]byte{protocol.EncodeTimeSync(c.t)}, nil }
