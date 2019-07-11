package api

import (
	"fmt"
	"net/url"
	"regexp"
	"time"

	"github.com/lintmx/dd-recorder/utils"
	"github.com/tidwall/gjson"
	"go.uber.org/zap"
)

/**
Thanks:
- https://github.com/ytdl-org/youtube-dl/blob/master/youtube_dl/extractor/youtube.py
- https://github.com/streamlink/streamlink/blob/master/src/streamlink/plugins/youtube.py
*/

var (
	youtubeLiveURL = "https://www.youtube.com/channel/%s/live"
	youtubeChatURL = "https://www.youtube.com/live_chat?is_popout=1&v=%s"
	youtubeChatAPI = "https://www.youtube.com/live_chat/get_live_chat?continuation=%s&pbj=1"
	userAgent      = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_14_4) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/73.0.3683.103 Safari/537.36"

	initInvalidData = "contents.liveChatRenderer.continuations.0.invalidationContinuationData"
	initTimeoutData = "contents.liveChatRenderer.continuations.0.timedContinuationData"
	initMessageData = "contents.liveChatRenderer.actions.#.addChatItemAction.item.liveChatTextMessageRenderer"

	continuationInvalidData = "response.continuationContents.liveChatContinuation.continuations.0.invalidationContinuationData"
	continuationTimeoutData = "response.continuationContents.liveChatContinuation.continuations.0.timedContinuationData"
	continuationMessageData = "response.continuationContents.liveChatContinuation.actions.#.addChatItemAction.item.liveChatTextMessageRenderer"
)

// YouTubeLive youtube live api
type YouTubeLive struct {
	BaseAPI
	videoID string
}

// NewYouTubeLive return a youtubeLive struct
func NewYouTubeLive(base *BaseAPI) *YouTubeLive {
	youtubeLive := YouTubeLive{
		BaseAPI: *base,
	}
	regexURL := regexp.MustCompile(`^(?:https?:\/\/)?www\.youtube\.com\/channel\/([^\/]+)(?:[\/])?(?:live)?`)
	if result := regexURL.FindStringSubmatch(youtubeLive.GetLiveURL()); result != nil {
		youtubeLive.liveID = result[1]
		if err := youtubeLive.RefreshLiveInfo(); err != nil {
			zap.L().Error("Init Live API", zap.String("url", youtubeLive.GetLiveURL()))
			return nil
		}
		return &youtubeLive
	}

	return nil
}

// get youtube live page js
func (y *YouTubeLive) getLiveInfo() (string, error) {
	body, err := utils.HTTPGet(fmt.Sprintf(youtubeLiveURL, y.liveID))

	if err != nil {
		return "", fmt.Errorf("Http Error - youtubeLiveURL - %s", err.Error())
	} else if body == "" {
		return "", fmt.Errorf("youtubeLiveURL download failed")
	}

	liveStatusRegex := []*regexp.Regexp{
		regexp.MustCompile(`;ytplayer\.config\s*=\s*({.+?});ytplayer`),
		regexp.MustCompile(`;ytplayer\.config\s*=\s*({.+?});`),
	}

	for _, re := range liveStatusRegex {
		data := re.FindStringSubmatch(body)

		if len(data) >= 2 {
			return data[1], nil
		}
	}

	// TODO: get live info by video id

	return "", fmt.Errorf("youtubeLive get live info error")
}

func (y *YouTubeLive) getChatInfo() (string, error) {
	return "", nil
}

// RefreshLiveInfo refresh live info
func (y *YouTubeLive) RefreshLiveInfo() error {
	liveInfo, err := y.getLiveInfo()
	liveData := gjson.Get(liveInfo, "args.player_response").String()

	if err != nil {
		return err
	}

	status := gjson.Get(liveData, "playabilityStatus.status")
	if status.Exists() && status.String() == "OK" {
		y.liveStatus = true
	} else {
		y.liveStatus = false
	}

	y.liveAuthor = gjson.Get(liveInfo, "args.author").String()
	y.liveTitle = gjson.Get(liveInfo, "args.title").String()
	y.videoID = gjson.Get(liveInfo, "args.video_id").String()

	return nil
}

// GetStreamURLs return live stream url map
func (y *YouTubeLive) GetStreamURLs() ([]StreamURL, error) {
	streamURLs := []StreamURL{}
	liveInfo, err := y.getLiveInfo()

	if err != nil {
		return streamURLs, err
	}

	liveData := gjson.Get(liveInfo, "args.player_response").String()

	if hlsURL, err := url.Parse(gjson.Get(liveData, "streamingData.hlsManifestUrl").String()); err == nil {
		streamURLs = append(streamURLs, StreamURL{
			PlayURL:  *hlsURL,
			FileType: "ts",
		})
	}

	// gjson.Get(liveData, "streamingData.adaptiveFormats.#.url").ForEach(func(key, value gjson.Result) bool {
	// 	liveURL, err := url.Parse(value.String())

	// 	if err != nil {
	// 		return true
	// 	}

	// 	streamURLs = append(streamURLs, StreamURL{
	// 		PlayURL:  *liveURL,
	// 		FileType: "mp4",
	// 	})

	// 	return true
	// })

	return streamURLs, nil
}

// GetDanmaku push danmaku in chan
func (y *YouTubeLive) GetDanmaku(done chan struct{}) (<-chan *DanmakuMessage, error) {
	msgChan := make(chan *DanmakuMessage)
	header := make(map[string]string)
	header["User-Agent"] = userAgent
	re := regexp.MustCompile(`window\[\"ytInitialData\"\]\s*=\s*({.+?});\s*\<\/script\>`)

	go func() {
		defer close(msgChan)
		for {
		DanmakuRestart:
			body, err := utils.HTTPGetWithHeader(
				fmt.Sprintf(youtubeChatURL, y.videoID),
				header,
			)
			if err != nil {
				continue
			}

			data := re.FindStringSubmatch(body)
			if len(data) < 2 {
				continue
			}

			var continuation string
			var timeOutMs int64
			if continuationData := gjson.Get(data[1], initInvalidData); continuationData.Exists() {
				timeOutMs = continuationData.Get("timeoutMs").Int()
				continuation = continuationData.Get("continuation").String()
			} else if continuationData := gjson.Get(data[1], initTimeoutData); continuationData.Exists() {
				timeOutMs = continuationData.Get("timeoutMs").Int()
				continuation = continuationData.Get("continuation").String()
			} else {
				continue
			}

			gjson.Get(data[1], initMessageData).ForEach(func(key, value gjson.Result) bool {
				msgChan <- &DanmakuMessage{
					Content:  value.Get("message.runs.0.text").String(),
					SendTime: value.Get("timestampUsec").Int() / 1e6,
					Type:     1,
					UserName: value.Get("authorName.simpleText").String(),
				}
				return true
			})

			for {
				select {
				case <-done:
					return
				case <-time.After(time.Duration(timeOutMs) * time.Millisecond):
					body, err := utils.HTTPGetWithHeader(
						fmt.Sprintf(youtubeChatAPI, continuation),
						header,
					)
					if err != nil {
						goto DanmakuRestart
					}

					if continuationData := gjson.Get(body, continuationInvalidData); continuationData.Exists() {
						timeOutMs = continuationData.Get("timeoutMs").Int()
						continuation = continuationData.Get("continuation").String()
					} else if continuationData := gjson.Get(body, continuationTimeoutData); continuationData.Exists() {
						timeOutMs = continuationData.Get("timeoutMs").Int()
						continuation = continuationData.Get("continuation").String()
					} else {
						goto DanmakuRestart
					}

					gjson.Get(body, continuationMessageData).ForEach(func(key, value gjson.Result) bool {
						msgChan <- &DanmakuMessage{
							Content:  value.Get("message.runs.0.text").String(),
							SendTime: value.Get("timestampUsec").Int() / 1e6,
							Type:     1,
							UserName: value.Get("authorName.simpleText").String(),
						}
						return true
					})
				}
			}
		}
	}()

	return msgChan, nil
}
