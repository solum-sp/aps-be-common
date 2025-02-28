package errorx

import (
	"fmt"


	"github.com/solum-sp/aps-be-common/common/loader"
)

var defaulPath = "config/errors.json"
func Load(cfg any, path ...string) error {
	errorFilePath := defaulPath
	if len(path) > 0 && path[0] != ""{
		errorFilePath = path[0]
	}
	
	err := loader.Load(cfg, errorFilePath)
	if err != nil {
		return fmt.Errorf("failed to load errors config: %w", err)
	}
	return nil

}