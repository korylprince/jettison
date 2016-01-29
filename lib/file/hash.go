package file

import (
	"io"
	"os"

	"github.com/OneOfOne/xxhash/native"
)

//Hash returns the xxHash64 of the given path, or an error if one occurred
func Hash(path string) (uint64, error) {
	h := xxhash.New64()
	f, err := os.Open(path)
	if err != nil {
		return 0, err
	}
	defer f.Close()
	_, err = io.Copy(h, f)
	if err != nil {
		return 0, err
	}
	return h.Sum64(), nil
}
