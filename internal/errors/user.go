package errors

import (
	"fmt"
)

type UserNotFoundError struct {
	Err error
}

func NewUserNotFoundError(login, password string) error {
	return &UserNotFoundError{
		Err: fmt.Errorf(`%w`, fmt.Errorf("%s %s", login, password)),
	}
}

func (u UserNotFoundError) Unwrap() error {
	return u.Err
}

func (u UserNotFoundError) Error() string {
	return fmt.Sprintf("user not found with this credentials: %s", u.Err)
}

type NotEnoughMoneyError struct {
	Err error
}

func NewNotEnoughMoneyError(availableSum float64) error {
	return &NotEnoughMoneyError{
		Err: fmt.Errorf(`%w`, fmt.Errorf("%f", availableSum)),
	}
}

func (n NotEnoughMoneyError) Unwrap() error {
	return n.Err
}

func (n NotEnoughMoneyError) Error() string {
	return fmt.Sprintf("not enough money. Available: %s", n.Err)
}
