package errors

import (
	"errors"
	"fmt"
)

type OrderAlreadyCreatedByAnotherUserError struct {
	Err error
}

func NewOrderAlreadyCreatedByAnotherUser(userID string) error {
	return &OrderAlreadyCreatedByAnotherUserError{
		Err: fmt.Errorf(`%w`, errors.New(userID)),
	}
}

func (o OrderAlreadyCreatedByAnotherUserError) Unwrap() error {
	return o.Err
}

func (o OrderAlreadyCreatedByAnotherUserError) Error() string {
	return fmt.Sprintf(
		"another usert with id %s has already plased an order with this number",
		o.Err)
}
