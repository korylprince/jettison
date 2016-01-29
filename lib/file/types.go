package file

import (
	"encoding/json"
	"strconv"
)

//Definition is a go representation of a json config:
//map[group]map[origin_path]destination_path
//origin_path and destination_path must both be files or both be directories
type Definition map[string]map[string]string

//Set is a mapping, map[xxHash (64 bit)]:path
type Set map[uint64]string

//MarshalJSON returns the JSON representation of the Set or an error if one occurs
func (s Set) MarshalJSON() ([]byte, error) {
	new := make(map[string]string)
	for k, v := range s {
		new[strconv.FormatUint(k, 10)] = v
	}
	return json.Marshal(new)
}

//VersionedSet is a Set grouped with it's version (sum of all hashes in set)
type VersionedSet struct {
	Set     Set
	Version uint64
}
