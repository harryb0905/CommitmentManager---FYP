package quark

// Token represents a lexical token
type Token int

const (
  // Special tokens
  ILLEGAL Token = iota
  EOF
  WS

  // Literals
  IDENT

  // Misc characters
  LBRACKET // [
  RBRACKET // ]
  EQUALS   // =
  COMMA    // ,

  // Keywords
  SPEC
  TO
  CREATE
  DETACH
  DISCHARGE
)
