package quark

import (
	"strings"
)

// Parses a .quark spec file and returns a parsed go Spec struct
func Parse(comSpec string) (spec *Spec, err error) {
  if spec, err := NewParser(strings.NewReader(comSpec)).Parse(); err != nil {
    return nil, err
	} else {
    return spec, nil
  }
}