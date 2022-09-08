package common

import "encoding/json"

// StringPtr returns pointer to a string
func StringPtr(v string) *string {
	return &v
}

// Int32Ptr returns pointer to a int32
func Int32Ptr(v int32) *int32 {
	return &v
}

// Int64Ptr returns pointer to a int64
func Int64Ptr(v int64) *int64 {
	return &v
}

func DeserializeTaskToken(taskToken []byte) *TaskToken {

	token := &TaskToken{}
	err := json.Unmarshal(taskToken, token)
	if err != nil {
		return nil
	}
	return token
}
