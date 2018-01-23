package domain

import (
	"testing"
	"io/ioutil"
	"github.com/BurntSushi/toml"
	"github.com/davecgh/go-spew/spew"
)

func TestUnmarshal(t *testing.T) {
	b, err := ioutil.ReadFile("testdata/test.toml")
	if err != nil {
		t.Fatal(err)
	}

	var tf TestFile
	if err := toml.Unmarshal(b, &tf); err != nil {
		t.Fatal(err)
	}

	spew.Dump(tf)
}

//func TestEnvFormat(t *testing.T) {
//	const envFile = `
//		base_service_url="httpbin.org"
//		example_value="some example value"
//		type_classification="standard"`
//
//	b, err := ioutil.ReadAll(strings.NewReader(envFile))
//	if err != nil {
//		t.Fatal(err)
//	}
//	m := make(map[string]interface{})
//	if err := toml.Unmarshal(b, &m); err != nil {
//		t.Fatal(err)
//	}
//
//	Equals(t, "httpbin.org", m["base_service_url"])
//	Equals(t, "some example value", m["example_value"])
//	Equals(t, "standard", m["type_classification"])
//}
//
//func TestLitmusFileFormat(t *testing.T) {
//	const testFile = `
//		[[litmus.test]]
//		name="httpbin get - check body"
//		method= "GET"
//		url= "http://{{.base_service_url}}/get"
//		[litmus.test.query]
//		foo = "bar"
//		baz = "qux"
//		[litmus.test.headers]
//		Content-Type = "application/json"
//		[[litmus.test.getters]]
//		type="body"
//		path="headers.Connection"
//		exp="close"
//		[[litmus.test.getters]]
//		type="body"
//		path="headers.Connection"
//		exp="close"
//		set="some_key"`
//
//	b, err := ioutil.ReadAll(strings.NewReader(testFile))
//	if err != nil {
//		t.Fatal(err)
//	}
//	var file TestFile
//	if err = toml.Unmarshal(b, &file); err != nil {
//		t.Fatal(err)
//	}
//
//	Equals(t, "httpbin get - check body", file.Litmus.Test[0].Name)
//	Equals(t, "GET", file.Litmus.Test[0].Method)
//	Equals(t, "http://{{.base_service_url}}/get", file.Litmus.Test[0].URL)
//
//	Equals(t, "bar", file.Litmus.Test[0].Query["foo"])
//	Equals(t, "application/json", file.Litmus.Test[0].Headers["Content-Type"])
//
//	Equals(t, "body", file.Litmus.Test[0].Getters[0].Type)
//	Equals(t, "headers.Connection", file.Litmus.Test[0].Getters[0].Path)
//	Equals(t, "close", file.Litmus.Test[0].Getters[0].Expected)
//
//	Equals(t, "body", file.Litmus.Test[0].Getters[1].Type)
//	Equals(t, "headers.Connection", file.Litmus.Test[0].Getters[1].Path)
//	Equals(t, "close", file.Litmus.Test[0].Getters[1].Expected)
//	Equals(t, "some_key", file.Litmus.Test[0].Getters[1].Set)
//}
//
//func TestGetterConfigs_Filter(t *testing.T) {
//	configs := GetterConfigs{
//		GetterConfig{Set: "a", Type: "body"},
//		GetterConfig{Set: "b", Type: "head"},
//		GetterConfig{Set: "c", Type: "body"},
//	}
//
//	got := configs.Filter("body")
//
//	Equals(t, 2, len(got))
//	Equals(t, "a", got[0].Set)
//	Equals(t, "c", got[1].Set)
//}

func Test_expandEnv(t *testing.T) {
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
