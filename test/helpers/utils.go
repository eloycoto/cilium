package helpers

import (
	"bytes"
	"fmt"
	"html/template"
	"io/ioutil"
	"os"
	"time"
)

//Sleep sleep a function for a specific time
func Sleep(delay time.Duration) {
	time.Sleep(delay * time.Second)
}

//CountValues Filter an array of strings and return the number of matches and
//the len of the array
func CountValues(key string, data []string) (int, int) {
	var result int

	for _, x := range data {
		if x == key {
			result++
		}
	}
	return result, len(data)
}

//RenderTemplateToFile render a string using go templates to a file
func RenderTemplateToFile(filename string, tmplt string, perm os.FileMode) error {
	t, err := template.New("").Parse(tmplt)
	if err != nil {
		return err
	}
	content := new(bytes.Buffer)
	err = t.Execute(content, nil)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(filename, content.Bytes(), perm)
	if err != nil {
		return err
	}
	return nil
}

//TimeoutConfig struct to define a timeout and Ticker
type TimeoutConfig struct {
	Ticker  time.Duration
	Timeout time.Duration
}

//WithTimeout helper function that execute a function each TimeoutConfig.Ticker(default 3) and
// it'll die if on timeout no true is returned
func WithTimeout(body func() bool, msg string, config *TimeoutConfig) error {
	if config.Ticker == 0 {
		config.Ticker = 5
	}

	done := time.After(config.Timeout * time.Second)
	ticker := time.NewTicker(config.Ticker * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if body() {
				return nil
			}
		case <-done:
			return fmt.Errorf("Timeout reached: %s", msg)
		}
	}
}
