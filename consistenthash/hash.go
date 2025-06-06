package consistenthash

import (
	"hash/crc32"
	"sort"
	"strconv"
)

// Hash 定义为 []byte 到 uint32 的映射函数
type Hash func(data []byte) uint32

type Map struct {
	hash     Hash              // 哈希函数
	replicas int               // 实节点对应虚拟节点的倍数
	keys     []uint32          // 所有虚拟节点的键值
	hashMap  map[uint32]string // 虚拟节点对实节点的映射
}

// New 创建一个一致性哈希实例
// replicas 实节点对应虚拟节点的倍数
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

// Add 添加实节点
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

// Get 根据 key 获取最接近的实节点
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
