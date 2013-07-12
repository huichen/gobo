gobo
====

新浪微博golang SDK

# 安装

```
go get github.com/huichen/gobo
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
	access_token  = flag.String("access_token", "", "用户的access token")
)

func main() {
	// 解析命令行参数
        flag.Parse()

	// 调用API
	var posts gobo.Statuses
	params := gobo.Params{"screen_name": "人民日报", "count": "10"}
	err := weibo.Call("statuses/user_timeline", "get", *access_token, params, &posts)
	
	// 处理返回结果
	if err != nil {
		fmt.Println(err)
	} else if len(posts.Statuses) > 0 {
		for _, p := range posts.Statuses {
			fmt.Println(p.Text)
		}
	}
}
```

access_token可以通过<a href="http://open.weibo.com/tools/console">API测试工具</a>或者gobo.Authenticator得到，用命令行参数-access_token传入即可。


更多例子见 <a href="https://github.com/huichen/gobo/blob/master/example/client.go">example/client.go</a>