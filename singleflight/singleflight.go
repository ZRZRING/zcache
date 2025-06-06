package singleflight

import "sync"

// call 代表正在进行中，或已经结束的请求
type call struct {
	wg  sync.WaitGroup // 避免重入
	val interface{}
	err error
}

// Group 主数据结构，管理不同 key 的请求
type Group struct {
	mu sync.Mutex // 保护 m 不被并发读写
	m  map[string]*call
}

// Do 针对相同的 key，无论 Do 被调用多少次，函数 fn 都只会被调用一次，等待 fn 调用结束返回返回值或错误
func (g *Group) Do(key string, fn func() (interface{}, error)) (interface{}, error) {
	g.mu.Lock()
	if g.m == nil {
		g.m = make(map[string]*call)
	}
	if c, ok := g.m[key]; ok {
		g.mu.Unlock()
		// 请求进行中，等待
		c.wg.Wait()
		// 请求结束，返回结果
		return c.val, c.err
	}
	// 第一次，请求还没有进行，准备发起请求
	c := new(call)
	// 发起请求前加锁
	c.wg.Add(1)
	// 添加到 g.m，表明 key 已经有对应的请求在处理
	g.m[key] = c
	g.mu.Unlock()

	// 调用 fn，发起请求
	c.val, c.err = fn()
	// 请求结束
	c.wg.Done()

	g.mu.Lock()
	// 更新 g.m
	delete(g.m, key)
	g.mu.Unlock()

	return c.val, c.err
}
