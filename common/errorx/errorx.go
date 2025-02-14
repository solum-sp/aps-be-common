package errorx

import (
	"encoding/json"
	"fmt"
	"os"


)

var errMessages = make(map[string]string)

type CustomError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

const defaultPath = "config/errors.json"

// LoadErrors loads a JSON file containing a mapping of error codes to human-readable messages.
// The file path is optional; if not provided, it defaults to "config/errors.json".
// The loaded messages are stored in the errMessages map, which can be accessed
// using the Get and GetMessage functions.
func LoadErrors(filePath ...string) error{
	path := defaultPath
	if len(filePath) > 0 {
		path = filePath[0]
	}

	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("loading errors.json file failed: %s", err)
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	err = decoder.Decode(&errMessages); 
	if err != nil {
		return fmt.Errorf("decoding errors.json file failed: %s", err)
	}

	return nil

}

// Get retrieves a CustomError corresponding to the provided error code.
// If the code exists in the error messages map, it returns a CustomError
// with the associated code and message. If the code does not exist, it
// returns a CustomError with code "unknown_error" and a default message.

func Get(code string) *CustomError {
	if msg, exists := errMessages[code]; exists {
		return &CustomError{Code: code, Message: msg}
	}
	return &CustomError{Code: "unknown_error", Message: "An unknown error occurred"}
}

// GetMessage retrieves the error message associated with the given error code.
// If the code is not found, it returns "undefined error".
func GetMessage(code string) string {
	if msg, exists := errMessages[code]; exists {
		return msg
	}
	return "undefined error"
}