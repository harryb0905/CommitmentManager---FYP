package quark

// Token represents a lexical token.
type Token int

const (
  // Special tokens
  ILLEGAL Token = iota
  EOF
  WS

  // Literals
  IDENT // main

  // Misc characters
  LBRACE   // {
  RBRACE   // }
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
