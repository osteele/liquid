package errors

import "fmt"

type UndefinedFilter string

func (e UndefinedFilter) Error() string {
	return fmt.Sprintf("undefined filter: %s", string(e))
}
