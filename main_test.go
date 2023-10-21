package main

import (
	"context"
	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/mock"
	"math/rand"
	"net/url"
	"strings"

	"net/http"
	"testing"
)

type httpClientStub struct{}

func (c *httpClientStub) Get(u string) (*http.Response, error) {
	_, err := url.Parse(u)
	if err != nil {
		return nil, err
	}
	return httpmock.NewStringResponse(200, strings.Repeat("a", rand.Intn(1000))), nil
}

type httpClientMock struct {
	mock.Mock
}

func (c *httpClientMock) Get(u string) (*http.Response, error) {
	args := c.Called(u)
	return args.Get(0).(*http.Response), args.Error(1)
}

type logStub struct{}

func (*logStub) Print(v ...any) {}

var path = "./example/simple.txt"

func Test_run(t *testing.T) {
	ctx := context.Background()
	log := &logStub{}
	type args struct {
		path string
		size int
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "empty",
			args: args{
				path: "",
				size: 0,
			},
			wantErr: true,
		},
		{
			name: "invalid path",
			args: args{
				path: "121",
				size: 0,
			},
			wantErr: true,
		},
		{
			name: "valid path",
			args: args{
				path: path,
				size: 10,
			},
			wantErr: false,
		},
		//...
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := run(ctx, log, tt.args.path, tt.args.size, http.DefaultClient); (err != nil) != tt.wantErr {
				t.Errorf("run() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Fuzz_getData(f *testing.F) {
	f.Add("https://myoffice.ru")
	f.Add("https://tinkoff.ru")
	f.Add("https://ya.ru")
	f.Fuzz(func(t *testing.T, u string) {
		_, _, _, _ = getData(new(httpClientStub), u)
	})
}

func Benchmark_run10(b *testing.B) {
	ctx := context.Background()
	log := &logStub{}
	for i := 0; i < b.N; i++ {
		_ = run(ctx, log, path, 10, new(httpClientStub))
	}
}

//...
