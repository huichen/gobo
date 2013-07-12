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

// 微博API域名
const (
	apiDomain string = "https://api.weibo.com"
)

// Params是一个JSON key到value的映射表，包含所有API的输入参数。注意：
// 1. 所有的value都是string类型的，其它类型请转化为string
// 2. Params不应当包含access_token，因为access_token已经是Call和Upload函数的参数
// 3. Upload函数的Params参数不应当包含pic参数，上传的图片必须通过reader和imageFormat指定
type Params map[string]string

type errorString struct {
	s string
}

type Weibo struct {
	redirectUri  string
	clientId     string
	clientSecret string
	initialized  bool
	httpClient   *http.Client
}

func (e *errorString) Error() string {
	return e.s
}

// 初始化结构体
//
// 在调用其它函数之前必须首先初始化。
func (wb *Weibo) Init(redirectUri string, clientId string, clientSecret string) {
	wb.redirectUri = redirectUri
	wb.clientId = clientId
	wb.clientSecret = clientSecret
	wb.httpClient = new(http.Client)
	wb.initialized = true
}

// 得到认证URI
func (wb *Weibo) GetAuthURI() (string, error) {
	// 检查结构体是否初始化
	if !wb.initialized {
		return "", &errorString{"Weibo结构体尚未初始化"}
	}

	return fmt.Sprintf("%s/oauth2/authorize?redirect_uri=%s&response_type=code&client_id=%s", apiDomain, wb.redirectUri, wb.clientId), nil
}

// 给定认证code得到access token
func (wb *Weibo) GetAccessToken(code string) (AccessToken, error) {
	// 检查结构体是否初始化
	token := AccessToken{}
	if !wb.initialized {
		return token, &errorString{"Weibo结构体尚未初始化"}
	}

	// 生成请求URI
	requestUri := fmt.Sprintf("%s/oauth2/access_token", apiDomain)
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

// 调用微博API
// 
// 该函数调用除了statuses/upload（见Upload函数）外的所有微博API。
//
// 输入参数
// 	method		API方法名，比如 "/statuses/user_timeline" 又如 "comments/show"
//	httpMethod	HTTP请求方式，只能是"get"或者"post"之一，否则出错
//	token		用户授权的access_token
//	params		JSON输入参数，见Params结构体的注释
//	v		API服务器的JSON输出将被还原成该结构体	
//
// 当出现异常时输出非nil错误
func (wb *Weibo) Call(method string, httpMethod string, token string, params Params, v interface{}) error {
	// 检查结构体是否初始化
	if !wb.initialized {
		return &errorString{"Weibo结构体尚未初始化"}
	}

	if httpMethod == "get" {
		return wb.get_http_request(fmt.Sprintf("%s/2/%s.json", apiDomain, method), token, params, v)
	} else if httpMethod == "post" {
		return wb.post_http_request(fmt.Sprintf("%s/2/%s.json", apiDomain, method), token, params, nil, "", v)
	}
	return &errorString{"HTTP方法只能是\"get\"或者\"post\""}
}

// 调用Weibo API之/statuses/upload （发图片微博）
//
// 输入参数
//	token		用户授权的access_token
//	params		JSON输入参数，见Params结构体的注释
//	reader		包含图片的二进制流
//	imageFormat	图片的格式，比如 "jpg" 又如 "png"
//	v		API服务器的JSON输出将被还原成该结构体	
//
// 当出现异常时输出非nil错误
func (wb *Weibo) Upload(token string, params Params, reader io.Reader, imageFormat string, v interface{}) error {
	// 检查结构体是否初始化
	if !wb.initialized {
		return &errorString{"Weibo结构体尚未初始化"}
	}

	return wb.post_http_request(fmt.Sprintf("%s/2/statuses/upload.json", apiDomain), token, params, reader, imageFormat, v)
}

// 向微博API服务器发送GET请求
func (wb *Weibo) get_http_request(uri string, token string, params Params, v interface{}) error {
	// 生成请求URI
	var buffer bytes.Buffer
	buffer.WriteString(fmt.Sprintf("%s?access_token=%s", uri, token))
	for k, v := range params {
		if k != "" && v != "" {
			buffer.WriteString(fmt.Sprintf("&%s=%s", k, v))
		}
	}
	requestUri := buffer.String()

	// 发送GET请求
	resp, err := wb.httpClient.Get(requestUri)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// 解析API服务器返回内容
	bytes, _ := ioutil.ReadAll(resp.Body)
	if resp.StatusCode == 200 {
		err := json.Unmarshal(bytes, &v)
		if err != nil {
			return err
		}
		return nil
	} else {
		var e WeiboError
		err := json.Unmarshal(bytes, &e)
		if err != nil {
			return err
		}
		return e
	}
	return nil
}

// 向微博API服务器发送POST请求
//
// 输入参数的含义请见Upload函数注释。当reader == nil时使用form模式，否则使用multipart。
func (wb *Weibo) post_http_request(uri string, token string, params Params, reader io.Reader, imageFormat string, v interface{}) error {
	// 生成POST请求URI
	requestUri := fmt.Sprintf("%s?access_token=%s", uri, token)

	// 生成POST内容
	var body_buffer bytes.Buffer
	var w *multipart.Writer
	if reader == nil {
		pb := url.Values{}
		pb.Add("access_token", token)

		for key, value := range params {
			if key != "" && value != "" {
				pb.Add(key, value)
			}
		}
		body_buffer = *bytes.NewBufferString(pb.Encode())
	} else {
		w = multipart.NewWriter(&body_buffer)
		wr, _ := w.CreateFormFile("pic", "image."+imageFormat)
		io.Copy(wr, reader)
		for key, value := range params {
			if key != "" && value != "" {
				w.WriteField(key, value)
			}
		}
		w.Close()
	}

	// 发送POST请求
	req, err := http.NewRequest("POST", requestUri, &body_buffer)
	if err != nil {
		return err
	}
	if reader == nil {
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	} else {
		req.Header.Set("Content-Type", w.FormDataContentType())
	}
	resp, err := wb.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// 解析API服务器返回内容
	bytes, _ := ioutil.ReadAll(resp.Body)
	if resp.StatusCode == 200 {
		err := json.Unmarshal(bytes, &v)
		if err != nil {
			return err
		}
		return nil
	} else {
		var e WeiboError
		err := json.Unmarshal(bytes, &e)
		if err != nil {
			return err
		}
		return e
	}
	return nil
}
