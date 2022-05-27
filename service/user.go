package service

import (
	"MedalHelper/dto"
	"MedalHelper/manager"
	"MedalHelper/util"
	"sync"

	"github.com/TwiN/go-color"
	"github.com/tidwall/gjson"
)

type User struct {
	// 用户ID
	Uid int
	// 用户名称
	Name string
	// 是否登录
	isLogin bool

	// 登录凭证
	accessKey string
	// 被禁止的房间ID
	bannedUIDs []int
	// 用户所有勋章
	medals []dto.MedalInfo
	// 用户等级小于20的勋章
	medalsLow []dto.MedalInfo
	// 今日亲密度没满的勋章
	remainMedals []dto.MedalInfo
	// 最大重试次数
	retryTimes int32
}

func NewUser(accessKey string, uids []int) User {
	return User{
		accessKey:    accessKey,
		bannedUIDs:   uids,
		medals:       make([]dto.MedalInfo, 0, 10),
		medalsLow:    make([]dto.MedalInfo, 0, 10),
		remainMedals: make([]dto.MedalInfo, 0, 10),
		retryTimes:   10,
	}
}

func (user User) info(format string, v ...interface{}) {
	format = color.Green + "[INFO] " + color.Reset + format
	format = color.Reset + color.Blue + user.Name + color.Reset + " " + format
	util.PrintColor(format, v...)
}

func (user *User) loginVerify() bool {
	resp, err := manager.LoginVerify(user.accessKey)
	if err != nil || resp.Data.Mid == 0 {
		user.isLogin = false
		return false
	}
	user.Uid = resp.Data.Mid
	user.Name = resp.Data.Name
	user.isLogin = true
	user.info("登录成功")
	return true
}

func (user *User) signIn() error {
	signInfo, err := manager.SignIn(user.accessKey)
	if err != nil {
		return nil
	}
	resp := gjson.Parse(signInfo)
	if resp.Get("code").Int() == 0 {
		signed := resp.Get("data.hadSignDays").String()
		all := resp.Get("data.allDays").String()
		user.info("签到成功, 本月签到次数: %s/%s", signed, all)
	} else {
		user.info("%s", resp.Get("message").String())
	}

	userInfo, err := manager.GetUserInfo(user.accessKey)
	if err != nil {
		return nil
	}
	level := userInfo.Data.Exp.UserLevel
	unext := userInfo.Data.Exp.Unext
	user.info("当前用户UL等级: %d, 还差 %d 经验升级", level, unext)
	return nil
}

func (user *User) setMedals() {
	medals := manager.GetFansMedalAndRoomID(user.accessKey)
	for _, medal := range medals {
		if util.IntContain(user.bannedUIDs, medal.Medal.TargetID) != -1 {
			continue
		}
		if medal.RoomInfo.RoomID == 0 {
			continue
		}
		user.medals = append(user.medals, medal)
		if medal.Medal.Level < 20 {
			user.medalsLow = append(user.medalsLow, medal)
			if medal.Medal.TodayFeed < 1200 {
				user.remainMedals = append(user.remainMedals, medal)
			}
		}
	}
}

func (user *User) checkMedals() {
	user.setMedals()
	user.info("20级以下牌子共 %d 个, 完成任务 %d 个",
		len(user.medalsLow),
		len(user.medalsLow)-len(user.remainMedals),
	)
}

func (user *User) Init() bool {
	if user.loginVerify() {
		user.signIn()
		user.setMedals()
		return true
	} else {
		util.Error("用户登录失败, accessKey: %s", user.accessKey)
		return false
	}
}

func (user *User) Start(wg *sync.WaitGroup) {
	if user.isLogin {
		task := NewTask(*user, []IAction{
			&Like{},
			&Share{},
			&Danmaku{},
		})
		task.Start()
		user.checkMedals()
	} else {
		util.Error("用户未登录, accessKey: %s", user.accessKey)
	}
	wg.Done()
}
