package zcache

import (
	"fmt"
	"log"
	"reflect"
	"testing"
)

func TestGetter(t *testing.T) {
	var f Getter = GetterFunc(func(key string) ([]byte, error) {
		return []byte(key), nil
	})
	expect := []byte("key")
	if v, _ := f.Get("key"); !reflect.DeepEqual(v, expect) {
		t.Errorf("expect %v, got %v", expect, v)
	}
}

var db = map[string]string{
	"zrzring": "60",
	"qwaszx":  "70",
	"qweasd":  "80",
}

func TestGetGroup(t *testing.T) {
	loadCounts := make(map[string]int, len(db))
	z := NewGroup("scores", 2<<10, GetterFunc(
		func(key string) ([]byte, error) {
			log.Println("[DB] search key", key)
			if v, hit := db[key]; hit {
				if _, hit := loadCounts[key]; !hit {
					loadCounts[key] = 0
				}
				loadCounts[key] += 1
				return []byte(v), nil
			}
			return nil, fmt.Errorf("%s not found", key)
		}))
	for k, v := range db {
		// 在缓存为空的情况下，能够通过回调函数获取到源数据
		if msg, err := z.Get(k); err != nil || msg.String() != v {
			t.Fatalf("failed to get value, err:%v", err)
		}
		// 在缓存已经存在的情况下，是否直接从缓存中获取
		if _, err := z.Get(k); err != nil || loadCounts[k] > 1 {
			t.Fatalf("cache %s miss", k)
		}
	}
	if msg, err := z.Get("Unknown"); err == nil {
		t.Fatalf("the value of unknown should be empty, but got %s", msg.String())
	}
}
