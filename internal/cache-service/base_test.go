package cacheservice

import (
	"bytes"
	"encoding/json"
)

func compareJSONs(a, b any) bool {
	by1, err := json.Marshal(a)
	if err != nil {
		panic("mop")
	}
	by2, err := json.Marshal(b)
	if err != nil {
		panic("mop")
	}

	return bytes.Equal(by1, by2)
}
