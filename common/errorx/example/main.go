package main

import (
	"fmt"
	"log"

	// "github.com/solum-sp/aps-be-common/errorx"
	errorx "github.com/solum-sp/aps-be-common/common/errorx"
)
func main() {
	// Load errors from JSON
	err := errorx.LoadErrors("error.json")
	if err != nil {
		log.Fatalf("Failed to load error messages: %v", err)
	}

	// Example usage
	fmt.Println(errorx.Get("104041"))
	
	fmt.Println(errorx.Get("10500")) // Should return "An unexpected error occurred"
	fmt.Println(errorx.GetMessage("10500"))
}
