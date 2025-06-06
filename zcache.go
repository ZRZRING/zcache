package zcache

import (
	"fmt"
	"log"
	"sync"
	"zcache/singleflight"
	pb "zcache/zcachepb"
)

// Group 是 zcache 最核心的数据结构，负责与用户的交互，并且控制缓存值存储和获取的流程
type Group struct {
	name      string              // 缓存的名称
	getter    Getter              // 缓存未命中时获取源数据的回调
	mainCache cache               // 缓存主体
	peers     PeerPicker          // 节点选择器
	loader    *singleflight.Group // 同一个 key 每个节点只被访问一次，防止缓存击穿
}

type Getter interface {
	Get(key string) ([]byte, error)
}

// GetterFunc 接口型函数，实现了 Getter 接口，使用时可以传入一个结构体或函数
// 因为我们并不知道应该怎么获取数据源，所以将这个操作交给用户来实现
type GetterFunc func(key string) ([]byte, error)

func (f GetterFunc) Get(key string) ([]byte, error) {
	return f(key)
}

var (
	mu     sync.RWMutex
	groups = make(map[string]*Group)
)

// NewGroup 创建一个 Group 实例
func NewGroup(name string, cacheBytes int64, getter Getter) *Group {
	if getter == nil {
		panic("getter 为空，缺少获取数据源的回调函数")
	}
	mu.Lock()
	defer mu.Unlock()
	g := &Group{
		name:      name,
		getter:    getter,
		mainCache: cache{cacheBytes: cacheBytes},
		loader:    &singleflight.Group{},
	}
	groups[name] = g
	return g
}

// GetGroup 用名称获取 Group 实例
func GetGroup(name string) *Group {
	mu.RLock()
	defer mu.RUnlock()
	g := groups[name]
	return g
}

// Get 从缓存中获取数据，如果不存在则调用 load 方法从数据源获取数据
func (g *Group) Get(key string) (ByteView, error) {
	if key == "" {
		return ByteView{}, fmt.Errorf("key 字段为空")
	}
	if v, hit := g.mainCache.get(key); hit {
		log.Println("[zcache] 缓存命中")
		return v, nil
	}
	return g.load(key)
}

// RegisterPeers 注册远程节点选择器
func (g *Group) RegisterPeers(peers PeerPicker) {
	if g.peers != nil {
		panic("重复注册节点")
	}
	g.peers = peers
}

func (g *Group) load(key string) (value ByteView, err error) {
	viewi, err := g.loader.Do(key, func() (interface{}, error) {
		if g.peers != nil {
			if peer, ok := g.peers.PickPeer(key); ok {
				if value, err = g.getFromPeer(peer, key); err == nil {
					return value, nil
				}
				log.Println("[zcache] 远程节点获取数据失败", err)
			}
		}
		return g.getLocally(key)
	})
	if err == nil {
		return viewi.(ByteView), nil
	}
	return
}

func (g *Group) getLocally(key string) (ByteView, error) {
	bytes, err := g.getter.Get(key)
	if err != nil {
		return ByteView{}, err
	}
	value := ByteView{b: cloneBytes(bytes)}
	g.populateCache(key, value)
	return value, nil
}

func (g *Group) getFromPeer(peer PeerGetter, key string) (ByteView, error) {
	req := &pb.Request{
		Group: g.name,
		Key:   key,
	}
	res := &pb.Response{}
	err := peer.Get(req, res)
	if err != nil {
		return ByteView{}, err
	}
	return ByteView{b: res.Value}, nil
}

func (g *Group) populateCache(key string, value ByteView) {
	g.mainCache.add(key, value)
}
