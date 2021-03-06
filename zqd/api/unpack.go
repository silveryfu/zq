package api

import (
	"encoding/json"
	"fmt"
)

// unpack transforms a piped json stream into the appropriate api response
// and returns it as an empty interface so that the caller can receive
// a stream of objects, check their types, and process them accordingly.
func unpack(b []byte) (interface{}, error) {
	var v struct {
		Type string `json:"type"`
	}
	if err := json.Unmarshal(b, &v); err != nil {
		return nil, err
	}
	var out interface{}
	switch v.Type {
	case "TaskStart":
		out = &TaskStart{}
	case "TaskEnd":
		out = &TaskEnd{}
	case "SearchRecords":
		out = &SearchRecords{}
	case "SearchWarning":
		out = &SearchWarning{}
	case "SearchStats":
		out = &SearchStats{}
	case "SearchEnd":
		out = &SearchEnd{}
	case "PcapPostStatus":
		out = &PcapPostStatus{}
	case "PcapPostWarning":
		out = &PcapPostWarning{}
	case "LogPostStatus":
		out = &LogPostStatus{}
	case "LogPostWarning":
		out = &LogPostWarning{}
	case "":
		return nil, fmt.Errorf("no type field in search result: %s", string(b))
	default:
		return nil, fmt.Errorf("unknown type in results stream: %s", v.Type)
	}
	if err := json.Unmarshal(b, out); err != nil {
		return nil, err
	}
	return out, nil
}
