package types

import (
	"strconv"
	"testing"
)

func TestSmsClientPicker(t *testing.T) {
	p := NewSmsClientPicker()
	var keys []string
	for i := 0; i < 5; i++ {
		key := "test-sms-" + strconv.Itoa(i)
		keys = append(keys, key)
		p.Add(key, &MockClient{})
	}

	for i := 0; i < 100; i++ {
		t.Log(p.Pick())
	}

	for _, key := range keys {
		t.Log(p.Get(key))
	}

	p.Remove(keys...)

	for _, key := range keys {
		t.Log(p.Get(key))
	}
}

func TestEmailClientPicker(t *testing.T) {
	p := NewEmailClientPicker()
	var keys []string
	for i := 0; i < 5; i++ {
		key := "test-email-" + strconv.Itoa(i)
		keys = append(keys, key)
		p.Add(key, &MockClient{})
	}

	for i := 0; i < 100; i++ {
		t.Log(p.Pick())
	}

	for _, key := range keys {
		t.Log(p.Get(key))
	}

	p.Remove(keys...)

	for _, key := range keys {
		t.Log(p.Get(key))
	}
}
