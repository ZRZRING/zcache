package zcache

import pb "zcache/zcachepb"

// PeerPicker 定义从其他节点获取数据的接口
type PeerPicker interface {
	PickPeer(key string) (peer PeerGetter, ok bool)
}

// PeerGetter 从远程节点获取缓存值，每个远程节点都实现了这个接口，用于获取缓存值
type PeerGetter interface {
	Get(in *pb.Request, out *pb.Response) error
}
