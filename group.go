// 群组管理

package easymob

import (
	"encoding/json"
	"errors"
)

// AddGroupMemberOne 添加群组成员
func (c *Client) AddGroupMemberOne(groupid, userid string) error {
	path := c.makePath("chatgroups", []string{groupid, "users", userid}, nil)
	var reply commonReply
	err := c.httpcall(path, "POST", nil, &reply, true)
	if err != nil {
		return err
	}
	var result struct{ Result bool }
	json.Unmarshal(reply.Data, &result)
	if !result.Result {
		return errors.New("添加群组成员失败")
	}
	return nil
}
