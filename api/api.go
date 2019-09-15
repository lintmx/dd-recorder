package api

import (
	"net/url"
)

var platformNameMap = map[string]string{
	"live.bilibili.com": "哔哩哔哩",
	"www.youtube.com":   "YouTube",
}

// LiveAPI interface
type LiveAPI interface {
	GetLiveURL() string
	RefreshLiveInfo() error
	GetLiveStatus() bool
	GetPlatformName() string
	GetTitle() string
	GetAuthor() string
	GetLiveID() string
	GetStreamURLs() ([]StreamURL, error)
	GetDanmaku(chan struct{}) (<-chan *DanmakuMessage, error)
}

// BaseAPI live info
type BaseAPI struct {
	liveURL    *url.URL
	liveID     string
	liveTitle  string
	liveAuthor string
	liveStatus bool
}

// StreamURL store stream info
type StreamURL struct {
	PlayURL  url.URL
	FileType string
}

// DanmakuMessage store danmaku msg
type DanmakuMessage struct {
	Content  string
	SendTime int64
	Type     uint8
	UserName string
}

// GetLiveURL get live url
func (b *BaseAPI) GetLiveURL() string {
	return b.liveURL.String()
}

// GetLiveStatus get live status
func (b *BaseAPI) GetLiveStatus() bool {
	return b.liveStatus
}

// GetPlatformName return a name for live platform
func (b *BaseAPI) GetPlatformName() string {
	return platformNameMap[b.liveURL.Host]
}

// GetTitle return live title
func (b *BaseAPI) GetTitle() string {
	return b.liveTitle
}

// GetAuthor return live author
func (b *BaseAPI) GetAuthor() string {
	return b.liveAuthor
}

// GetLiveID return live id
func (b *BaseAPI) GetLiveID() string {
	return b.liveID
}

// Check select api
func Check(url *url.URL) LiveAPI {
	base := &BaseAPI{
		liveURL: url,
	}

	// switch live api
	switch url.Host {
	case "www.youtube.com":
		return NewYouTubeLive(base)
	case "live.bilibili.com":
		return NewBilibiliLive(base)
	}

	return nil
}
