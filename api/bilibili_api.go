package api

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/lintmx/dd-recorder/utils"
	"github.com/tidwall/gjson"
	"go.uber.org/zap"
	"net/url"
	"regexp"
	"time"
)

// bilibili websocket protocol
const (
	HeaderLen   = 16
	ProtocolVer = 1
	SequenceID  = 1
)

// websocket operation protocol
const (
	OperationTypeHeart   = 2
	OperationTypeMessage = 5
	OperationTypeEnter   = 7
)

var (
	bilibiliRealRoomIDAPI = "https://api.live.bilibili.com/room/v1/Room/room_init?id=%s"
	bilibiliRoomInfoAPI   = "https://api.live.bilibili.com/room/v1/Room/get_info?room_id=%d"
	bilibiliRoomAnchorAPI = "https://api.live.bilibili.com/live_user/v1/UserInfo/get_anchor_in_room?roomid=%d"
	bilibiliPlayURLAPI    = "https://api.live.bilibili.com/room/v1/Room/playUrl?cid=%d"
	bilibiliDanmakuAPI    = "https://api.live.bilibili.com/room/v1/Danmu/getConf?room_id=%d&platform=pc&player=web"
)

// BilibiliLive bilibili live api
type BilibiliLive struct {
	BaseAPI
	roomID int64
}

type danmakuInitMsg struct {
	ClientVer string `json:"clientver"`
	Platform  string `json:"platform"`
	ProtoVer  int    `json:"protover"`
	RoomID    int    `json:"roomid"`
	UID       int    `json:"uid"`
}

// NewBilibiliLive return a bilibililive struct
func NewBilibiliLive(base *BaseAPI) *BilibiliLive {
	bilibiliLive := BilibiliLive{
		BaseAPI: *base,
	}
	regexURL := regexp.MustCompile(`^(?:https?:\/\/)?live\.bilibili\.com\/(\d+)[\/\?\#]?.*$`)
	if result := regexURL.FindStringSubmatch(bilibiliLive.GetLiveURL()); result != nil {
		bilibiliLive.liveID = result[1]
		if err := bilibiliLive.RefreshLiveInfo(); err != nil {
			zap.L().Error("Init Live API", zap.String("url", bilibiliLive.GetLiveURL()))
			return nil
		}

		return &bilibiliLive
	}

	return nil
}

func (b *BilibiliLive) getRealRoomID() error {
	body, err := utils.HTTPGet(fmt.Sprintf(bilibiliRealRoomIDAPI, b.liveID))

	if err != nil {
		return fmt.Errorf("Http Error - bilibiliRealRoomIDAPI - %s", err.Error())
	} else if code := gjson.Get(body, "code"); !code.Exists() {
		return fmt.Errorf("bilibiliRealRoomIDAPI is broken")
	} else if code.Int() != 0 {
		return fmt.Errorf("bilibiliRealRoomIDAPI - %s", gjson.Get(body, "msg").String())
	}

	b.roomID = gjson.Get(body, "data.room_id").Int()

	return nil
}

// RefreshLiveInfo refresh live info
func (b *BilibiliLive) RefreshLiveInfo() error {
	if b.roomID == 0 {
		if err := b.getRealRoomID(); err != nil {
			return err
		}
	}

	// get live title and live status
	body, err := utils.HTTPGet(fmt.Sprintf(bilibiliRoomInfoAPI, b.roomID))

	if err != nil {
		return fmt.Errorf("Http Error - bilibiliRoomInfoAPI - %s", err.Error())
	} else if code := gjson.Get(body, "code"); !code.Exists() {
		return fmt.Errorf("bilibiliRoomInfoAPI is broken")
	} else if code.Int() != 0 {
		return fmt.Errorf("bilibiliRoomInfoAPI - %s", gjson.Get(body, "msg").String())
	}

	status := gjson.Get(body, "data.live_status").Int() == 1

	b.liveStatus = status
	b.liveTitle = gjson.Get(body, "data.title").String()

	// get live author
	body, err = utils.HTTPGet(fmt.Sprintf(bilibiliRoomAnchorAPI, b.roomID))

	if err != nil {
		return fmt.Errorf("Http Error - bilibiliRoomAnchorAPI - %s", err.Error())
	} else if code := gjson.Get(body, "code"); !code.Exists() {
		return fmt.Errorf("bilibiliRoomAnchorAPI is broken")
	} else if code.Int() != 0 {
		return fmt.Errorf("bilibiliRoomAnchorAPI - %s", gjson.Get(body, "msg").String())
	}

	b.liveAuthor = gjson.Get(body, "data.info.uname").String()

	return nil
}

// GetStreamURLs return live stream url map
func (b *BilibiliLive) GetStreamURLs() ([]StreamURL, error) {
	streamURLs := []StreamURL{}
	if b.roomID == 0 {
		if err := b.getRealRoomID(); err != nil {
			return streamURLs, err
		}
	}

	// get live title and live status
	body, err := utils.HTTPGet(fmt.Sprintf(bilibiliPlayURLAPI, b.roomID))

	if err != nil {
		return streamURLs, fmt.Errorf("Http Error - bilibiliPlayURLAPI - %s", err.Error())
	} else if code := gjson.Get(body, "code"); !code.Exists() {
		return streamURLs, fmt.Errorf("bilibiliPlayURLAPI is broken")
	} else if code.Int() != 0 {
		return streamURLs, fmt.Errorf("bilibiliPlayURLAPI - %s", gjson.Get(body, "msg").String())
	}

	gjson.Get(body, "data.durl.#.url").ForEach(func(key, value gjson.Result) bool {
		liveURL, err := url.Parse(value.String())

		if err != nil {
			return true
		}

		streamURL := StreamURL{
			PlayURL:  *liveURL,
			FileType: "ts",
		}

		streamURLs = append(streamURLs, streamURL)

		return true
	})

	return streamURLs, nil
}

// GetDanmaku push danmaku in chan
func (b *BilibiliLive) GetDanmaku(done chan struct{}) (<-chan *DanmakuMessage, error) {
	msgChan := make(chan *DanmakuMessage)

	go func() {
		// heart packet
		heartTicker := time.NewTicker(30 * time.Second)
		defer heartTicker.Stop()
		for {
			DanmakuRestart:
			if b.roomID == 0 {
				if err := b.getRealRoomID(); err != nil {
					continue
				}
			}

			// get danmaku url
			body, err := utils.HTTPGet(fmt.Sprintf(bilibiliDanmakuAPI, b.roomID))

			if err != nil {
				continue
			} else if code := gjson.Get(body, "code"); !code.Exists() {
				continue
			} else if code.Int() != 0 {
				continue
			}

			// get danmaku websocket url
			danmakuURL := &url.URL{}

			gjson.Get(body, "data.host_server_list").ForEach(func(key, value gjson.Result) bool {
				addr := gjson.Parse(value.String())
				if addr.Get("host").Exists() && addr.Get("wss_port").Exists() {
					danmakuURL.Scheme = "wss"
					danmakuURL.Host = fmt.Sprintf("%s:%d", addr.Get("host").String(), addr.Get("wss_port").Int())
					danmakuURL.Path = "/sub"
					return false
				}
				return true
			})

			if danmakuURL == (&url.URL{}) {
				continue
			}

			conn, _, err := websocket.DefaultDialer.Dial(danmakuURL.String(), nil)
			if err != nil {
				continue
			}

			init, _ := json.Marshal(&danmakuInitMsg{
				ClientVer: "1.5.10.1",
				Platform:  "web",
				ProtoVer:  1,
				RoomID:    int(b.roomID),
				UID:       2,
			})

			// enter room packet
			conn.WriteMessage(websocket.BinaryMessage, msgEncode(init, OperationTypeEnter))
			exitChan := make(chan struct{})

			go danmakuReceive(conn, msgChan, exitChan)

			for {
				select {
				case <-done:
					conn.Close()
					select {
					case <-exitChan:
						close(msgChan)
					}
					return
				case <-exitChan:
					goto DanmakuRestart
				case <-heartTicker.C:
					conn.WriteMessage(websocket.BinaryMessage, msgEncode([]byte{}, OperationTypeHeart))
				}
			}
		}
	}()

	return msgChan, nil
}

func danmakuReceive(conn *websocket.Conn, msgChan chan *DanmakuMessage, exitChan chan struct{}) {
	defer close(exitChan)
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			return
		}
		for {
			if len(message) <= 16 {
				break
			}
			bodyLen := binary.BigEndian.Uint32(message[:4])
			operation := binary.BigEndian.Uint32(message[8:12])
			if operation == OperationTypeMessage {
				if msg := danmakuDecode(message[12:bodyLen]); msg != nil {
					msgChan <- msg
				}
			}
			message = append(message[bodyLen:])
		}
	}

}

func msgEncode(body []byte, operation uint32) []byte {
	buffer := bytes.NewBuffer([]byte{})
	binary.Write(buffer, binary.BigEndian, uint32(len(body))+uint32(HeaderLen))
	binary.Write(buffer, binary.BigEndian, uint16(HeaderLen))
	binary.Write(buffer, binary.BigEndian, uint16(ProtocolVer))
	binary.Write(buffer, binary.BigEndian, operation)
	binary.Write(buffer, binary.BigEndian, uint32(SequenceID))
	binary.Write(buffer, binary.BigEndian, body)
	return buffer.Bytes()
}

func danmakuDecode(content []byte) *DanmakuMessage {
	body := gjson.Parse(string(content))

	switch body.Get("cmd").String() {
	case "DANMU_MSG":
		// filter gift danmaku
		if body.Get("info.0.5").Int() == 0 {
			return nil
		}
		return &DanmakuMessage{
			Content:  body.Get("info.1").String(),
			SendTime: body.Get("info.0.4").Int(),
			Type:     1,
			UserName: body.Get("info.2.1").String(),
		}
	}

	return nil
}
