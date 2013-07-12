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
func (auth *Authenticator) Init(redirectUri string, clientId string, clientSecret string) error {
	// 检查结构体是否已经初始化
	if !auth.initialized {
		return &ErrorString{"Weibo结构体已经初始化"}
	}

	auth.redirectUri = redirectUri
	auth.clientId = clientId
	auth.clientSecret = clientSecret
	auth.httpClient = new(http.Client)
	auth.initialized = true
	return nil
}

// 得到认证URI
func (auth *Authenticator) GetAuthURI() (string, error) {
	// 检查结构体是否初始化
	if !auth.initialized {
		return "", &ErrorString{"Weibo结构体尚未初始化"}
	}

	return fmt.Sprintf("%s/oauth2/authorize?redirect_uri=%s&response_type=code&client_id=%s", ApiDomain, auth.redirectUri, auth.clientId), nil
}

// 给定认证code得到access token
func (auth *Authenticator) GetAccessToken(code string) (AccessToken, error) {
	// 检查结构体是否初始化
	token := AccessToken{}
	if !auth.initialized {
		return token, &ErrorString{"Weibo结构体尚未初始化"}
	}

	// 生成请求URI
	requestUri := fmt.Sprintf("%s/oauth2/access_token", ApiDomain)
	queries := url.Values{}
	queries.Add("client_id", auth.clientId)
	queries.Add("client_secret", auth.clientSecret)
	queries.Add("redirect_uri", auth.redirectUri)
	queries.Add("grant_type", "authorization_code")
	queries.Add("code", code)

	// 发送POST Form请求
	resp, err := auth.httpClient.PostForm(requestUri, queries)
	if err != nil {
		return token, err
	}
	defer resp.Body.Close()

	// 解析返回内容
	json.NewDecoder(resp.Body).Decode(&token)
	return token, nil
}
