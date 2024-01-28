package openapi

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNormalizeURL(t *testing.T) {
	cases := []struct {
		url    string
		expect string
	}{
		{url: "test.com/api", expect: "http://test.com/api"},
		{url: "test.cn", expect: "http://test.cn"},
		{url: "https://test.com", expect: "https://test.com"},
		{url: "http://test.com", expect: "http://test.com"},
		{url: "", expect: ""},
		{url: "test", expect: "http://test"},
	}

	for _, c := range cases {
		got := NormalizeURL(c.url)
		assert.Equal(t, c.expect, got)
	}
}

func TestNormalizeNamespace(t *testing.T) {
	cases := []struct {
		namespace string
		expect    string
	}{
		{namespace: "", expect: ""},
		{namespace: "application", expect: "application"},
		{namespace: "application.properties", expect: "application"},
		{namespace: "test.properties", expect: "test"},
		{namespace: "contract-service.yaml", expect: "contract-service.yaml"},
		{namespace: "proof-service.yml", expect: "proof-service.yml"},
		{namespace: "gateway.json", expect: "gateway.json"},
		{namespace: "test.go", expect: "test.go"},
		{namespace: ".go", expect: ".go"},
		{namespace: "go", expect: "go"},
		{namespace: ".", expect: "."},
	}

	for _, c := range cases {
		got := NormalizeNamespace(c.namespace)
		assert.Equal(t, c.expect, got)
	}
}
