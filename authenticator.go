// 定义Authenticator结构体
//
// 该结构体实现了微博API认证功能。
package gobo

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

type Authenticator struct {
	redirectUri  string
	clientId     string
	clientSecret string
	initialized  bool
	httpClient   *http.Client
}

// 初始化结构体
//
// 在调用其它函数之前必须首先初始化。
func (wb *Authenticator) Init(redirectUri string, clientId string, clientSecret string) error {
	// 检查结构体是否已经初始化
	if !wb.initialized {
		return &ErrorString{"Weibo结构体已经初始化"}
	}

	wb.redirectUri = redirectUri
	wb.clientId = clientId
	wb.clientSecret = clientSecret
	wb.httpClient = new(http.Client)
	wb.initialized = true
	return nil
}

// 得到认证URI
func (wb *Authenticator) GetAuthURI() (string, error) {
	// 检查结构体是否初始化
	if !wb.initialized {
		return "", &ErrorString{"Weibo结构体尚未初始化"}
	}

	return fmt.Sprintf("%s/oauth2/authorize?redirect_uri=%s&response_type=code&client_id=%s", ApiDomain, wb.redirectUri, wb.clientId), nil
}

// 给定认证code得到access token
func (wb *Authenticator) GetAccessToken(code string) (AccessToken, error) {
	// 检查结构体是否初始化
	token := AccessToken{}
	if !wb.initialized {
		return token, &ErrorString{"Weibo结构体尚未初始化"}
	}

	// 生成请求URI
	requestUri := fmt.Sprintf("%s/oauth2/access_token", ApiDomain)
	v := url.Values{}
	v.Add("client_id", wb.clientId)
	v.Add("client_secret", wb.clientSecret)
	v.Add("redirect_uri", wb.redirectUri)
	v.Add("grant_type", "authorization_code")
	v.Add("code", code)

	// 发送POST Form请求
	resp, err := wb.httpClient.PostForm(requestUri, v)
	if err != nil {
		return token, err
	}
	defer resp.Body.Close()

	// 解析返回内容
	json.NewDecoder(resp.Body).Decode(&token)
	return token, nil
}
