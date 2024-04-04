package peer

import (
	"common/console"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

const (
	DefaultCacheDir = "_cache"
)

type HttpGetter struct {
	baseURL string
}

func NewHttpGetter(baseURL string) Getter {
	return &HttpGetter{baseURL: baseURL}
}

func (h *HttpGetter) Get(key string) ([]byte, error) {
	u := fmt.Sprintf(
		"%s/%v/%v",
		h.baseURL,
		url.QueryEscape(DefaultCacheDir),
		url.QueryEscape(key),
	)

	res, err := http.Get(u)
	defer func() {
		if err == nil {
			if err := res.Body.Close(); err != nil {
				console.Error(err.Error())
			}
		}

	}()
	if err != nil {
		return nil, err
	}

	if res.StatusCode != http.StatusOK {
		bytes, err := io.ReadAll(res.Body)
		if err != nil {
			return nil, fmt.Errorf("status not ok , and get body err")
		}
		return nil, fmt.Errorf("CacheServer returned: [ %v ]  \nerr:\n %s", res.Status, string(bytes))
	}

	bytes, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response body: %v", err)
	}

	return bytes, nil
}
