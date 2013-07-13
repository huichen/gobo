// 样例程序：利用goroutines并行抓取微博
package main

import (
	"flag"
	"fmt"
	"github.com/huichen/gobo"
	"sort"
	"strconv"
	"time"
)

const (
	NUM_THREADS       = 20
	STATUSES_PER_PAGE = 100
)

var (
	access_token = flag.String("access_token", "", "用户的access token")
	weibo        = gobo.Weibo{}
)

// 为了方便将微博排序定义下列结构体和函数

type StatusSlice []*gobo.Status

func (ss StatusSlice) Len() int {
	return len(ss)
}
func (ss StatusSlice) Swap(i, j int) {
	ss[i], ss[j] = ss[j], ss[i]
}
func (ss StatusSlice) Less(i, j int) bool {
	return ss[i].Id > ss[j].Id
}

func getUserStatusesWithGoroutines() {
	fmt.Println("==== 测试并行调用 statuses/user_timeline ====")

	// 记录初始时间
	t0 := time.Now()

	// 为每个线程建立通道，从子线程中抓取的微博依次压入相应通道中
	output := [NUM_THREADS]chan *gobo.Status{}
	for i := 0; i < NUM_THREADS; i++ {
		output[i] = make(chan *gobo.Status, STATUSES_PER_PAGE)
	}

	// 启动线程
	for i := 0; i < NUM_THREADS; i++ {
		// 开辟NUM_THREADS个新线程负责分页抓取微博
		go func(page int, outputChannel chan *gobo.Status) {
			var posts gobo.Statuses
			params := gobo.Params{"screen_name": "人民日报", "count": strconv.Itoa(STATUSES_PER_PAGE), "page": strconv.Itoa(page)}
			err := weibo.Call("statuses/user_timeline", "get", *access_token, params, &posts)
			if err != nil {
				fmt.Println(err)
				close(outputChannel)
				return
			}
			fmt.Printf("线程%d抓取的微博数 %d\n", page, len(posts.Statuses))
			for _, p := range posts.Statuses {
				select {
				case outputChannel <- p:
				default:
				}
			}
			close(outputChannel)
		}(i+1, output[i])
	}

	// 循环监听线程通道
	numCompletedThreads := 0
	statuses := make([]*gobo.Status, 0, NUM_THREADS*STATUSES_PER_PAGE) // 长度为零但预留足够容量
	completedChannels := map[int]bool{}
	for numCompletedThreads < NUM_THREADS { // 仅当所有通道关闭时退出循环
		for i, ch := range output {
			if !completedChannels[i] {
				status, more := <-ch
				if more {
					statuses = append(statuses, status)
				} else if !completedChannels[i] {
					completedChannels[i] = true
					numCompletedThreads++
				}
			}
		}
	}

	// 将所有的微博按照id顺序排序打印
	sort.Sort(StatusSlice(statuses))

	// 删除掉重复的微博
	newStatuses := make([]*gobo.Status, 0, len(statuses))
	for i := 0; i < len(statuses); i++ {
		if i > 0 && statuses[i].Id == statuses[i-1].Id {
			continue
		}
		newStatuses = append(newStatuses, statuses[i])
	}
	fmt.Printf("\n抓取的总微博数 %d\n", len(newStatuses))

	// 记录终止时间
	t1 := time.Now()
	fmt.Printf("并行抓取花费时间 %v\n", t1.Sub(t0))
}

func main() {
	flag.Parse()
	getUserStatusesWithGoroutines()
}
