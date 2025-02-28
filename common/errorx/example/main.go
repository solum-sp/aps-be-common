package main

import (
	"fmt"
	"log"

	"github.com/golobby/container/v3"
	errorx "github.com/solum-sp/aps-be-common/common/errorx"
	"github.com/solum-sp/aps-be-common/common/errorx/example/otherpackage"
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
func init() {
	container.Singleton(func() otherpackage.SysError {
		var errCfg otherpackage.SysError
		err := errorx.Load(&errCfg,"errors.json")
		if err != nil {
			log.Fatalf("Failed to load error messages: %v", err)
		}
		
		return errCfg
	})
	
}



func main() {
	var sysError otherpackage.SysError
	container.Resolve(&sysError)

	fmt.Println("Error message:",sysError.ErrUserDoesNotExist)
}
