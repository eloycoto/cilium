package helpers

import "time"

func Sleep(delay time.Duration) {
	time.Sleep(delay * time.Second)
}
