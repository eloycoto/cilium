package helpers

import (
	"fmt"
	"time"
)

var timeout = 300 * time.Second
var basePath = "/vagrant/"

//GetFilePath return the file with an absolute path
func GetFilePath(filename string) string {
	return fmt.Sprintf("%s%s", basePath, filename)
}
