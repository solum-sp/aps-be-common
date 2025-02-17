package main

import (
	"fmt"
	"log"

	errorx "github.com/solum-sp/aps-be-common/common/errorx"
)

/*
content in errors.json:

{
    "104041": "User does not exist",
    
    "104001": "Incorrect password",
    "104002": "Invalid JWT token",
    "104003": "Token has expired",
    
    "10500": "An unexpected error occurred"
}
*/

type commonResponse struct {
	code string
	message string
}

func main() {

	// Load errors from JSON
	err := errorx.LoadErrors("errors.json")
	if err != nil {
		log.Fatalf("Failed to load error messages: %v", err)
	}

	// Example usage
	fmt.Println(errorx.Get("104041"))
	fmt.Println(errorx.Get("10500")) // Should return "An unexpected error occurred"
	fmt.Println(errorx.GetMessage("10500"))

	e := errorx.Get("10500")

	c := commonResponse{
		code: e.Code,
		message: e.Message,
	}

	_ = errorx.Get("105002", "test new error")
	fmt.Println(c)

	fmt.Println("new error added:", errorx.Get("105002"))

	fmt.Println("Path to errors.json:", errorx.GetErrorFilePath())
	
}
