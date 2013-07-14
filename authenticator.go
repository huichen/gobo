package gobo

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
)

// Authenticator结构体实现了微博应用授权功能
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
	if auth.initialized {
		return &ErrorString{"Authenticator结构体已经初始化"}
	}

	auth.redirectUri = redirectUri
	auth.clientId = clientId
	auth.clientSecret = clientSecret
	auth.httpClient = new(http.Client)
	auth.initialized = true
	return nil
}

// 得到授权URI
func (auth *Authenticator) Authorize() (string, error) {
	// 检查结构体是否初始化
	if !auth.initialized {
		return "", &ErrorString{"Authenticator结构体尚未初始化"}
	}

	return fmt.Sprintf("%s/oauth2/authorize?redirect_uri=%s&response_type=code&client_id=%s", ApiDomain, auth.redirectUri, auth.clientId), nil
}

// 从授权码得到访问令牌
func (auth *Authenticator) AccessToken(code string) (AccessToken, error) {
	// 检查结构体是否初始化
	token := AccessToken{}
	if !auth.initialized {
		return token, &ErrorString{"Authenticator结构体尚未初始化"}
	}

	// 生成请求URI
	queries := url.Values{}
	queries.Add("client_id", auth.clientId)
	queries.Add("client_secret", auth.clientSecret)
	queries.Add("redirect_uri", auth.redirectUri)
	queries.Add("grant_type", "authorization_code")
	queries.Add("code", code)

	// 发送请求
	err := auth.sendPostHttpRequest("oauth2/access_token", queries, &token)
	return token, err
}

// 得到访问令牌对应的信息
func (auth *Authenticator) GetTokenInfo(token string) (AccessTokenInfo, error) {
	// 检查结构体是否初始化
	info := AccessTokenInfo{}
	if !auth.initialized {
		return info, &ErrorString{"Authenticator结构体尚未初始化"}
	}

	// 生成请求URI
	queries := url.Values{}
	queries.Add("access_token", token)

	// 发送请求
	err := auth.sendPostHttpRequest("oauth2/get_token_info", queries, &info)
	return info, err
}

// 解除访问令牌的授权
func (auth *Authenticator) Revokeoauth2(token string) error {
	// 检查结构体是否初始化
	if !auth.initialized {
		return &ErrorString{"Authenticator结构体尚未初始化"}
	}

	// 生成请求URI
	queries := url.Values{}
	queries.Add("access_token", token)

	// 发送请求
	type Result struct {
		Result string
	}
	var result Result
	err := auth.sendPostHttpRequest("oauth2/revokeoauth2", queries, &result)
	return err
}

func (auth *Authenticator) sendPostHttpRequest(apiName string, queries url.Values, response interface{}) error {
	// 生成请求URI
	requestUri := fmt.Sprintf("%s/%s", ApiDomain, apiName)

	// 发送POST Form请求
	resp, err := auth.httpClient.PostForm(requestUri, queries)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// 解析返回内容
	bytes, _ := ioutil.ReadAll(resp.Body)
	if resp.StatusCode == 200 {
		err := json.Unmarshal(bytes, &response)
		if err != nil {
			return err
		}
	} else {
		var weiboErr WeiboError
		err := json.Unmarshal(bytes, &weiboErr)
		if err != nil {
			return err
		}
		return weiboErr
	}
	return nil
}
