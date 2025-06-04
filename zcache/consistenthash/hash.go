package consistenthash

import (
	"hash/crc32"
	"sort"
	"strconv"
)

// Hash 定义为 []byte 到 uint32 的 映射函数
type Hash func(data []byte) uint32

type Map struct {
	hash     Hash              // 哈希函数
	replicas int               // 实节点 对应 虚拟节点 的 倍数
	keys     []uint32          // 所有 虚拟节点 的 键值
	hashMap  map[uint32]string // 虚拟节点 对 实节点 的 映射
}

// New 创建一个 一致性哈希 实例
// replicas 实节点 对应 虚拟节点 的 倍数
// fn 哈希函数，默认为 crc32.ChecksumIEEE
func New(replicas int, fn Hash) *Map {
	m := &Map{
		hash:     fn,
		replicas: replicas,
		hashMap:  make(map[uint32]string),
	}
	if m.hash == nil {
		m.hash = crc32.ChecksumIEEE
	}
	return m
}

// Add 添加 实节点
func (m *Map) Add(keys ...string) {
	for _, key := range keys {
		for i := range m.replicas {
			hash := m.hash([]byte(strconv.Itoa(i) + key))
			m.keys = append(m.keys, hash)
			m.hashMap[hash] = key
		}
	}
	sort.Slice(m.keys, func(i, j int) bool {
		return m.keys[i] < m.keys[j]
	})
}

func (m *Map) Get(key string) string {
	if len(m.keys) == 0 {
		return ""
	}
	hash := m.hash([]byte(key))
	idx := sort.Search(len(m.keys), func(i int) bool {
		return m.keys[i] >= hash
	})
	idx = idx % len(m.keys)
	return m.hashMap[m.keys[idx]]
}
