// 用户集成

package easymob

import (
	"bytes"
	"encoding/json"
)

// User 环信用户
type User struct {
	Name     string `json:"username"`
	Password string `json:"password"`
	NickName string `json:"nickname,omitempty"`
}

type commonReply struct {
	Action    string          `json:"action"`
	App       string          `json:"application"`
	Params    interface{}     `json:"params"`
	Path      string          `json:"path"`
	URI       string          `json:"uri"`
	Entities  json.RawMessage `json:"entities"`
	Data      json.RawMessage `json:"data"`
	Timestamp int64           `json:"timestamp"`
	Duration  int64           `json:"duration"`
	OrgName   string          `json:"organization"`
	AppName   string          `json:"applicationName"`
}

// RegisterUserReplay 用户注册响应结果
type RegisterUserReplay struct {
	UUID              string
	Type              string
	Created, Modified int64
	UserName          string
	Activated         bool
}

// RegisterUser 注册单个用户
func (c *Client) RegisterUser(u User) (RegisterUserReplay, error) {
	var reply commonReply
	if err := c.httpcall(c.makePath("users", nil, nil), "POST", u, &reply, true); err != nil {
		return RegisterUserReplay{}, err
	}
	var users []RegisterUserReplay
	err := json.NewDecoder(bytes.NewBuffer(reply.Entities)).Decode(&users)
	if err != nil {
		return RegisterUserReplay{}, err
	}
	return users[0], nil
}

// RegisterUsers 批量注册用户
func (c *Client) RegisterUsers(users []User) ([]RegisterUserReplay, error) {
	var reply commonReply
	uri := c.makePath("users", nil, nil)
	if err := c.httpcall(uri, "POST", users, &reply, true); err != nil {
		return nil, err
	}
	var result []RegisterUserReplay
	err := json.NewDecoder(bytes.NewBuffer(reply.Entities)).Decode(&result)
	return result, err
}

// BlockUser 加黑名单
func (c *Client) BlockUser(owner string, blockUsers ...string) error {
	uri := c.makePath("users", []string{owner, "blocks", "users"}, nil)
	var req struct {
		UserNames []string `json:"usernames"`
	}
	req.UserNames = append(req.UserNames, blockUsers...)
	var reply commonReply
	return c.httpcall(uri, "POST", req, &reply, true)
}

// RemoveBlockUser 移除黑名单
func (c *Client) RemoveBlockUser(owner, blockuser string) error {
	uri := c.makePath("users", []string{owner, "blocks", "users", blockuser}, nil)
	var reply commonReply
	return c.httpcall(uri, "DELETE", nil, &reply, true)
}
