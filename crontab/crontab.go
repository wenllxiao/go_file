package main

import (
	"fmt"

	"github.com/robfig/cron/v3"
	"github.com/zeromicro/go-zero/core/logx"
)

func main() {
	LoopCrontabTask()
}

// LoopCrontabTask 定时任务
func LoopCrontabTask() {
	c := cron.New(cron.WithSeconds())
	/*spec的格式："second min hour dayOfMonth  month dayOfWeek"*/
	// specDay := "00 00 7 * * *" //每天7:00
	// 	spec := "10 * * * *" //每10分钟
	specDay := "0 */2 * * * *" //每2分钟一次
	_, err := c.AddFunc(specDay, func() {
		task()
	})
	if err != nil {
		logx.Error("start  ticker err:", err)
	}
	c.Start()
	select {}
}

func task() {
	fmt.Println("task")
}
