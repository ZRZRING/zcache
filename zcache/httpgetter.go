package zcache

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
)

// PeerGetter 是一个接口，用于从远程节点获取缓存值
// 每个远程节点都实现了这个接口，用于获取缓存值

type PeerGetter interface {
	Get(group, key string) ([]byte, error)
}

type httpGetter struct {
	baseURL string
}

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

// 静态类型检查
var _ PeerGetter = (*httpGetter)(nil)
