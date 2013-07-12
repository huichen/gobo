package main

import (
	"flag"
	"fmt"
	"github.com/huichen/gobo"
	"math/rand"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

const (
	NUM_THREADS    = 10
	POSTS_PER_PAGE = 100
)

var (
	access_token = flag.String("access_token", "", "用户的access token")
	image        = flag.String("image", "", "上传图片的位置")
	random       = rand.New(rand.NewSource(time.Now().UnixNano()))
	weibo        = gobo.Weibo{}
)

func showUser() {
	fmt.Println("==== 测试 users/show ====")
	var user gobo.User
	params := gobo.Params{"screen_name": "人民日报"}
	err := weibo.Call("users/show", "get", *access_token, params, &user)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Printf("%#v\n", user)
	}
}

func getFriendsStatuses() {
	fmt.Println("==== 测试 statuses/friends_timeline ====")
	var posts gobo.Statuses
	params := gobo.Params{"count": "10"}
	err := weibo.Call("statuses/friends_timeline", "get", *access_token, params, &posts)
	if err != nil {
		fmt.Println(err)
	} else {
		for _, p := range posts.Statuses {
			fmt.Println(p.Text)
		}
	}
}

func getUserStatus() {
	fmt.Println("==== 测试 statuses/user_timeline ====")
	var posts gobo.Statuses
	params := gobo.Params{"screen_name": "人民日报", "count": "1"}
	err := weibo.Call("statuses/user_timeline", "get", *access_token, params, &posts)
	if err != nil {
		fmt.Println(err)
	} else if len(posts.Statuses) > 0 {
		fmt.Printf("%#v\n", posts.Statuses[0])
	}
}

func getUserStatuses() {
	fmt.Println("==== 测试并行调用 statuses/user_timeline ====")
	input := make(chan int, NUM_THREADS)
	output := make(chan int, NUM_THREADS)
	for i := 1; i <= NUM_THREADS; i++ {
		go func(page int) {
			var posts gobo.Statuses
			params := gobo.Params{"screen_name": "人民日报", "count": strconv.Itoa(POSTS_PER_PAGE), "page": strconv.Itoa(page)}
			weibo.Call("statuses/user_timeline", "get", *access_token, params, &posts)
			for _, p := range posts.Statuses {
				fmt.Println(page, ":", p.Text)
			}
			output <- 1
		}(i)
		input <- i
	}
	for i := 0; i < NUM_THREADS; i++ {
		<-output
	}
}

func updateStatus() {
	fmt.Println("==== 测试 statuses/update ====")
	var post gobo.Status
	params := gobo.Params{"status": "测试" + strconv.Itoa(rand.Int())}
	err := weibo.Call("statuses/update", "post", *access_token, params, &post)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Printf("%#v\n", post)
	}
}

func uploadStatus() {
	fmt.Println("==== 测试 statuses/upload ====")
	var post gobo.Status
	params := gobo.Params{"status": "测试" + strconv.Itoa(rand.Int())}
	img, err := os.Open(*image)
	if err != nil {
		fmt.Println(err)
	}
	err = weibo.Upload(*access_token, params, img, filepath.Ext(*image), &post)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Printf("%#v\n", post)
	}
}

func main() {
	flag.Parse()
	showUser()
	getFriendsStatuses()
	getUserStatus()
	getUserStatuses()
	//updateStatus()
	//uploadStatus()
}
