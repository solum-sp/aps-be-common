package main

import (
	"log"

	"github.com/solum-sp/aps-be-common/common/loader"
)

type SysError struct {
	ErrIncorrectPassword string `json:"104001"`
	ErrInvalidJWTToken   string `json:"104002"`
	ErrTokenExpired      string `json:"104003"`
	ErrUserDoesNotExist  string `json:"104041"`
	ErrNoContent         string `json:"NoContent"`
}


func main() {
	var ec SysError
	err := loader.Load(&ec)
	if err != nil {
		log.Fatalf("Failed to load error messages: %v", err)
	}
}