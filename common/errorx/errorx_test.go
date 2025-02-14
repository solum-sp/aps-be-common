package errorx

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

var mockErrorData = map[string]string{
	"104041": "User does not exist",
	"104001": "Incorrect password",
	"10500":  "An unexpected error occurred",
}

func createTempErrorFile(t *testing.T) string {
	tempFile, err := os.CreateTemp("", "errors-*.json")
	assert.NoError(t, err)

	data, err := json.Marshal(mockErrorData)
	assert.NoError(t, err)

	_, err = tempFile.Write(data)
	assert.NoError(t, err)

	tempFile.Close()
	return tempFile.Name()
}

func TestLoadErrors(t *testing.T) {
	tmpFilePath := createTempErrorFile(t)
	defer os.Remove(tmpFilePath)

	err := LoadErrors(tmpFilePath)
	assert.NoError(t, err)

	for key, expectedMessage := range mockErrorData {
		message := GetMessage(key)
		assert.Equal(t, expectedMessage, message)
	}
}

func TestGet(t *testing.T) {
	errMessages = mockErrorData
	actualerr := Get("104041")
	// assert.Equal(t, "104041", err.Code)
	expected := &CustomError{Code: "104041", Message: "User does not exist"}
	assert.Equal(t, expected, actualerr)

	actualerr = Get("10400109")
	expected = &CustomError{Code: "unknown_error", Message: "An unknown error occurred"}
	assert.Equal(t, expected, actualerr)
}

func TestGetMessage(t *testing.T) {
	errMessages = mockErrorData
	message := GetMessage("104041")
	assert.Equal(t, "User does not exist", message)

	message = GetMessage("104001")
	assert.Equal(t, "Incorrect password", message)

	message = GetMessage("100909")
	assert.Equal(t, "undefined error", message)
}
