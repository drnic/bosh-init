package errors

import (
	"fmt"
)

type ExplainableError struct {
	header string
	errors []error
}

func NewExplainableError(header string) *ExplainableError {
	return &ExplainableError{header: header}
}

func (e *ExplainableError) Error() string {
	output := e.header

	for _, err := range e.errors {
		if output != "" {
			output += "\n"
		}

		output += fmt.Sprintf("* %s", err.Error())
	}

	return output
}

func (e *ExplainableError) AddError(err error) {
	e.errors = append(e.errors, err)
}

func (e *ExplainableError) HasErrors() bool {
	return len(e.errors) > 0
}
