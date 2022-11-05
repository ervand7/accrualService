package errors

import (
	"errors"
	"fmt"
)

type LoginAlreadyExistsError struct {
	Err error
}

func NewLoginAlreadyExistsError(login string) error {
	return &LoginAlreadyExistsError{
		Err: fmt.Errorf(`%w`, errors.New(login)),
	}
}

func (l LoginAlreadyExistsError) Unwrap() error {
	return l.Err
}

func (l LoginAlreadyExistsError) Error() string {
	return fmt.Sprintf("%s already exists", l.Err)
}
