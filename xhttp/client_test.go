package xhttp

import (
	"context"
	"io"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUrlValuesAdd(t *testing.T) {
	values := make(url.Values)
	values.Add("bankcard", "bankcard1")
	values.Add("idcard", "idcard")
	values.Add("mobile", "mobile")
	values.Add("name", "name")
	values.Add("bankcard", "bankcard2")

	assert.Equal(t, "bankcard=bankcard1&bankcard=bankcard2&idcard=idcard&mobile=mobile&name=name", values.Encode())
}

func TestUrlValuesSet(t *testing.T) {
	values := make(url.Values)
	values.Set("bankcard", "bankcard1")
	values.Set("idcard", "idcard")
	values.Set("mobile", "mobile")
	values.Set("name", "name")
	values.Set("bankcard", "bankcard2")

	assert.Equal(t, "bankcard=bankcard2&idcard=idcard&mobile=mobile&name=name", values.Encode())
}

func TestClient(t *testing.T) {
	type args struct {
		method string
		url    string
		header map[string]string
		body   io.Reader
	}
	cases := []struct {
		title string
		args  args
	}{
		{
			title: "do get ip",
			args: args{
				method: "GET",
				url:    "https://www.httpbin.org/ip",
				header: map[string]string{
					"Accept-Language": "zh-CN,zh;q=0.9,en;q=0.8",
					"User-Agent":      "Go-HTTP-Request",
				},
				body: nil,
			},
		},
		{
			title: "do get method",
			args: args{
				method: "GET",
				url:    "https://www.httpbin.org/get",
				header: map[string]string{
					"Accept-Language": "zh-CN,zh;q=0.9,en;q=0.8",
					"User-Agent":      "Go-HTTP-Request",
				},
				body: nil,
			},
		},
		{
			title: "do post method",
			args: args{
				method: "POST",
				url:    "https://www.httpbin.org/post",
				header: map[string]string{
					"Accept-Language": "zh-CN,zh;q=0.9,en;q=0.8",
					"User-Agent":      "Go-HTTP-Request",
				},
				body: strings.NewReader("a=b&c=d"),
			},
		},
	}

	client := NewClient()
	for _, c := range cases {
		t.Run(c.title, func(t *testing.T) {
			req, err := client.GetRequest(context.Background(), c.args.method, c.args.url, c.args.header, c.args.body)
			if err == nil {
				_, resp, err := client.GetResponse(req)
				if err == nil {
					t.Log(string(resp))
				}
			}
		})
	}
}
