package client

import (
	"math/rand"
	"time"

	"github.com/Mrs4s/MiraiGo/client/pb/oidb"
	"github.com/Mrs4s/MiraiGo/message"
	"github.com/Mrs4s/MiraiGo/protocol/packets"
	"github.com/Mrs4s/MiraiGo/utils"
	"github.com/pkg/errors"
	"google.golang.org/protobuf/proto"
)

type musicTypeInfo struct {
	appID       uint64
	appType     uint32
	platform    uint32
	sdkVersion  string
	packageName string
	signature   string
}

var musicType = [...]musicTypeInfo{
	{ // QQMusic
		appID:       100497308,
		appType:     1,
		platform:    1,
		sdkVersion:  "0.0.0",
		packageName: "com.tencent.qqmusic",
		signature:   "cbd27cd7c861227d013a25b2d10f0799",
	},
	{ // NeteaseMusic
		appID:       100495085,
		appType:     1,
		platform:    1,
		sdkVersion:  "0.0.0",
		packageName: "com.netease.cloudmusic",
		signature:   "da6b069da1e2982db3e386233f68d76d",
	},
	{ // MiguMusic
		appID:       1101053067,
		appType:     1,
		platform:    1,
		sdkVersion:  "0.0.0",
		packageName: "cmccwm.mobilemusic",
		signature:   "6cdc72a439cef99a3418d2a78aa28c73",
	},
	{ // KugouMusic
		appID:       205141,
		appType:     1,
		platform:    1,
		sdkVersion:  "0.0.0",
		packageName: "com.kugou.android",
		signature:   "fe4a24d80fcf253a00676a808f62c2c6",
	},
	{ // KuwoMusic
		appID:       100243533,
		appType:     1,
		platform:    1,
		sdkVersion:  "0.0.0",
		packageName: "cn.kuwo.player",
		signature:   "bf9ff4ffb4c558a34ee3fd52c223ebf5",
	},
}

// SendGroupMusicShare 发送群聊音乐卡片
func (c *QQClient) SendGroupMusicShare(target int64, msg *message.MusicShareElement) (*message.GroupMessage, error) {
	ch := make(chan *message.GroupMessage)
	eid := utils.RandomString(6)
	c.onGroupMessageReceipt(eid, func(c *QQClient, e *groupMessageReceiptEvent) {
		for _, elem := range e.Msg.Elements {
			if elem.Type() == message.LightApp || elem.Type() == message.Service {
				ch <- e.Msg
			}
		}
	})
	defer c.onGroupMessageReceipt(eid)
	_, _ = c.sendAndWait(c.buildRichMsgSendingPacket(target, msg, 1)) // rsp is empty chunk
	select {
	case ret := <-ch:
		return ret, nil
	case <-time.After(time.Second * 5):
		return nil, errors.New("timeout")
	}
}

// SendFriendMusicShare 发送好友音乐卡片
func (c *QQClient) SendFriendMusicShare(target int64, msg *message.MusicShareElement) {
	_, _ = c.sendAndWait(c.buildRichMsgSendingPacket(target, msg, 0))
}

// OidbSvc.0xb77_9
func (c *QQClient) buildRichMsgSendingPacket(target int64, msg *message.MusicShareElement, sendType uint32) (uint16, []byte) {
	seq := c.nextSeq()
	tp := musicType[msg.MusicType] // MusicType
	body := &oidb.DB77ReqBody{
		AppId:   tp.appID,
		AppType: tp.appType,
		MsgStyle: func() uint32 {
			if msg.MusicUrl == "" {
				return 0
			}
			return 4
		}(),
		ClientInfo: &oidb.DB77ClientInfo{
			Platform:           tp.platform,
			SdkVersion:         tp.sdkVersion,
			AndroidPackageName: tp.packageName,
			AndroidSignature:   tp.signature,
		},
		ExtInfo:  &oidb.DB77ExtInfo{MsgSeq: rand.Uint64()},
		SendType: sendType,
		RecvUin:  uint64(target),
		RichMsgBody: &oidb.DB77RichMsgBody{
			Title:      msg.Title,
			Summary:    msg.Summary,
			Brief:      msg.Brief,
			Url:        msg.Url,
			PictureUrl: msg.PictureUrl,
			MusicUrl:   msg.MusicUrl,
		},
	}
	b, _ := proto.Marshal(body)
	payload := c.packOIDBPackage(2935, 9, b)
	packet := packets.BuildUniPacket(c.Uin, seq, "OidbSvc.0xb77_9", 1, c.OutGoingPacketSessionId, EmptyBytes, c.sigInfo.d2Key, payload)
	return seq, packet
}
