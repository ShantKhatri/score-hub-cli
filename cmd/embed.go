package cmd

import (
	_ "embed"
)

//go:embed index.yaml
var indexYAML []byte

func getEmbeddedIndex() []byte {
	return indexYAML
}
