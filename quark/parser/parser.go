package quark

import (
  "fmt"
  "io"
)

// Spec represents a specification of a contract
type Spec struct {
  Constraint     *Constraint
  CreateEvent    *Event
  DetachEvent    *Event
  DischargeEvent *Event
}

// A constraint details who is involved in the specification and the spec name
type Constraint struct {
  Name      string
  Debter    string
  Creditor  string
}

// An event (such as Offer, Pay)
type Event struct {
  Name  string
  Args  []Arg
}

// Data field inside the event paramater list
type Arg struct {
  Name  string
  Value string
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
func (p *Parser) Parse() (spec *Spec, error) {
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

  // Get Debter/From name
  if tok, lit := p.scanIgnoreWhitespace(); tok == IDENT {
    com.Constraint.Debter = lit
  } else {
    return nil, fmt.Errorf("found %q, expected debter name", lit)
  }

  // Next we should see the "TO" keyword.
  if tok, lit := p.scanIgnoreWhitespace(); tok != TO {
    return nil, fmt.Errorf("found %q, expected 'to'", lit)
  }

  // Get Creditor/To name
  if tok, lit := p.scanIgnoreWhitespace(); tok == IDENT {
    com.Constraint.Creditor = lit
  } else {
    return nil, fmt.Errorf("found %q, expected creditor name", lit)
  }

  // Obtain 'create' statement + args
  com.CreateEvent = &Event{}
  if err := NewEvent(CREATE, com.CreateEvent, p); err != nil {
    return nil, err
  }

  // Obtain 'detach' statement + args
  com.DetachEvent = &Event{}
  if err := NewEvent(DETACH, com.DetachEvent, p); err != nil {
    return nil, err
  }

  // Obtain 'discharge' statement + args
  com.DischargeEvent = &Event{}
  if err := NewEvent(DISCHARGE, com.DischargeEvent, p); err != nil {
    return nil, err
  }

  // Return the successfully parsed statement.
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
    return fmt.Errorf("found %q, expected '%s'", lit, evname)
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
  // Next we should loop over all our comma-delimited fields for this event
  for {
    // Read a field.
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
    } else {
      // Add just arg name (no given value)
      event.AddArg(Arg{
        Name: lit,
      })
      p.unscan()
    }

    // Detect close bracket
    if tok, _ := p.scanIgnoreWhitespace(); tok == RBRACKET {
      break
    } else {
      p.unscan()
    }

    if tok, lit := p.scanIgnoreWhitespace(); tok == COMMA {
      continue
    } else {
      return fmt.Errorf("found %q, expected ',' or ']'", lit)
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
func (p *Parser) unscan() { p.buf.n = 1 }
