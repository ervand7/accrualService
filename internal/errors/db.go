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

func (s LoginAlreadyExistsError) Unwrap() error {
	return s.Err
}

func (s LoginAlreadyExistsError) Error() string {
	return fmt.Sprintf("%s already exists", s.Err)
}
