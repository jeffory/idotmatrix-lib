package command

// Command represents an encodable device command.
type Command interface {
	Encode() ([][]byte, error)
	Chunked() bool
}
