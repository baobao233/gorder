package util

import "encoding/json"

func MarshallString(v any) (string, error) {
	bytes, err := json.Marshal(v)
	return string(bytes), err
}
