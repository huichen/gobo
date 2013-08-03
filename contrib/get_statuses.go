package contrib

import (
	"github.com/huichen/gobo"
	"math"
	"sort"
	"time"
)

const (
	STATUSES_PER_PAGE = 100
	MAX_THREADS       = 20
)

// 并行抓取指定用户的微博
//
// 输入参数：
//	weibo		gobo.Weibo结构体指针
//	access_token	用户的访问令牌
// 	userName	微博用户名
//	userId		微博用户ID，注意仅当userName为空字符串时使用此值
//	numStatuses	需要抓取的总微博数，注意由于新浪的限制，最多只能抓取最近2000条微博，当此参数大于2000时取2000
//      timeout		超时退出，单位为毫秒，当值为0时不设超时
//
// 返回按照ID逆序排序的微博
func GetStatuses(weibo *gobo.Weibo, access_token string, userName string, userId int64, numStatuses int, timeout int) ([]*gobo.Status, error) {
	// 检查输入参数的有效性
	if userName == "" && userId == 0 {
		return nil, &gobo.ErrorString{"userName和userId不可以都是无效值"}
	}

	// 计算需要启动的进程数
	if numStatuses <= 0 {
		return nil, &gobo.ErrorString{"抓取微博数必须大于零"}
	}
	numThreads := int(math.Ceil(float64(numStatuses) / STATUSES_PER_PAGE))
	if numThreads > MAX_THREADS {
		numThreads = MAX_THREADS
	}

	// output通道中收集所有线程抓取的微博
	output := make(chan *gobo.Status, STATUSES_PER_PAGE*numThreads)

	// done通道中收集线程抓取微博的数目，并负责通知主线程是否全部子线程已经完成
	done := make(chan int, numThreads)

	// 启动子线程
	for i := 0; i < numThreads; i++ {
		// 开辟numThreads个新线程负责分页抓取微博
		go func(page int) {
			var posts gobo.Statuses
			var params gobo.Params
			if userName != "" {
				params = gobo.Params{"screen_name": userName, "count": STATUSES_PER_PAGE, "page": page}
			} else {
				params = gobo.Params{"uid": userId, "count": STATUSES_PER_PAGE, "page": page}
			}
			err := weibo.Call("statuses/user_timeline", "get", access_token, params, &posts)
			if err != nil {
				done <- 0
				return
			}
			for _, p := range posts.Statuses {
				select {
				case output <- p:
				default:
				}
			}
			done <- len(posts.Statuses)
		}(i + 1)
	}

	// 循环监听线程通道
	numCompletedThreads := 0
	numReceivedStatuses := 0
	numTotalStatuses := 0
	statuses := make([]*gobo.Status, 0, numThreads*STATUSES_PER_PAGE) // 长度为零但预留足够容量
	isTimeout := false
	t0 := time.Now()
	for {
		// 非阻塞监听output和done通道
		select {
		case status := <-output:
			statuses = append(statuses, status)
			numReceivedStatuses++
		case numThreadStatuses := <-done:
			numCompletedThreads++
			numTotalStatuses = numTotalStatuses + numThreadStatuses
		case <-time.After(time.Second): // 让子线程飞一会儿
		}

		// 超时退出
		if timeout > 0 {
			t1 := time.Now()
			if t1.Sub(t0).Nanoseconds() > int64(timeout)*1000000 {
				isTimeout = true
				break
			}
		}

		// 当所有线程完成并且从output通道收集齐全部微博时退出循环
		if numCompletedThreads == numThreads && numTotalStatuses == numReceivedStatuses {
			break
		}
	}

	if isTimeout {
		return nil, &gobo.ErrorString{"抓取超时"}
	}

	// 将所有的微博按照id顺序排序
	sort.Sort(StatusSlice(statuses))

	// 删除掉重复的微博
	sortedStatuses := make([]*gobo.Status, 0, len(statuses))
	numStatusesToReturn := 0
	for i := 0; i < len(statuses); i++ {
		// 跳过重复微博
		if i > 0 && statuses[i].Id == statuses[i-1].Id {
			continue
		}

		sortedStatuses = append(sortedStatuses, statuses[i])
		numStatusesToReturn++

		// 最多返回numStatuses条微博
		if numStatusesToReturn == numStatuses {
			break
		}
	}
	return sortedStatuses, nil
}

// 为了方便将微博排序定义下列结构体和成员函数

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
