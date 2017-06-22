// 发送消息

package easymob

import (
	"encoding/json"
	"fmt"
	"io"
	"time"
)

type TargetType string

// 发送目标用户类型
const (
	TargetUser   TargetType = "users"      // 给用户发消息
	TargetGroups TargetType = "chatgroups" // 给群发消息
	TargetRoom   TargetType = "chatrooms"  // 给聊天室发消息
)

// Message 消息
type Message struct {
	TargetType TargetType  `json:"target_type"`
	Target     []string    `json:"target"`
	From       string      `json:"from,omitempty"`
	Content    interface{} `json:"msg"`
	Ext        interface{} `json:"ext,omitempty"`
}

// TxtMessage 文本消息
type TxtMessage struct {
	Type string `json:"type"`
	Msg  string `json:"msg"`
}

// CmdMessage 透传消息
type CmdMessage struct {
	Type   string `json:"type"`
	Action string `json:"action"`
}

// ImgMessage 图片消息
type ImgMessage struct {
	Type     string `json:"type"`
	URL      string `json:"url"`
	FileName string `json:"filename"`
	Secret   string `json:"secret"`
	Size     struct {
		W int `json:"width"`
		H int `json:"height"`
	} `json:"size"`
}

// SendMessage 发送消息
func (c *Client) SendMessage(m Message) error {
	path := c.makePath("messages", nil, nil)
	var reply commonReply
	return c.httpcall(path, "POST", m, &reply, true)
}

// CreateTxtMessage 创建文本消息
func (c *Client) CreateTxtMessage(content string) TxtMessage {
	return TxtMessage{Type: "txt", Msg: content}
}

// CreateCmdMessage 创建透传消息
func (c *Client) CreateCmdMessage(action string) CmdMessage {
	return CmdMessage{Type: "cmd", Action: action}
}

// CreateImgMessage 创建图片消息
func (c *Client) CreateImgMessage(file io.Reader) (msg ImgMessage, err error) {
	entity, err := c.uploadImgAndVoice(file)
	if err != nil {
		return
	}
	msg = ImgMessage{
		URL:      entity.URL,
		Type:     "img",
		FileName: fmt.Sprintf("%d.jpg", time.Now().Unix()),
		Secret:   entity.Secret,
	}
	return
}

// DownloadMessages 下载聊天记录
func (c *Client) DownloadMessages(t time.Time) ([]string, error) {
	path := c.makePath("chatmessages", []string{t.Format("2006010215")}, nil)
	var reply commonReply
	err := c.httpcall(path, "GET", nil, &reply, true)
	if err != nil {
		return []string{}, err
	}
	var result []struct {
		URL string `json:"url"`
	}
	err = json.Unmarshal(reply.Data, &result)
	var files []string
	for _, res := range result {
		files = append(files, res.URL)
	}
	return files, err
}
