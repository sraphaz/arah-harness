// Package arahharness expõe metadados do repositório compartilhados pelos
// binários Go. A versão vem do arquivo VERSION na raiz — fonte única de
// verdade, embutida em build via go:embed (sem ldflags).
package arahharness

import (
	_ "embed"
	"strings"
)

//go:embed VERSION
var rawVersion string

// Version é a versão do ARAH Harness lida do arquivo VERSION.
var Version = strings.TrimSpace(rawVersion)
