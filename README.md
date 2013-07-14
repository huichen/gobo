gobo
====

新浪微博Go语言SDK，支持所有<a href="http://open.weibo.com/wiki/微博API">微博API功能</a>

# 安装/更新

```
go get -u github.com/huichen/gobo
```

# 使用

抓取<a href="http://weibo.com/rmrb">@人民日报</a>的最近10条微博:

```go
package main

import (
	"flag"
	"fmt"
	"github.com/huichen/gobo"
)

var (
	weibo = gobo.Weibo{}
	access_token = flag.String("access_token", "", "用户的访问令牌")
)

func main() {
	// 解析命令行参数
	flag.Parse()

	// 调用API
	var statuses gobo.Statuses
	params := gobo.Params{"screen_name": "人民日报", "count": 10}
	err := weibo.Call("statuses/user_timeline", "get", *access_token, params, &statuses)
	
	// 处理返回结果
	if err != nil {
		fmt.Println(err)
		return
	}
	for _, status := range statuses.Statuses {
		fmt.Println(status.Text)
	}
}
```

用命令行参数-access_token传入访问令牌，令牌可以通过<a href="http://open.weibo.com/tools/console">API测试工具</a>或者<a href="https://github.com/huichen/gobo/blob/master/examples/auth.go">gobo.Authenticator</a>得到。

更多API调用的例子见 <a href="https://github.com/huichen/gobo/blob/master/examples/weibo.go">examples/weibo.go</a>。
