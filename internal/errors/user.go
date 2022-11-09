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
