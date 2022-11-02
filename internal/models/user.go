package models

import "github.com/jackc/pgtype"

type User struct {
	ID       pgtype.UUID
	Login    string
	Password string
	Token    string
}
