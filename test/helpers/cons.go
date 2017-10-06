package helpers

import (
	"fmt"
	"time"
)

var timeout = 300 * time.Second
var basePath = "/vagrant/"

func GetFilePath(filename string) string {
	return fmt.Sprintf("%s%s", basePath, filename)
}
