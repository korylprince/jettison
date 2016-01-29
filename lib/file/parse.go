package file

import (
	"encoding/json"
	"os"
)

//Parse parses the given json file into a Definition, or returns an error if one occurs
func Parse(path string) (Definition, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var d Definition

	dec := json.NewDecoder(f)
	err = dec.Decode(&d)

	return d, err
}
