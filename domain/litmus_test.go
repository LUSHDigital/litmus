package domain

import (
	"testing"
)

func Test_ExpandEnv(t *testing.T) {
	globalEnv := map[string]interface{}{
		"key1": "value1",
		"key2": "value2",
	}
	requestEnv := map[string]interface{}{
		"key3":      123,
		"key4":      "{{.key1}}",
		"{{.key2}}": "whatever",
	}
	if err := modifyRequestEnv(requestEnv, globalEnv); err != nil {
		t.Fatal(err)
	}

	Equals(t, "whatever", requestEnv["value2"])
	Equals(t, "123", requestEnv["key3"])
	Equals(t, "value1", requestEnv["key4"])
}
