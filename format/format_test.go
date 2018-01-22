package format

import (
	"io/ioutil"
	"strings"
	"testing"

	"github.com/BurntSushi/toml"
	"github.com/LUSHDigital/litmus/test"
)

func TestEnvFormat(t *testing.T) {
	const envFile = `
		base_service_url="httpbin.org"
		example_value="some example value"
		type_classification="standard"`

	b, err := ioutil.ReadAll(strings.NewReader(envFile))
	if err != nil {
		t.Fatal(err)
	}
	m := make(map[string]interface{})
	if err := toml.Unmarshal(b, &m); err != nil {
		t.Fatal(err)
	}

	test.Equals(t, "httpbin.org", m["base_service_url"])
	test.Equals(t, "some example value", m["example_value"])
	test.Equals(t, "standard", m["type_classification"])
}

func TestLitmusFileFormat(t *testing.T) {
	const testFile = `
		[[litmus.test]]
		name="httpbin get - check body"
		method= "GET"
		url= "http://{{.base_service_url}}/get"
		[litmus.test.query]
		foo = "bar"
		baz = "qux"
		[litmus.test.headers]
		Content-Type = "application/json"
		[[litmus.test.getters]]
		type="body"
		path="headers.Connection"
		exp="close"
		[[litmus.test.getters]]
		type="body"
		path="headers.Connection"
		exp="close"
		set="some_key"`

	b, err := ioutil.ReadAll(strings.NewReader(testFile))
	if err != nil {
		t.Fatal(err)
	}
	var file TestFile
	if err = toml.Unmarshal(b, &file); err != nil {
		t.Fatal(err)
	}

	test.Equals(t, "httpbin get - check body", file.Litmus.Test[0].Name)
	test.Equals(t, "GET", file.Litmus.Test[0].Method)
	test.Equals(t, "http://{{.base_service_url}}/get", file.Litmus.Test[0].URL)

	test.Equals(t, "bar", file.Litmus.Test[0].Query["foo"])
	test.Equals(t, "application/json", file.Litmus.Test[0].Headers["Content-Type"])

	test.Equals(t, "body", file.Litmus.Test[0].Getters[0].Type)
	test.Equals(t, "headers.Connection", file.Litmus.Test[0].Getters[0].Path)
	test.Equals(t, "close", file.Litmus.Test[0].Getters[0].Expected)

	test.Equals(t, "body", file.Litmus.Test[0].Getters[1].Type)
	test.Equals(t, "headers.Connection", file.Litmus.Test[0].Getters[1].Path)
	test.Equals(t, "close", file.Litmus.Test[0].Getters[1].Expected)
	test.Equals(t, "some_key", file.Litmus.Test[0].Getters[1].Set)
}

func TestGetterConfigs_Filter(t *testing.T) {
	configs := GetterConfigs{
		GetterConfig{Set: "a", Type: "body"},
		GetterConfig{Set: "b", Type: "head"},
		GetterConfig{Set: "c", Type: "body"},
	}

	got := configs.Filter("body")

	test.Equals(t, 2, len(got))
	test.Equals(t, "a", got[0].Set)
	test.Equals(t, "c", got[1].Set)
}
