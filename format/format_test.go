package format

import (
	"io/ioutil"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/BurntSushi/toml"
)

var expected = map[string]interface{}{
	"base_service_url":    "httpbin.org",
	"example_value":       "some example value",
	"type_classification": "standard",
}

var envFile = `
	base_service_url="httpbin.org"
	example_value="some example value"
	type_classification="standard"`

var testFile = `
	[[litmus.test]]
	name="httpbin get - check body"
	method= "GET"
	url= "http://{{.base_service_url}}/get"
	[[litmus.test.getters]]
	type="body"
	path="headers.Connection"
	exp="close"
	[[litmus.test.getters]]
	type="body"
	path="headers.Connection"
	exp="close"
	set="some_key"`

func TestEnvFormat(t *testing.T) {
	b, err := ioutil.ReadAll(strings.NewReader(envFile))
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
	b, err := ioutil.ReadAll(strings.NewReader(testFile))
	if err != nil {
		t.Fatal(err)
	}
	var file TestFile
	if err = toml.Unmarshal(b, &file); err != nil {
		t.Fatal(err)
	}

	enc := toml.NewEncoder(os.Stdout)
	enc.Encode(file)
}
