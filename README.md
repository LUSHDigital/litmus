# litmus
Run automated HTTP requests from the command line.

## installation

```bash
$ go get -u github.com/codingconcepts/litmus
```

## usage

```bash
litmus -h
    -c string
        config path (default "routesdir")
  -e value
        environment variable
  -n string
        name of specific test to run
```

## example

In this example, we talk to a local service that manages "types".  We first create a type, capture the ID that was assigned to it in the database, then delete it using the ID captured:

### config

The `env.yaml` file contains environment configuration that's shared between test files.

```yaml
- key: type_name
  value: Litmus Type Name

- key: type_classification
  value: standard
```

### tests

The `*_test.yaml` files contain the requests that will be made.  They're executed in the order they appear in the directory.

```yaml
- name: create type - valid request
  method: POST
  url: http://{{.base_service_url}}/types
  headers:
    "Content-Type": "application/json"
  body: |
    {
      "name": "Request Runner Type",
      "classification": "standard"
    }
  getters:
  - { path: code, type: body, exp: 200 }
  - { path: data.type.id, type: body, set: type_id }

- name: delete type - valid id
  method: DELETE
  url: http://{{.base_service_url}}/types/{{.type_id}}
  getters:
  - { path: code, type: body, exp: 200 }
  - { path: data.type.id, type: body, exp: "{{.type_id}}" }
  - { path: data.type.rows_affected, type: body, exp: 1 }
```

### run command

```bash
$ litmus -c routesdir -e base_service_url=localhost
```

## todo
* interface for different types of path parsers (json, xml etc.)
* get response body into output for failures
* tidy up `main.go`
* rename to `litmus`