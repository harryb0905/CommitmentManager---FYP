package quark

import (
  "fmt"
  "io"
)

// Spec represents a commitment specification
type Spec struct {
  Constraint     *Constraint
  CreateEvent    *Event
  DetachEvent    *Event
  DischargeEvent *Event
}

// A constraint consists of the spec name and who is involved (debtor + creditor)
type Constraint struct {
  Name      string
  Debtor    string
  Creditor  string
}

// An event (such as Offer, Pay) + argument list
type Event struct {
  Name      string
  Args      []Arg
}

// Data field inside the event paramater list
type Arg struct {
  Name   string
  Value  string
}

// Parser represents a parser.
type Parser struct {
  s   *Scanner
  buf struct {
    tok Token  // last read token
    lit string // last read literal
    n   int    // buffer size (max=1)
  }
}

// Adds an argument to the Args slice in the Event struct
func (event *Event) AddArg(arg Arg) []Arg {
  event.Args = append(event.Args, arg)
  return event.Args
}

// Parse parses a spec
func (p *Parser) Parse() (*Spec, error) {
  com := &Spec{}

  // First token should be the "spec" keyword.
  if tok, lit := p.scanIgnoreWhitespace(); tok != SPEC {
    return nil, fmt.Errorf("found %q, expected 'spec'", lit)
  }

  // Get spec name
  com.Constraint = &Constraint{}
  tok, lit := p.scanIgnoreWhitespace();
  if tok == IDENT {
    com.Constraint.Name = lit
  } else {
    return nil, fmt.Errorf("found %q, expected specification name", lit)
  }

  // Get Debtor/From identifier
  if tok, lit := p.scanIgnoreWhitespace(); tok == IDENT {
    com.Constraint.Debtor = lit
  } else {
    return nil, fmt.Errorf("found %q, expected debtor identifier", lit)
  }

  // Next we should see the "TO" keyword.
  if tok, lit := p.scanIgnoreWhitespace(); tok != TO {
    return nil, fmt.Errorf("found %q, expected 'to'", lit)
  }

  // Get Creditor/To identifier
  if tok, lit := p.scanIgnoreWhitespace(); tok == IDENT {
    com.Constraint.Creditor = lit
  } else {
    return nil, fmt.Errorf("found %q, expected creditor identifier", lit)
  }

  // Obtain 'create' statement + args
  com.CreateEvent = &Event{}
  if err := NewEvent(CREATE, com.CreateEvent, p); err != nil {
    return nil, err
  }

  // Obtain 'detach' statement + args + deadline
  com.DetachEvent = &Event{}
  if err := NewEvent(DETACH, com.DetachEvent, p); err != nil {
    return nil, err
  }
  if err := GetDeadline(com.DetachEvent, p); err != nil {
    return nil, err
  }

  // Obtain 'discharge' statement + args + deadline
  com.DischargeEvent = &Event{}
  if err := NewEvent(DISCHARGE, com.DischargeEvent, p); err != nil {
    return nil, err
  }
  if err := GetDeadline(com.DischargeEvent, p); err != nil {
    return nil, err
  }

  // Return the successfully parsed statement
  return com, nil
}

// Parses an event found in the spec source code
func NewEvent(evname Token, event *Event, p *Parser) (error) {
  if tok, lit := p.scanIgnoreWhitespace(); tok == evname {
    tok_ev, lit_ev := p.scanIgnoreWhitespace();
    if tok_ev == IDENT {
      event.Name = lit_ev
    } else {
      return fmt.Errorf("found %q, expected event name for '%s'", lit_ev, evname)
    }
  } else {
    return fmt.Errorf("found %q, expected EVENT", lit)
  }
  // Get arguments (optional) for event fields
  if err := GetArgs(event, p); err != nil {
    return err
  }
  return nil
}

// Gets and parses the event argument list
func GetArgs(event *Event, p *Parser) (error) {
  // Detect left bracket to get arguments
  if tok, lit := p.scanIgnoreWhitespace(); tok != LBRACKET {
    return fmt.Errorf("found %q, expected '['", lit)
  }
  // Loop over all our comma-delimited fields for this event
  for {
    // Read a field + add arg name
    tok, lit := p.scanIgnoreWhitespace()
    if tok != IDENT {
      return fmt.Errorf("found %q, expected field", lit)
    }
    event.AddArg(Arg{
      Name: lit,
    })

    // Detect close bracket
    if tok, _ := p.scanIgnoreWhitespace(); tok == RBRACKET {
      break
    } else {
      p.unscan()
    }

    // Detect another arg
    if tok, lit := p.scanIgnoreWhitespace(); tok == COMMA {
      continue
    } else {
      return fmt.Errorf("found %q, expected ',' or ']'", lit)
    }
  }
  return nil
}

// Obtains the deadline value associated with the detach and discharge clauses
func GetDeadline(event *Event, p *Parser) (error) {
  tok, lit := p.scanIgnoreWhitespace()
  if tok != IDENT {
    return fmt.Errorf("found %q, expected field", lit)
  }
  // Detect possible value associated with arg
  tok_eq, _ := p.scanIgnoreWhitespace()
  if tok_eq == EQUALS {
    tok_val, lit_val := p.scanIgnoreWhitespace()
    if tok_val == IDENT {
      // Add arg name with associated arg value
      event.AddArg(Arg{
        Name: lit,
        Value: lit_val,
      })
    } else {
      return fmt.Errorf("found %q, expected value for %q when using '='", lit_val, lit)
    }
  }
  return nil
}

// NewParser returns a new instance of Parser.
func NewParser(r io.Reader) *Parser {
  return &Parser{s: NewScanner(r)}
}

// scan returns the next token from the underlying scanner.
// If a token has been unscanned then read that instead.
func (p *Parser) scan() (tok Token, lit string) {
  // If we have a token on the buffer, then return it.
  if p.buf.n != 0 {
    p.buf.n = 0
    return p.buf.tok, p.buf.lit
  }

  // Otherwise read the next token from the scanner.
  tok, lit = p.s.Scan()

  // Save it to the buffer in case we unscan later.
  p.buf.tok, p.buf.lit = tok, lit
  return
}

// scanIgnoreWhitespace scans the next non-whitespace token.
func (p *Parser) scanIgnoreWhitespace() (tok Token, lit string) {
  tok, lit = p.scan()
  if tok == WS {
    tok, lit = p.scan()
  }
  return
}

// unscan pushes the previously read token back onto the buffer.
func (p *Parser) unscan() {
  p.buf.n = 1
}
