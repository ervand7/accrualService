package database

import (
	"testing"

	"github.com/ervand7/go-musthave-diploma-tpl/internal/logger"
	"github.com/pressly/goose/v3"
	"github.com/stretchr/testify/assert"
)

func Downgrade() {
	db := database{}
	db.run()
	if err := goose.Run("down", db.conn, getMigrationsDir()); err != nil {
		logger.Logger.Error(err.Error())
	}
}

func UserIDFixture(
	storage *Storage, login, password, token string, t *testing.T,
) (userID string) {
	rows, err := storage.db.conn.Query(`
		insert into "public"."user" ("login", "password", "token") values
		($1, $2, $3) returning "user"."id"
		`, login, password, token,
	)
	assert.NoError(t, err)
	defer storage.db.closeRows(rows)

	for rows.Next() {
		err = rows.Scan(&userID)
		assert.NoError(t, err)
	}
	err = rows.Err()
	assert.NoError(t, err)

	return userID
}
