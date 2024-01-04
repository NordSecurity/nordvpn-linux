package client

import (
	"encoding/json"
)

func InterfaceToInt64(item interface{}) int64 {
	var i int64
	switch item := item.(type) {
	case json.Number:
		i, _ = item.Int64()
	case int64:
		i = item
	}
	return i
}
