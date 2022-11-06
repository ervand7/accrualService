package errors

import (
	"errors"
	"fmt"
)

type OrderAlreadyExistsError struct {
	Err             error
	FromCurrentUser bool
}

func NewOrderAlreadyExistsError(userID string, fromCurrentUser bool) error {
	return &OrderAlreadyExistsError{
		Err:             fmt.Errorf(`%w`, errors.New(userID)),
		FromCurrentUser: fromCurrentUser,
	}
}

func (o OrderAlreadyExistsError) Unwrap() error {
	return o.Err
}

func (o OrderAlreadyExistsError) Error() string {
	return "order already exists"
}
