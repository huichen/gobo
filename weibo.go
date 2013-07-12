// 定义Weibo结构体
//
// 该结构体定义了对微博API服务器的所有调用功能。
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

// Params是一个JSON key到value的映射表，包含所有API的输入参数。注意：
// 1. 所有的value都是string类型的，其它类型请转化为string
// 2. Params不应当包含access_token，因为access_token已经是Call和Upload函数的参数
// 3. Upload函数的Params参数不应当包含pic参数，上传的图片必须通过reader和imageFormat指定
type Params map[string]string

type Weibo struct {
	httpClient http.Client
}

// 调用微博API
//
// 该函数调用除了statuses/upload（见Upload函数）外的所有微博API。
//
// 输入参数
// 	method		API方法名，比如 "/statuses/user_timeline" 又如 "comments/show"
//	httpMethod	HTTP请求方式，只能是"get"或者"post"之一，否则出错
//	token		用户授权的access_token
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

// 调用Weibo API之/statuses/upload （发图片微博）
//
// 输入参数
//	token		用户授权的access_token
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
		if k != "" && v != "" {
			uriBuffer.WriteString(fmt.Sprintf("&%s=%s", k, v))
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

		for key, value := range params {
			if key != "" && value != "" {
				pb.Add(key, value)
			}
		}
		bodyBuffer = *bytes.NewBufferString(pb.Encode())
	} else {
		// 否则POST body使用multipart模式
		writer = multipart.NewWriter(&bodyBuffer)
		imagePartWriter, _ := writer.CreateFormFile("pic", "image."+imageFormat)
		io.Copy(imagePartWriter, reader)
		for key, value := range params {
			if key != "" && value != "" {
				writer.WriteField(key, value)
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
