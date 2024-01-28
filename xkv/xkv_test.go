package xkv

import (
	"testing"

	miniredis "github.com/alicebob/miniredis/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/redis"
)

var (
	s1, _ = miniredis.Run()
	s2, _ = miniredis.Run()
)

func TestStore(t *testing.T) {
	type testObj struct {
		Id   int64  `json:"id"`
		Name string `json:"name"`
	}

	runOnCluster(func(s *Store) {
		testKey1 := "cache:test:test_store:id:1"
		t1 := &testObj{Id: 1, Name: "testName"}
		err := s.Write(testKey1, t1)
		require.NoError(t, err)

		t2 := &testObj{}
		isExist, err := s.Read(testKey1, t2)
		require.NoError(t, err)
		assert.True(t, isExist)
		t.Log(t2)

		_, err = s.Del(testKey1)
		require.NoError(t, err)

		testKey2 := "cache:test:test_store:id:2"
		t3 := &testObj{}
		f1 := func() (interface{}, error) {
			return &testObj{Id: 2, Name: "testName2"}, nil
		}
		err = s.ReadOrGet(testKey2, t3, f1)
		require.NoError(t, err)
		t.Log(t3)

		_, err = s.Del(testKey2)
		require.NoError(t, err)

		testKey3 := "cache:test:test_store:id:3"
		t4 := make(map[string]*testObj)
		f2 := func() (interface{}, error) {
			m := make(map[string]*testObj)
			m["1"] = &testObj{Id: 1, Name: "1"}
			m["2"] = &testObj{Id: 2, Name: "2"}
			m["3"] = &testObj{Id: 3, Name: "3"}
			return &m, nil
		}
		err = s.ReadOrGet(testKey3, &t4, f2)
		require.NoError(t, err)
		t.Log(t4)
		for k := range t4 {
			t.Logf("key: %s, value: %+v\n", k, t4[k])
		}

		_, err = s.Del(testKey3)
		require.NoError(t, err)

		testKey4 := "cache:test:test_store:id:4"
		err = s.SetString(testKey4, "test")
		require.NoError(t, err)

		v, err := s.GetDel(testKey4)
		require.NoError(t, err)
		assert.Equal(t, "test", v)

		isExist, err = s.Exists(testKey4)
		require.NoError(t, err)
		assert.False(t, isExist)
	})
}

func runOnCluster(fn func(cluster *Store)) {
	s1.FlushAll()
	s2.FlushAll()

	store := NewStore([]cache.NodeConf{
		{
			RedisConf: redis.RedisConf{
				Host: s1.Addr(),
				Type: redis.NodeType,
			},
			Weight: 100,
		},
		{
			RedisConf: redis.RedisConf{
				Host: s2.Addr(),
				Type: redis.NodeType,
			},
			Weight: 100,
		},
	})

	fn(store)
}
