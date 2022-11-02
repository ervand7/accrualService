package config

import (
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestGetConfig_DefaultValues(t *testing.T) {
	assert.NotContainsf(t, os.Args, "-a", "flag -a was set")
	assert.Equal(t, os.Getenv("RUN_ADDRESS"), "")
	expectedRunAddress := ":8081"
	assert.Equal(t, GetConfig().RunAddress, expectedRunAddress)

	assert.NotContainsf(t, os.Args, "-d", "flag -d was set")
	assert.Equal(t, os.Getenv("DATABASE_URI"), "")
	expectedDatabaseURI := ""
	assert.Equal(t, GetConfig().DatabaseURI, expectedDatabaseURI)

	assert.NotContainsf(t, os.Args, "-r", "flag -r was set")
	assert.Equal(t, os.Getenv("ACCRUAL_SYSTEM_ADDRESS"), "")
	expectedAccrualSystemAddress := ":8080"
	assert.Equal(t, GetConfig().AccrualSystemAddress, expectedAccrualSystemAddress)
}

func TestGetConfig_EnvValues(t *testing.T) {
	runAddress := ":5000"
	err := os.Setenv("RUN_ADDRESS", runAddress)
	assert.NoError(t, err)
	assert.Equal(t, GetConfig().RunAddress, runAddress)

	databaseURI := "hello world"
	err = os.Setenv("DATABASE_URI", databaseURI)
	assert.NoError(t, err)
	assert.Equal(t, GetConfig().DatabaseURI, databaseURI)

	accrualSystemAddress := ":5001"
	err = os.Setenv("ACCRUAL_SYSTEM_ADDRESS", accrualSystemAddress)
	assert.NoError(t, err)
	assert.Equal(t, GetConfig().AccrualSystemAddress, accrualSystemAddress)
}

func TestGetConfig_FlagHasPriority(t *testing.T) {
	runAddressFlag := ":5000"
	runAddressEnv := ":5001"
	os.Args = []string{"test", "-a", runAddressFlag}
	err := os.Setenv("RUN_ADDRESS", runAddressEnv)
	assert.NoError(t, err)
	assert.Equal(t, GetConfig().RunAddress, runAddressFlag)

	databaseURIFlag := "https://hello"
	databaseURIEnv := "https://world"
	os.Args = []string{"test", "-d", databaseURIFlag}
	err = os.Setenv("BASE_URL", databaseURIEnv)
	assert.NoError(t, err)
	assert.Equal(t, GetConfig().DatabaseURI, databaseURIFlag)

	accrualSystemAddressFlag := ":5000"
	accrualSystemAddressEnv := ":5001"
	os.Args = []string{"test", "-r", accrualSystemAddressFlag}
	err = os.Setenv("RUN_ADDRESS", accrualSystemAddressEnv)
	assert.NoError(t, err)
	assert.Equal(t, GetConfig().AccrualSystemAddress, accrualSystemAddressFlag)
}
