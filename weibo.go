package gobo

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/url"
)

// Params类型用来表达微博API的JSON输入参数。注意：
// 	1. Params不应当包含访问令牌(access_token)，因为它已经是Call和Upload函数的参数
// 	2. 在Upload函数中，Params参数不应当包含pic参数，上传的图片内容和类型应当通过reader和imageFormat指定
type Params map[string]interface{}

// Weibo结构体定义了微博API调用功能
type Weibo struct {
	httpClient http.Client
}

// 调用微博API
//
// 该函数可用来调用除了statuses/upload（见Upload函数）和微博授权（见Authenticator结构体）外的所有微博API。
//
// 输入参数
// 	method		API方法名，比如 "/statuses/user_timeline" 又如 "comments/show"
//	httpMethod	HTTP请求方式，只能是"get"或者"post"之一，否则出错
//	token		用户授权的访问令牌
//	params		JSON输入参数，见Params结构体的注释
//	response	API服务器的JSON输出将被还原成该结构体
//
// 当出现异常时输出非nil错误
func (weibo *Weibo) Call(method string, httpMethod string, token string, params Params, response interface{}) error {
	apiUri := fmt.Sprintf("%s/%s/%s%s", ApiDomain, ApiVersion, method, ApiNamePostfix)
	if httpMethod == "get" {
		return weibo.sendGetHttpRequest(apiUri, token, params, response)
	} else if httpMethod == "post" {
		return weibo.sendPostHttpRequest(apiUri, token, params, nil, "", response)
	}
	return &ErrorString{"HTTP方法只能是\"get\"或者\"post\""}
}

// 调用/statuses/upload发带图片微博
//
// 输入参数
//	token		用户授权的访问令牌
//	params		JSON输入参数，见Params结构体的注释
//	reader		包含图片的二进制流
//	imageFormat	图片的格式，比如 "jpg" 又如 "png"
//	response	API服务器的JSON输出将被还原成该结构体
//
// 当出现异常时输出非nil错误
func (weibo *Weibo) Upload(token string, params Params, reader io.Reader, imageFormat string, response interface{}) error {
	apiUri := fmt.Sprintf("%s/%s/%s%s", ApiDomain, ApiVersion, UploadAPIName, ApiNamePostfix)
	return weibo.sendPostHttpRequest(apiUri, token, params, reader, imageFormat, response)
}

// 向微博API服务器发送GET请求
func (weibo *Weibo) sendGetHttpRequest(uri string, token string, params Params, response interface{}) error {
	// 生成请求URI
	var uriBuffer bytes.Buffer
	uriBuffer.WriteString(fmt.Sprintf("%s?access_token=%s", uri, token))
	for k, v := range params {
		value := fmt.Sprint(v)
		if k != "" && value != "" {
			uriBuffer.WriteString(fmt.Sprintf("&%s=%s", k, value))
		}
	}
	requestUri := uriBuffer.String()

	// 发送GET请求
	resp, err := weibo.httpClient.Get(requestUri)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// 解析API服务器返回内容
	bytes, _ := ioutil.ReadAll(resp.Body)
	if resp.StatusCode == 200 {
		err := json.Unmarshal(bytes, &response)
		if err != nil {
			return err
		}
		return nil
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

// 向微博API服务器发送POST请求
//
// 输入参数的含义请见Upload函数注释。当reader == nil时使用query string模式，否则使用multipart。
func (weibo *Weibo) sendPostHttpRequest(uri string, token string, params Params, reader io.Reader, imageFormat string, response interface{}) error {
	// 生成POST请求URI
	requestUri := fmt.Sprintf("%s?access_token=%s", uri, token)

	// 生成POST内容
	var bodyBuffer bytes.Buffer
	var writer *multipart.Writer
	if reader == nil {
		// reader为nil时无文件上传，因此POST body为简单的query string模式
		pb := url.Values{}
		pb.Add("access_token", token)

		for k, v := range params {
			value := fmt.Sprint(v)
			if k != "" && value != "" {
				pb.Add(k, value)
			}
		}
		bodyBuffer = *bytes.NewBufferString(pb.Encode())
	} else {
		// 否则POST body使用multipart模式
		writer = multipart.NewWriter(&bodyBuffer)
		imagePartWriter, _ := writer.CreateFormFile("pic", "image."+imageFormat)
		io.Copy(imagePartWriter, reader)
		for k, v := range params {
			value := fmt.Sprint(v)
			if k != "" && value != "" {
				writer.WriteField(k, value)
			}
		}
		writer.Close()
	}

	// 生成POST请求
	req, err := http.NewRequest("POST", requestUri, &bodyBuffer)
	if err != nil {
		return err
	}
	if reader == nil {
		// reader为nil时使用一般的内容类型
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	} else {
		// 否则使用带boundary的multipart类型
		req.Header.Set("Content-Type", writer.FormDataContentType())
	}

	// 发送请求
	resp, err := weibo.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// 解析API服务器返回内容
	bytes, _ := ioutil.ReadAll(resp.Body)
	if resp.StatusCode == 200 {
		err := json.Unmarshal(bytes, &response)
		if err != nil {
			return err
		}
		return nil
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
