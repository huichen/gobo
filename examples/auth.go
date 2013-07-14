// 例子程序：微博应用授权
// 展示功能包括得到授权URI，通过授权码得到访问令牌，获得令牌对应的信息和解除访问令牌授权等功能。
package main

import (
	"bufio"
	"flag"
	"fmt"
	"github.com/huichen/gobo"
	"os"
	"strings"
)

var (
	redirect_uri  = flag.String("redirect_uri", "", "应用的重定向地址")
	client_id     = flag.String("client_id", "", "应用的client id")
	client_secret = flag.String("client_secret", "", "应用的client secret")
	auth          = gobo.Authenticator{}
)

func main() {
	flag.Parse()

	// 初始化
	err := auth.Init(*redirect_uri, *client_id, *client_secret)
	if err != nil {
		fmt.Println(err)
		return
	}

	// 得到重定向地址
	uri, err := auth.Authorize()
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("请在浏览器中打开下面地址\n%s\n", uri)

	// 从终端读取用户输入的认证码
	fmt.Print("请输入浏览器返回的授权码：")
	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	code := strings.TrimSuffix(string([]byte(input)), "\n")

	// 从授权码得到token
	token, err := auth.AccessToken(code)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("访问令牌 = %#v\n", token)

	// 从token得到相关信息
	info, err := auth.GetTokenInfo(token.Access_Token)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("访问令牌信息 = %#v\n", info)

	// 解除token授权
	revokeErr := auth.Revokeoauth2(token.Access_Token)
	if revokeErr != nil {
		fmt.Println(revokeErr)
		return
	}
	fmt.Println("解除授权成功")
}
