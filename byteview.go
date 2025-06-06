package zcache

// ByteView 只读数据结构，用于表示缓存值
type ByteView struct {
	b []byte // 存储真实的缓存值
}

// Len 实现 Value 接口，返回其所占的内存大小
func (v ByteView) Len() int {
	return len(v.b)
}

func (v ByteView) String() string {
	return string(v.b)
}

func (v ByteView) At(i int) byte {
	return v.b[i]
}

// ByteSlice 返回一个拷贝，防止缓存值被外部程序修改
func (v ByteView) ByteSlice() []byte {
	return cloneBytes(v.b)
}

// 拷贝缓存值，防止缓存值被外部程序修改
func cloneBytes(b []byte) []byte {
	c := make([]byte, len(b))
	copy(c, b)
	return c
}
