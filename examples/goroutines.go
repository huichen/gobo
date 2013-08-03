// 例子程序：利用goroutines并行抓取微博
package main

import (
	"flag"
	"fmt"
	"github.com/huichen/gobo"
	"github.com/huichen/gobo/contrib"
	"time"
)

var (
	access_token = flag.String("access_token", "", "用户的访问令牌")
	weibo        = gobo.Weibo{}
	timeout      = flag.Int("timeout", 0, "超时，单位毫秒")
)

func main() {
	flag.Parse()
	fmt.Println("==== 测试并行调用 statuses/user_timeline ====")

	// 记录初始时间
	t0 := time.Now()

	// 抓微博
	statuses, err := contrib.GetStatuses(&weibo, *access_token,
		"人民日报",   // 微博用户名
		0,        // 微博用户ID，仅当用户名为空字符串时使用
		211,      // 抓取微博数
		*timeout) // 不设超时
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("抓取的总微博数 %d\n", len(statuses))

	// 记录终止时间
	t1 := time.Now()
	fmt.Printf("并行抓取花费时间 %v\n", t1.Sub(t0))

	// 打印最后五条微博内容
	for i, status := range statuses {
		if i == 5 {
			break
		}
		fmt.Println(status.Text)
	}
}
