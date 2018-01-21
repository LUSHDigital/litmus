package extract

import (
	"net/http"
	"testing"

	"github.com/codingconcepts/litmus/format"
	"github.com/h2non/gock"
)

const binPath = "https://httpbin.org"

func TestStatusCode(t *testing.T) {
	defer gock.Off()

	gock.New(binPath).
		Get("/get").
		Reply(200)

	okResp, err := http.Get(binPath + "/get")

	if err != nil {
		t.Fatal(err)
	}

	gock.New(binPath).
		Get("/get").
		Reply(500)

	failResp, err := http.Get(binPath + "/get")
	if err != nil {
		t.Fatal(err)
	}

	type args struct {
		r    format.RequestTest
		resp *http.Response
		in2  map[string]interface{}
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name:    "response is nil",
			wantErr: true,
		},
		{
			name: "code matches",
			args: args{
				r: format.RequestTest{
					WantsCode: 200,
				},
				resp: okResp,
				in2:  nil,
			},
			wantErr: false,
		},
		{
			name: "code does not match",
			args: args{
				r:    format.RequestTest{WantsCode: 200},
				resp: failResp,
				in2:  nil,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := StatusCode(tt.args.r, tt.args.resp, tt.args.in2); (err != nil) != tt.wantErr {
				t.Errorf("StatusCode() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestBody(t *testing.T) {
	defer gock.Off()
	gock.New(binPath).
		Get("/get").
		Reply(200).
		SetHeader("Content-Type", "application/json").
		BodyString(`{"hello":"world""}`)

	res, err := http.Get(binPath + "/get")
	if err != nil {
		t.Fatal(err)
	}

	gock.New(binPath).
		Get("/get").
		Reply(200).
		BodyString(`{"hello":"world""}`)

	failRes, err := http.Get(binPath + "/get")
	if err != nil {
		t.Fatal(err)
	}

	type args struct {
		r    format.RequestTest
		resp *http.Response
		env  map[string]interface{}
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "body matched and sets key",
			args: args{
				r: format.RequestTest{
					Getters: format.GetterConfigs{
						{Path: "hello", Type: "body", Expected: "world"},
						{Path: "hello", Type: "body", Expected: "world", Set: "some_key"},
					},
				},
				resp: res,
				env:  make(map[string]interface{}),
			},
			wantErr: false,
		},
		{
			name: "body matched and sets key on nil map",
			args: args{
				r: format.RequestTest{
					Getters: format.GetterConfigs{
						{Path: "hello", Type: "body", Expected: "world"},
						{Path: "hello", Type: "body", Expected: "world", Set: "some_key"},
					},
				},
				resp: res,
			},
			wantErr: true,
		},
		{
			name: "body path does not match",
			args: args{
				resp: res,
				r: format.RequestTest{
					Getters: format.GetterConfigs{
						{Path: "some.broken.path", Type: "body", Expected: "world"},
					},
				},
				env: nil,
			},
			wantErr: true,
		},
		{
			name: "body expectation does not match",
			args: args{
				resp: res,
				r: format.RequestTest{
					Getters: format.GetterConfigs{
						{Path: "hello", Type: "body", Expected: "wrong expectation"},
					},
				},
				env: nil,
			},
			wantErr: true,
		},
		{
			name: "missing content type",
			args: args{
				r: format.RequestTest{
					Getters: format.GetterConfigs{
						{Path: "some.broken.path", Type: "body", Expected: "world"},
					},
				},
				resp: failRes,
				env:  nil,
			},
			wantErr: true,
		},
		{
			name:    "response is nil",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := Body(tt.args.r, tt.args.resp, tt.args.env); (err != nil) != tt.wantErr {
				t.Errorf("Body() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestHeader(t *testing.T) {
	defer gock.Off()
	gock.New(binPath).
		Reply(200).
		SetHeader("X-Some-Header", "test")

	okRes, err := http.Get(binPath)
	if err != nil {
		t.Fatal(err)
	}

	gock.New(binPath).Reply(200)
	missingRes, err := http.Get(binPath)
	if err != nil {
		t.Fatal(err)
	}

	type args struct {
		r    format.RequestTest
		resp *http.Response
		env  map[string]interface{}
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "header matched",
			args: args{
				r:    format.RequestTest{Getters: format.GetterConfigs{{Type: "head", Path: "X-Some-Header", Expected: "test"}}},
				resp: okRes,
				env:  nil,
			},
			wantErr: false,
		},
		{
			name: "header matched and sets key",
			args: args{
				r:    format.RequestTest{Getters: format.GetterConfigs{{Type: "head", Path: "X-Some-Header", Expected: "test", Set: "some_key"}}},
				resp: okRes,
				env:  make(map[string]interface{}),
			},
			wantErr: false,
		},
		{
			name: "header matched and sets key on nil map",
			args: args{
				r:    format.RequestTest{Getters: format.GetterConfigs{{Type: "head", Path: "X-Some-Header", Expected: "test", Set: "some_key"}}},
				resp: okRes,
			},
			wantErr: true,
		},
		{
			name: "missing getter type",
			args: args{
				r:    format.RequestTest{Getters: format.GetterConfigs{{Type: "", Path: "X-Some-Header", Expected: "test"}}},
				resp: okRes,
				env:  nil,
			},
			wantErr: false,
		},
		{
			name: "header missing",
			args: args{
				r:    format.RequestTest{Getters: format.GetterConfigs{{Type: "head", Path: "X-Some-Header", Expected: "test"}}},
				resp: missingRes,
				env:  nil,
			},
			wantErr: true,
		},
		{
			name: "header value incorrect",
			args: args{
				r:    format.RequestTest{Getters: format.GetterConfigs{{Type: "head", Path: "X-Some-Header", Expected: "wrong expectation"}}},
				resp: okRes,
				env:  nil,
			},
			wantErr: true,
		},
		{
			name:    "nil response",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := Header(tt.args.r, tt.args.resp, tt.args.env); (err != nil) != tt.wantErr {
				t.Errorf("Header() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
