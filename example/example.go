// Copyright 2019 The WIZ Technology Co. Ltd. Authors. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License"): you may
// not use this file except in compliance with the License. You may obtain
// a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
// WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
// License for the specific language governing permissions and limitations
// under the License.
package main

import (
	"fmt"
	"github.com/lamjack/cron"
	"time"
)

func main() {
	c := cron.New()
	c.AddFunc("* * * * * *", func() {
		fmt.Println(time.Now().Format("15:04:05"), "每秒跑一次")
	})
	c.Start()

	// 5秒后动态添加一个计划任务,2秒跑一次
	time.Sleep(5 * time.Second)
	id, _ := c.AddFunc("*/2 * * * * *", func() {
		fmt.Println(time.Now().Format("15:04:05"), "2秒跑一次")
	})

	// 10秒后 删除新添加的任务
	time.Sleep(10 * time.Second)
	c.DeleteJob(string(id))

	select {}
}
