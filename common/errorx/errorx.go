package errorx

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
)

var (
	errMessages   = make(map[string]string)
	errorFilePath string     // Stores the last used file path
	mu            sync.Mutex // Protects errMessages map
	fileMu        sync.Mutex // Protects file writes
)

type CustomError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

const defaultPath = "config/errors.json"

// LoadErrors loads a JSON file containing a mapping of error codes to human-readable messages.
// The file path is optional; if not provided, it defaults to "config/errors.json".
// The loaded messages are stored in the errMessages map, which can be accessed
// using the Get and GetMessage functions.
func LoadErrors(filePath ...string) error {
	mu.Lock()
	defer mu.Unlock()

	if len(filePath) > 0 {
		errorFilePath = filePath[0]
	} else {
		errorFilePath = defaultPath
	}

	file, err := os.Open(errorFilePath)
	if err != nil {
		return fmt.Errorf("loading errors.json file failed: %s", err)
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	err = decoder.Decode(&errMessages)
	if err != nil {
		return fmt.Errorf("decoding errors.json file failed: %s", err)
	}

	return nil

}



// Get returns a *CustomError for the given code. If the code is not found in
// the loaded error messages, it returns a *CustomError with the code set to
// "unknown_error" and the message set to "An unknown error occurred". If a
// message is provided as the second argument, the code/message pair is added to
// the loaded error messages and the updated file is written to disk.
func Get(code string, message ...string) *CustomError {
	mu.Lock()
	defer mu.Unlock()

	if msg, exists := errMessages[code]; exists {
		return &CustomError{Code: code, Message: msg}
	}
	if len(message) == 0 {
		return &CustomError{Code: "unknown_error", Message: "An unknown error occurred"}
	}

	newMessage := message[0]
	errMessages[code] = newMessage

	if err := appendErrorToFile(code, newMessage); err != nil {
		fmt.Println("Error appending error to file:", err) // Return error to caller
	}

	return &CustomError{Code: code, Message: newMessage}
}

func appendErrorToFile(code, message string) error {
	fileMu.Lock()
	defer fileMu.Unlock()

	if errorFilePath == "" {
		return fmt.Errorf("no error file path set. Call LoadErrors() or SetErrorFilePath() first")
	}

	// 🔹 Ensure `code` and `message` are added to `errMessages`
	errMessages[code] = message

	// 🔹 Convert `errMessages` to JSON
	data, err := json.MarshalIndent(errMessages, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal error messages: %w", err)
	}

	// 🔹 Write updated JSON back to the file
	err = os.WriteFile(errorFilePath, data, 0644)
	if err != nil {
		return fmt.Errorf("failed to write to %s: %w", errorFilePath, err)
	}

	return nil
}

// GetMessage retrieves the error message associated with the given error code.
// If the code is not found, it returns "undefined error".
func GetMessage(code string) string {
	if msg, exists := errMessages[code]; exists {
		return msg
	}
	return "undefined error"
}

// === helper function ===

func GetErrorFilePath() string {
	mu.Lock()
	defer mu.Unlock()
	return errorFilePath
}
