package quark

import (
	"strings"
  "github.com/scc300/scc300-network/quark/parser"
)

// Parses a .quark spec file and returns a parsed go Spec struct
func Parse(comSpec string) (spec *quark.Spec, err string) {
  if spec, err := quark.NewParser(strings.NewReader(comSpec)).Parse(); err != nil {
		return nil, errstring(err)
	} else {
		return spec, "nil"
	}
}

// errstring returns the string representation of an error.
func errstring(err error) string {
	if err != nil {
		return err.Error()
	}
	return ""
}
