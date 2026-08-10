package kernel

import _ "embed"

//go:embed payload/kernel.zip
var embeddedZip []byte

// EmbeddedZip returns the go:embed kernel payload used by `arah kernel install`.
func EmbeddedZip() []byte {
	return embeddedZip
}
