package format

import (
	"io/ioutil"
	"os"
	"reflect"
	"testing"

	"github.com/BurntSushi/toml"
)

var expected = map[string]interface{}{
	"base_service_url":    "httpbin.org",
	"example_value":       "some example value",
	"type_classification": "standard",
}

func TestEnvFormat(t *testing.T) {
	b, err := ioutil.ReadFile("testdata/env.toml")
	if err != nil {
		t.Fatal(err)
	}
	m := make(map[string]interface{})
	if err := toml.Unmarshal(b, &m); err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(m, expected) {
		t.Fatal("data does not match the expected result")
	}
}

func TestLitmusFileFormat(t *testing.T) {
	b, err := ioutil.ReadFile("testdata/1_type_test.toml")
	if err != nil {
		t.Fatal(err)
	}
	var file LitmusFile
	if err = toml.Unmarshal(b, &file); err != nil {
		t.Fatal(err)
	}

	enc := toml.NewEncoder(os.Stdout)
	enc.Encode(file)
}
