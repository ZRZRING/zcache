package zcache

import (
	"fmt"
	"google.golang.org/protobuf/proto"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"zcache/consistenthash"
	pb "zcache/zcachepb"
)

const (
	defaultBasePath = "/zcache"
	defaultReplicas = 50
)

// HTTPPool 实现 http 请求的分布式缓存服务 api
type HTTPPool struct {
	self        string                 // 记录地址
	basePath    string                 // 通讯地址
	mu          sync.Mutex             // guards peers and httpGetters
	httpGetters map[string]*httpGetter // keyed by e.g. "http://10.0.0.2:8008"
	peers       *consistenthash.Map
}

// NewHTTPPool 创建一个新的 HTTPPool
func NewHTTPPool(self string) *HTTPPool {
	return &HTTPPool{
		self:     self,
		basePath: defaultBasePath,
	}
}

// Log 打印日志
func (p *HTTPPool) Log(format string, v ...interface{}) {
	log.Printf("[Server %s] %s", p.self, fmt.Sprintf(format, v...))
}

// ServeHTTP 处理 http 请求
func (p *HTTPPool) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !strings.HasPrefix(r.URL.Path, p.basePath) {
		panic("invalid path: " + r.URL.Path)
	}

	p.Log("%s %s", r.Method, r.URL.Path)

	// 路径格式 /<basePath>/<groupName>/<key>
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) != 4 {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	groupName := parts[2]
	key := parts[3]

	group := GetGroup(groupName)
	if group == nil {
		msg := fmt.Sprintf("no such group: %s", groupName)
		http.Error(w, msg, http.StatusNotFound)
		return
	}

	view, err := group.Get(key)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/octet-stream")
	_, err = w.Write(view.ByteSlice())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 将值作为 proto 消息写入响应体
	body, err := proto.Marshal(&pb.Response{Value: view.ByteSlice()})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/octet-stream")
	_, err = w.Write(body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// Set 实例化一致性哈希算法，并且添加了传入的节点
func (p *HTTPPool) Set(peers ...string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.peers = consistenthash.New(defaultReplicas, nil)
	p.peers.Add(peers...)
	p.httpGetters = make(map[string]*httpGetter, len(peers))
	for _, peer := range peers {
		p.httpGetters[peer] = &httpGetter{
			baseURL: peer + p.basePath,
		}
	}
}

// PickPeer 包装了一致性哈希算法的 Get() 方法，根据具体的 key，选择节点，返回节点对应的 HTTP 客户端
func (p *HTTPPool) PickPeer(key string) (PeerGetter, bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if peer := p.peers.Get(key); peer != p.self && peer != "" {
		p.Log("Pick peer %s", peer)
		return p.httpGetters[peer], true
	}
	return nil, false
}

// 静态类型检查
var _ PeerPicker = (*HTTPPool)(nil)

type httpGetter struct {
	baseURL string
}

// Get 实现了 PeerGetter 接口的 Get() 方法
func (h *httpGetter) Get(group, key string) ([]byte, error) {
	u := fmt.Sprintf(
		"%v/%v/%v",
		h.baseURL,
		url.QueryEscape(group),
		url.QueryEscape(key),
	)
	res, err := http.Get(u)
	if err != nil {
		return nil, err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			panic("res.Body.Close() fault!")
		}
	}(res.Body)

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server returned: %v", res.Status)
	}

	bytes, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response body: %v", err)
	}

	return bytes, nil
}

// GetPB 实现了 PeerGetterPB 接口的 Get() 方法
func (h *httpGetter) GetPB(in *pb.Request, out *pb.Response) (err error) {
	u := fmt.Sprintf(
		"%v/%v/%v",
		h.baseURL,
		url.QueryEscape(in.GetGroup()),
		url.QueryEscape(in.GetKey()),
	)
	res, err := http.Get(u)
	if err != nil {
		return
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			panic("res.Body.Close() fault!")
		}
	}(res.Body)

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("server returned: %v", res.Status)
	}

	bytes, err := io.ReadAll(res.Body)
	if err != nil {
		return fmt.Errorf("reading response body: %v", err)
	}

	if err = proto.Unmarshal(bytes, out); err != nil {
		return fmt.Errorf("decoding response body: %v", err)
	}

	return
}

// 静态类型检查
var _ PeerGetter = (*httpGetter)(nil)
