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

	// output通道中收集所有线程抓取的微博
	output := make(chan *gobo.Status, STATUSES_PER_PAGE*NUM_THREADS)

	// done通道中收集线程抓取微博的数目，并负责通知主线程是否全部子线程已经完成
	done := make(chan int, NUM_THREADS)

	// 启动子线程
	for i := 0; i < NUM_THREADS; i++ {
		// 开辟NUM_THREADS个新线程负责分页抓取微博
		go func(page int) {
			var posts gobo.Statuses
			params := gobo.Params{"screen_name": "人民日报", "count": strconv.Itoa(STATUSES_PER_PAGE), "page": strconv.Itoa(page)}
			err := weibo.Call("statuses/user_timeline", "get", *access_token, params, &posts)
			if err != nil {
				fmt.Println(err)
				done <- 0
				return
			}
			for _, p := range posts.Statuses {
				select {
				case output <- p:
				default:
				}
			}
			fmt.Printf("线程%d抓取的微博数 %d\n", page, len(posts.Statuses))
			done <- len(posts.Statuses)
		}(i + 1)
	}

	// 循环监听线程通道
	numCompletedThreads := 0
	numReceivedStatuses := 0
	numTotalStatuses := 0
	statuses := make([]*gobo.Status, 0, NUM_THREADS*STATUSES_PER_PAGE) // 长度为零但预留足够容量
	for {
		// 非阻塞监听output和done通道
		select {
		case status := <-output:
			statuses = append(statuses, status)
			numReceivedStatuses++
		case numStatuses := <-done:
			numCompletedThreads++
			numTotalStatuses = numTotalStatuses + numStatuses
		case <-time.After(time.Second): // 让子线程飞一会儿
		}

		// 仅当所有线程完成并且从output通道收集齐全部微博时退出循环
		if numCompletedThreads == NUM_THREADS && numTotalStatuses == numReceivedStatuses {
			break
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
