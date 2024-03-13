package apollo

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"os"
	"testing"
	"time"

	agollo "github.com/philchia/agollo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sliveryou/micro-pkg/apollo/internal/mockserver"
)

var addr = ":18080"

func TestMain(m *testing.M) {
	setup()
	code := m.Run()
	teardown()
	os.Exit(code)
}

func setup() {
	go mockserver.Run(addr)
	// wait for mock server to run
	time.Sleep(5 * time.Millisecond)
}

func teardown() {
	mockserver.Close()
}

func getApollo() (*Apollo, error) {
	c := Config{
		AppID:           "SampleApp",
		Cluster:         "default",
		NameSpaceNames:  []string{"application", "client.json", "service.yaml"},
		CacheDir:        "../testdata",
		MetaAddr:        "localhost" + addr,
		AccessKeySecret: "",
	}

	return NewApollo(c)
}

func TestApolloStart(t *testing.T) {
	client, err := getApollo()
	require.NoError(t, err)
	defer client.Stop()
	defer os.Remove(client.GetDumpFileName())

	mockserver.Set("application", "key", "value")
	updates := make(chan struct{}, 1)
	defer close(updates)

	client.OnUpdate(func(event *agollo.ChangeEvent) {
		updates <- struct{}{}
	})

	select {
	case <-updates:
	case <-time.After(time.Millisecond * 30000):
	}

	val := client.GetString("key")
	assert.Equal(t, "value", val)
	keys := client.GetAllKeys()
	assert.Len(t, keys, 1)
	releaseKey := client.GetReleaseKey()
	assert.Empty(t, releaseKey)

	mockserver.Set("application", "key", "newvalue")
	select {
	case <-updates:
	case <-time.After(time.Millisecond * 30000):
	}

	val = client.GetNamespaceValue("application", "key")
	assert.Equal(t, "newvalue", val)
	content := client.GetPropertiesContent()
	assert.Equal(t, "key=newvalue\n", content)
	keys = client.GetAllKeys()
	assert.Len(t, keys, 1)

	mockserver.Delete("application", "key")
	select {
	case <-updates:
	case <-time.After(time.Millisecond * 30000):
	}

	val = client.GetNamespaceValue("application", "key")
	assert.Empty(t, val)
	keys = client.GetAllKeys()
	assert.Empty(t, keys)

	mockserver.Set("client.json", "content", `{"name":"agollo"}`)
	select {
	case <-updates:
	case <-time.After(time.Millisecond * 30000):
	}

	val = client.GetNamespaceContent("client.json")
	assert.Equal(t, `{"name":"agollo"}`, val)
	err = client.SubscribeToNamespaces("new_namespace.json")
	require.NoError(t, err)

	mockserver.Set("new_namespace.json", "content", "1")
	select {
	case <-updates:
	case <-time.After(time.Millisecond * 30000):
	}

	val = client.GetNamespaceContent("new_namespace.json")
	assert.Equal(t, "1", val)

	content = `TestYaml:
  TestStrings: [ "abc", "123" ] # TestStrings
  TestString: "test string" # TestString
`
	mockserver.Set("service.yaml", "content", content)
	select {
	case <-updates:
	case <-time.After(time.Millisecond * 30000):
	}

	val = client.GetNamespaceContent("service.yaml")
	assert.Equal(t, content, val)
}

func TestUnmarshalYaml(t *testing.T) {
	client, err := getApollo()
	require.NoError(t, err)
	defer client.Stop()
	defer os.Remove(client.GetDumpFileName())

	content := `TestYaml:
  TestStrings: [ "abc", "123" ] # TestStrings
  TestString: "test string" # TestString
  TestInt: 123
  TestInt32: 456
  TestInt64: 789
`
	mockserver.Set("service.yaml", "content", content)
	updates := make(chan struct{}, 1)
	defer close(updates)

	client.OnUpdate(func(event *agollo.ChangeEvent) {
		updates <- struct{}{}
	})

	select {
	case <-updates:
	case <-time.After(time.Millisecond * 30000):
	}

	val := client.GetNamespaceContent("service.yaml")
	assert.Equal(t, content, val)

	type TestYaml struct {
		TestStrings []string
		TestString  string
		TestInt     int
		TestInt32   int32
		TestInt64   int64
	}
	type TestConfig struct {
		TestYaml TestYaml
	}
	obj := TestConfig{}
	err = UnmarshalYaml(val, &obj)
	require.NoError(t, err)
	assert.Len(t, obj.TestYaml.TestStrings, 2)
	assert.Equal(t, []string{"abc", "123"}, obj.TestYaml.TestStrings)
	assert.Equal(t, "test string", obj.TestYaml.TestString)
	assert.Equal(t, 123, obj.TestYaml.TestInt)
	assert.Equal(t, int32(456), obj.TestYaml.TestInt32)
	assert.Equal(t, int64(789), obj.TestYaml.TestInt64)

	obj = TestConfig{TestYaml: TestYaml{TestStrings: []string{"789", "456", "123", "000"}, TestString: "string", TestInt: 1, TestInt32: 2, TestInt64: 3}}
	err = UnmarshalYaml(val, &obj)
	require.NoError(t, err)
	assert.Len(t, obj.TestYaml.TestStrings, 4)
	assert.Equal(t, []string{"abc", "123", "123", "000"}, obj.TestYaml.TestStrings)
	assert.Equal(t, "test string", obj.TestYaml.TestString)
	assert.Equal(t, 123, obj.TestYaml.TestInt)
	assert.Equal(t, int32(456), obj.TestYaml.TestInt32)
	assert.Equal(t, int64(789), obj.TestYaml.TestInt64)

	obj = TestConfig{TestYaml: TestYaml{TestStrings: []string{"789", "456", "123", "000"}, TestString: "string", TestInt: 1, TestInt32: 2, TestInt64: 3}}
	err = UnmarshalYaml(val, &obj, true)
	require.NoError(t, err)
	assert.Len(t, obj.TestYaml.TestStrings, 2)
	assert.Equal(t, []string{"abc", "123"}, obj.TestYaml.TestStrings)
	assert.Equal(t, "test string", obj.TestYaml.TestString)
	assert.Equal(t, 123, obj.TestYaml.TestInt)
	assert.Equal(t, int32(456), obj.TestYaml.TestInt32)
	assert.Equal(t, int64(789), obj.TestYaml.TestInt64)

	fileName := client.GetDumpFileName()
	b, err := os.ReadFile(fileName)
	require.NoError(t, err)

	m := map[string]map[string]string{}
	err = gob.NewDecoder(bytes.NewReader(b)).Decode(&m)
	require.NoError(t, err)
	fmt.Println(m)
}
