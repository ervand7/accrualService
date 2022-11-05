package errors

import (
	"errors"
	"fmt"
)

type OrderAlreadyCreatedByCurrentUserError struct {
	Err error
}

func NewOrderAlreadyCreatedByCurrentUser(userID string) error {
	return &OrderAlreadyCreatedByCurrentUserError{
		Err: fmt.Errorf(`%w`, errors.New(userID)),
	}
}

func (o OrderAlreadyCreatedByCurrentUserError) Unwrap() error {
	return o.Err
}

func (o OrderAlreadyCreatedByCurrentUserError) Error() string {
	return fmt.Sprintf(
		"current usert with id %s has already plased an order with this number",
		o.Err)
}
