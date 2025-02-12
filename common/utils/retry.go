package utils

import (
	"fmt"
	"log"
	"time"
)

func Retry[T any](attempts int, sleep time.Duration, f func() (T, error)) (result T, err error) {
	for i := 0; i < attempts; i++ {
		if i > 0 {
			log.Printf("Failed to execute function, attempt %d/%d: %s\n", i, attempts, err)
			log.Printf("Sleeping for %s\n", sleep)
			time.Sleep(sleep)
			sleep *= 2
		}
		result, err = f()
		if err == nil {
			return result, nil
		}
	}
	return result, fmt.Errorf("after %d attempts, last error: %s", attempts, err)
}
