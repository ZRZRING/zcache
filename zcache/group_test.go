package zcache

import (
	"fmt"
	"log"
	"testing"
)

var db = map[string]string{
	"Tom":  "630",
	"Jack": "589",
	"Sam":  "567",
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
		if msg, err := z.Get(k); err != nil || msg.String() != v {
			t.Fatalf("failed to get value, err:%v", err)
		}
		if _, err := z.Get(k); err != nil || loadCounts[k] > 1 {
			t.Fatalf("cache %s miss", k)
		}
	}
	if msg, err := z.Get("Unknown"); err == nil {
		t.Fatalf("the value of unknown should be empty, but got %s", msg.String())
	}
}
