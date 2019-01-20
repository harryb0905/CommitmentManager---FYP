package quark

import (
	"strings"
  "github.com/scc300/scc300-network/quark/parser"
)

// Parses a .quark spec file and returns a parsed go Spec struct
func Parse(comSpec string) (spec *quark.Spec, err error) {
  if spec, err := quark.NewParser(strings.NewReader(comSpec)).Parse(); err != nil {
		return nil, err
	} else {
		return spec, nil
	}
}