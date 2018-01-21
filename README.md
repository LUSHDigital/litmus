# Litmus
Run automated HTTP requests from the command line.

## Installation

```bash
go get github.com/{...}/litmus
```

## Usage

```bash
Litmus helps you perform HTTP tests by providing a simple, TOML based DSL
for you to write endpoint tests with

Usage:
  litmus [flags]

Flags:
  -c, --config string            path to configuration folder
  -e, --env *pkg.KeyValuePairs   environment variables: example baseurl=httpbin.org"
  -h, --help                     help for litmus
  -n, --test string              name of specific test to run

```

## Example

In this example, we talk to [httpbin](http://httpbin.org/), perform some assertions and set some environment variables for later reuse.

### Configuration

The `env.toml` file contains environment configuration that's shared between test files.

```toml
base_service_url="httpbin.org"
example_value="some example value"
int_value=123
```

### Writing Tests

The `*_test.toml` files contain the requests that will be made.  They're executed in the order they appear in the directory.

```yaml
[litmus]

# this test checks if a 200 content response
# is obtainable from the server
[[litmus.test]]
	name="httpbin get - check code"
	method="GET"
	url="https://{{.base_service_url}}/get"
	wants_code=200

# this test checks if the body contains the "Connection"
# field, set to the value "close"
[[litmus.test]]
	name="httpbin get - check body"
	method="GET"
	url="http://{{.base_service_url}}/get"
	want_code=200
[litmus.test.query]
    foo = "bar"
    baz = "qux"
[[litmus.test.getters]] # multiple getters arrays are ok!
	type="body"
	path="headers.Connection"
	exp="close"
	set="some_key"
	# here, if the path "headers.Connection" exists in
	# the JSON body returned by the request,
	# we capture that value and set 'some_key'
	# in the environment. This value can
	# be reused in future requests if needed!

# This is an example for a post request
[[litmus.test]]
name= "httpbin post - returns post data"
method= "POST"
url="http://{{.base_service_url}}/post"
wants_code= 200
# note that we reuse the previously set value in the body.
body='''
{
	"from_env":"{{.example_value}}",
	"test":"{{.some_key}}"
}
'''
[litmus.test.headers]
Content-Type = "application/json"
[[litmus.test.getters]]
# etc...
```

### Run command

```bash
# simple call
litmus -c path/to/tests

# setting environment variables on the fly.
# note that this will supersede anything set in the env.toml
litmus -c path/to/tests -e base_service_url=localhost
```

## Roadmap
* Display response body on failure.
* Multiple header values support, currently only the first match will be checked, possibly with optional indexer:
  * `Content-Type,0 == Content-Type`
