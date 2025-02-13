package main

import (
	_ "github.com/solum-sp/aps-be-common/common/cache"
	_ "github.com/solum-sp/aps-be-common/common/config"
	_ "github.com/solum-sp/aps-be-common/common/event"
	"github.com/solum-sp/aps-be-common/common/logger"
	_ "github.com/solum-sp/aps-be-common/common/logger"
	_ "github.com/solum-sp/aps-be-common/common/utils"
)

func main() {
	l, err := logger.NewLogger(logger.Config{
		Service: "test",
		Level:   logger.DebugLv,
	})
	if err != nil {
		panic(err)
	}

	l.Error("Hello, world!", "key", "value")
	l.Info("Hello, world!", "key", "value")

}
