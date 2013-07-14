package gobo

import (
	"fmt"
)

// 微博API返回对象数据结构
//
// 结构体根据下面的文档定义
// http://open.weibo.com/wiki/常见返回对象数据结构
//
// JSON字段名和golang结构体字段名有这样的一一对应关系
//
//	JSON字段:			golang结构体字段:
//	name_of_a_field			Name_Of_A_Field

type Status struct {
	Created_At              string
	Id                      int64
	Mid                     string
	Text                    string
	Idstr                   string
	Source                  string
	Favorited               bool
	Trucated                bool
	In_Reply_To_Status_Id   string
	In_Reply_To_User_Id     string
	In_Reply_To_Screen_Name string
	Thumbnail_Pic           string
	Bmiddle_Pic             string
	Original_Pic            string
	Geo                     *Geo
	User                    *User
	Retweeted_Status        *Status
	Reposts_Count           int
	Comments_Count          int
	Attitudes_Count         int
	Mlevel                  int
	Visible                 *Visible
	Pic_Urls                []*Pic_Url
}

type Comment struct {
	Created_At    string
	Id            int64
	Text          string
	Source        string
	User          *User
	Mid           string
	Idstr         string
	Status        string
	Reply_Comment *Comment
}

type User struct {
	Id                 int64
	Idstr              string
	Screen_Name        string
	Name               string
	Province           string
	City               string
	Location           string
	Description        string
	Url                string
	Profile_Image_Url  string
	Profile_Url        string
	Domain             string
	Weihao             string
	Gender             string
	Followers_Count    int
	Friends_Count      int
	Statuses_Count     int
	Favourites_Count   int
	Created_At         string
	Following          bool
	Allow_All_Act_Msg  bool
	Geo_Enabled        bool
	Verified           bool
	Verified_Type      int
	Remark             string
	Status             *Status
	Allow_All_Comment  bool
	Avatar_Large       string
	Verified_Reason    string
	Follow_Me          bool
	Online_Status      int
	Bi_Followers_Count int
	Lang               string
}

type Privacy struct {
	Comment  int
	Geo      int
	Message  int
	Realname int
	Badge    int
	Mobile   int
	Webim    int
}

type Remind struct {
	Status         int
	Follower       int
	Cmt            int
	Dm             int
	Mention_Status int
	Mention_Cmt    int
	Group          int
	Private_Group  int
	Notice         int
	Invite         int
	Badge          int
	Photo          int
}

type Url_Short struct {
	Url_Short string
	Url_Long  string
	Type      int
	Result    bool
}

type Geo struct {
	Longitude     string
	Latitude      string
	City          string
	Province      string
	City_Name     string
	Province_Name string
	Address       string
	Pinyin        string
	More          string
}

// 其他的常用结构体

type ErrorString struct {
	S string
}

func (e *ErrorString) Error() string {
	return "Gobo错误：" + e.S
}

type WeiboError struct {
	Err        string `json:"Error"`
	Error_Code int64
	Request    string
}

func (e WeiboError) Error() string {
	return fmt.Sprintf("微博API访问错误 %d [%s] %s", e.Error_Code, e.Request, e.Err)
}

type AccessToken struct {
	Access_Token string
	Remind_In    string
	Expires_In   int
	Uid          string
}

type AccessTokenInfo struct {
	Uid        int64
	Appkey     string
	Scope      string
	Created_At int
	Expire_In  int
}

type Statuses struct {
	Statuses []*Status
}

type Visible struct {
	Type    int
	List_Id int
}

type Pic_Url struct {
	Thumbnail_Pic string
}
