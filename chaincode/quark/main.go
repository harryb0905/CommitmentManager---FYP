package quark

import (
	"strings"
)

// Parses a commitment specification and returns a parsed spec as a Go Struct
func Parse(comSpec string) (spec *Spec, err error) {
  if spec, err := NewParser(strings.NewReader(comSpec)).Parse(); err != nil {
    return nil, err
	} else {
    return spec, nil
  }
}