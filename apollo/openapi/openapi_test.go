package openapi

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseFormat(t *testing.T) {
	cases := []struct {
		namespace    string
		expectFormat Format
		expectIsOk   bool
	}{
		{namespace: "", expectFormat: "", expectIsOk: false},
		{namespace: "application", expectFormat: FormatProperties, expectIsOk: true},
		{namespace: "application.properties", expectFormat: FormatProperties, expectIsOk: true},
		{namespace: "contract-service.yaml", expectFormat: FormatYAML, expectIsOk: true},
		{namespace: "proof-service.yml", expectFormat: FormatYML, expectIsOk: true},
		{namespace: "gateway.json", expectFormat: FormatJSON, expectIsOk: true},
		{namespace: "test.go", expectFormat: "", expectIsOk: false},
		{namespace: ".go", expectFormat: "", expectIsOk: false},
		{namespace: "go", expectFormat: FormatProperties, expectIsOk: true},
		{namespace: ".", expectFormat: "", expectIsOk: false},
	}

	for _, c := range cases {
		gotFormat, gotIsOk := ParseFormat(c.namespace)
		assert.Equal(t, c.expectFormat, gotFormat)
		assert.Equal(t, c.expectIsOk, gotIsOk)
	}
}

func TestTrimFormat(t *testing.T) {
	cases := []struct {
		namespace string
		expect    string
	}{
		{namespace: "", expect: ""},
		{namespace: "application", expect: "application"},
		{namespace: "application.properties", expect: "application"},
		{namespace: "contract-service.yaml", expect: "contract-service"},
		{namespace: "proof-service.yml", expect: "proof-service"},
		{namespace: "gateway.json", expect: "gateway"},
		{namespace: "test.go", expect: "test"},
		{namespace: ".go", expect: ""},
		{namespace: "go", expect: "go"},
		{namespace: ".", expect: ""},
	}

	for _, c := range cases {
		got := TrimFormat(c.namespace)
		assert.Equal(t, c.expect, got)
	}
}
