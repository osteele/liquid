package expressions

// These strings match lexer tokens.
const (
	AssignStatementSelector = "%assign "
	CycleStatementSelector  = "{%cycle "
	LoopStatementSelector   = "%loop "
	WhenStatementSelector   = "{%when "
)

// A Statement is the result of parsing a string.
type Statement struct{ parseValue }

// Expression returns a statement's expression function.
// func (s *Statement) Expression() Expression { return &expression{s.val} }

// An Assignment is a parse of an {% assign %} statement
type Assignment struct {
	Variable string
	ValueFn  Expression
}

// A Cycle is a parse of an {% assign %} statement
type Cycle struct {
	Group  string
	Values []string
}

// A Loop is a parse of a {% loop %} statement
type Loop struct {
	Variable string
	Expr     Expression
	loopModifiers
}

type loopModifiers struct {
	Limit    *int
	Offset   int
	Reversed bool
	Cols     int
}

// A When is a parse of a {% when %} clause
type When struct {
	Exprs []Expression
}

// ParseStatement parses an statement into an Expression that can evaluated to return a
// structure specific to the statement.
func ParseStatement(sel, source string) (*Statement, error) {
	p, err := parse(sel + source)
	if err != nil {
		return nil, err
	}
	return &Statement{*p}, nil
}
