package resources

import (
	"bytes"
	"embed"
	"io"
)

//go:embed *
var f embed.FS

func LoadResourceFile(filePath string) (io.Reader, error) {
	_bytes, err := f.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	return bytes.NewReader(_bytes), nil
}
