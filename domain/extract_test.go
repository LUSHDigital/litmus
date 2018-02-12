package domain

import (
	"net/http"
	"testing"

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
		r    *RequestTest
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
				r: &RequestTest{
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
				r:    &RequestTest{WantsCode: 200},
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
	gock.New("/").
		Reply(200).
		SetHeader("content-type", "application/json").
		BodyString(`{"hello":"world"}`)

	res, err := http.Get("/")
	if err != nil {
		t.Fatal(err)
	}

	type args struct {
		r    *RequestTest
		resp *http.Response
		env  map[string]interface{}
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "simple",
			args: args{
				r: &RequestTest{
					Body: map[string]interface{}{
						"hello": "world",
					},
				},
				resp: res,
				env:  map[string]interface{}{},
			},
			wantErr: false,
		},
		{
			name: "complex",
			args: args{
				r: &RequestTest{
					Body: map[string]interface{}{
						"val": map[string]interface{}{
							"hello": "world",
						},
					},
				},
				resp: res,
				env:  map[string]interface{}{},
			},
			wantErr: false,
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
	gock.New("/").
		Reply(200).
		SetHeader("content-type", "application/json").
		BodyString(`{"hello":"world"}`)

	res, err := http.Get("/")
	if err != nil {
		t.Fatal(err)
	}
	type args struct {
		r    *RequestTest
		resp *http.Response
		env  map[string]interface{}
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "simple",
			args: args{
				r: &RequestTest{
					Head: map[string]interface{}{
						"content-type": "application/json"},
				},
				resp: res,
				env:  make(map[string]interface{}),
			},
			wantErr: false,
		},
		{
			name: "complex",
			args: args{
				r: &RequestTest{
					Head: map[string]interface{}{
						"some_key": map[string]interface{}{
							"content-type": "application/json",
						},
					},
				},
				resp: res,
				env:  make(map[string]interface{}),
			},
			wantErr: false,
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
