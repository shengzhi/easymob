// Package easymob 环信服务端集成SDK
package easymob

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"
)

const apidomain = "http://a1.easemob.com" //"https://a1.easemob.com"

// Client 环信客户端
type Client struct {
	appkey, clientid, clientSecret string
	orgName, appName               string
	token                          struct {
		value   string
		expired time.Time
	}
	httpClient *http.Client
}

// NewClient 创建客户端
func NewClient(appkey, clientid, clientsecret string) *Client {
	client := &Client{
		appkey:       appkey,
		clientid:     clientid,
		clientSecret: clientsecret,
	}
	client.parseAppKey()
	client.httpClient = &http.Client{
		Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}},
		Timeout:   time.Second * 30,
	}
	return client
}

func (c *Client) parseAppKey() {
	key := strings.SplitN(c.appkey, "#", 2)
	c.orgName, c.appName = key[0], key[1]
}

type token struct {
	Value       string `json:"access_token"`
	ExpiresIn   int64  `json:"expires_in"`
	Application string `json:"application"`
}
type tokenReq struct {
	GrantType    string `json:"grant_type"`
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
}

func (c *Client) tokenValue() string {
	if c.token.value == "" || time.Now().After(c.token.expired) {
		v, err := c.getToken()
		if err != nil {
			log.Printf("Get Access Token error:%v\r\n", err)
			return ""
		}
		c.token.value = v.Value
		c.token.expired = time.Now().Add(time.Second * time.Duration(v.ExpiresIn))
	}
	return c.token.value
}

func (c *Client) getToken() (token, error) {
	var res token
	req := tokenReq{GrantType: "client_credentials", ClientID: c.clientid, ClientSecret: c.clientSecret}
	err := c.httpcall(c.makePath("token", nil, nil), "POST", req, &res, false)
	return res, err
}

func (c *Client) makePath(resource string, segs []string, args url.Values) string {
	path := fmt.Sprintf("%s/%s/%s/%s", apidomain, c.orgName, c.appName, resource)
	if len(segs) > 0 {
		path = fmt.Sprintf("%s/%s", path, strings.Join(segs, "/"))
	}
	if len(args) > 0 {
		path = fmt.Sprintf("%s?%s", path, args.Encode())
	}
	return path
}

func (c *Client) httpcall(uri, method string, request, response interface{}, needToken bool) error {
	var body bytes.Buffer
	if request != nil {
		if err := json.NewEncoder(&body).Encode(request); err != nil {
			return err
		}
	}
	req, err := http.NewRequest(method, uri, &body)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	if needToken {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.tokenValue()))
	}

	reqdata, _ := httputil.DumpRequest(req, true)
	fmt.Println(string(reqdata))
	res, err := c.httpClient.Do(req)
	if res != nil {
		defer res.Body.Close()
	}
	if err != nil {
		return err
	}
	if res.StatusCode != 200 {
		data, _ := ioutil.ReadAll(res.Body)
		return fmt.Errorf("Code:%d,Body:%s", res.StatusCode, string(data))
	}
	return json.NewDecoder(res.Body).Decode(response)
}

type mediaEntity struct {
	UUID   string `json:"uuid"`
	Type   string `json:"type"`
	Secret string `json:"share-secret"`
	URL    string
}

func (c *Client) uploadImgAndVoice(file io.Reader) (mediaEntity, error) {
	var reply commonReply
	if err := c.postFile(file, &reply); err != nil {
		return mediaEntity{}, err
	}
	var entities []mediaEntity
	if err := json.Unmarshal(reply.Entities, &entities); err != nil {
		return mediaEntity{}, err
	}
	if len(entities) <= 0 {
		return mediaEntity{}, fmt.Errorf("上传失败")
	}
	result := entities[0]
	result.URL = fmt.Sprintf("%s/%s", reply.URI, result.UUID)
	return result, nil
}

func (c *Client) postFile(file io.Reader, response interface{}) error {
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	part, err := w.CreateFormFile("file", "")
	if err != nil {
		return err
	}

	if _, err = io.Copy(part, file); err != nil {
		return err
	}

	w.Close()
	uri := c.makePath("chatfiles", nil, nil)
	req, err := http.NewRequest("POST", uri, &buf)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", w.FormDataContentType())
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.tokenValue()))
	req.Header.Set("restrict-access", "true")

	res, err := c.httpClient.Do(req)
	if res != nil {
		defer res.Body.Close()
	}
	if err != nil {
		return err
	}
	if res.StatusCode != 200 {
		data, _ := ioutil.ReadAll(res.Body)
		return fmt.Errorf("upload file error:%s", string(data))
	}
	return json.NewDecoder(res.Body).Decode(response)
}
