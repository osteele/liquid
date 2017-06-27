package errors

import "fmt"

// UndefinedFilter is an error that the named filter is not defined.
type UndefinedFilter string

func (e UndefinedFilter) Error() string {
	return fmt.Sprintf("undefined filter: %s", string(e))
}
