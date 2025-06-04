package zcache

// ByteView 只读数据结构，用于表示缓存值
// b 用来存储真实的缓存值，选择 byte 类型是为了能够支持任意的数据类型的存储，例如字符串、图片等

type ByteView struct {
	b []byte
}

func (v ByteView) Len() int {
	return len(v.b)
}

func (v ByteView) String() string {
	return string(v.b)
}

func (v ByteView) At(i int) byte {
	return v.b[i]
}

// ByteSlice 方法返回一个拷贝，防止缓存值被外部程序修改
func (v ByteView) ByteSlice() []byte {
	return cloneBytes(v.b)
}

// 拷贝缓存值，防止缓存值被外部程序修改
func cloneBytes(b []byte) []byte {
	c := make([]byte, len(b))
	copy(c, b)
	return c
}
