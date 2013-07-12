gobo
====

微博golang SDK

# 安装

```
go get github.com/huichen/gobo
```

# 使用

抓取<a href="http://weibo.com/rmrb">@人民日报</a>的最近10条微博（请将代码中的 *redirect_url*, *client_id*, *client_secret* 和 *access_token* 正确地替换掉）:

```go
package main

import (
	"github.com/huichen/gobo"
)

func main() {
	// 初始化
	var wb gobo.Weibo
	wb.Init(<redirect_uri>, <client_id>, <client_secret>)
	
	// 调用API
	var posts gobo.Statuses
	params := gobo.Params{"screen_name": "人民日报", "count": "10"}
	err := wb.Call("statuses/user_timeline", "get", <access_token>, params, &posts)
	
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

更多例子见 <a href="https://github.com/huichen/gobo/blob/master/example/client.go">example/client.go</a>
