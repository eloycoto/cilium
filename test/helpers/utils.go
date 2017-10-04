package helpers

import (
	"bytes"
	"html/template"
	"io/ioutil"
	"os"
	"time"
)

func Sleep(delay time.Duration) {
	time.Sleep(delay * time.Second)
}

//CountValues: Filter an array of strings and return the number of matches and
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
