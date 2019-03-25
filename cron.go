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
package cron

import (
	"github.com/satori/go.uuid"
	"runtime"
	"sort"
	"strings"
	"sync"
	"time"
)

type Cron struct {
	entries  []*Entry
	add      chan *Entry
	stop     chan struct{}
	delete   chan string
	running  bool
	location *time.Location
	mutex    sync.Mutex
}

// Start 开始定时器,如果定时器未开启,则开启一个协程运行定时器,否则什么都不做
func (c *Cron) Start() {
	if c.running {
		return
	}
	c.running = true
	go c.run()
}

// Run 运行定时器,如果定时器未开启,这个方法会阻塞,否则什么都不做
func (c *Cron) Run() {
	if c.running {
		return
	}
	c.running = true
	c.run()
}

// Stop 停止定时器
func (c *Cron) Stop() {
	if !c.running {
		return
	}
	c.stop <- struct{}{}
	c.running = false
}

func (c *Cron) AddFunc(spec string, cmd func()) ([]byte, error) {
	return c.AddJob(spec, FuncJob(cmd))
}

func (c *Cron) AddJob(spec string, cmd Job) ([]byte, error) {
	schedule, err := Parse(spec)
	if err != nil {
		return nil, err
	}
	return []byte(c.Schedule(schedule, cmd)), nil
}

// DeleteJob 通过传入任务ID删除任务,这个方法是线程安全
func (c *Cron) DeleteJob(id string) {
	if !c.running {
		c.deleteEntry(id)
	}
	c.delete <- id
}

func (c *Cron) Schedule(schedule Schedule, cmd Job) string {
	id := uuid.NewV4().String()
	entry := &Entry{
		ID:       id,
		Schedule: schedule,
		Job:      cmd,
	}
	if !c.running {
		c.entries = append(c.entries, entry)
	}
	c.add <- entry
	return id
}

func (c *Cron) run() {
	now := c.now()
	for _, entry := range c.entries {
		entry.Next = entry.Schedule.Next(now)
	}

	for {
		sort.Sort(byTime(c.entries))

		var timer *time.Timer
		if len(c.entries) == 0 || c.entries[0].Next.IsZero() {
			timer = time.NewTimer(100000 * time.Hour)
		} else {
			timer = time.NewTimer(c.entries[0].Next.Sub(now))
		}

		for {
			select {
			case now = <-timer.C:
				now = now.In(c.location)
				for _, e := range c.entries {
					if e.Next.After(now) || e.Next.IsZero() {
						break
					}
					go c.runWithRecovery(e.Job)
					e.Prev = e.Next
					e.Next = e.Schedule.Next(now)
				}
			case newEntry := <-c.add:
				timer.Stop()
				now = c.now()
				newEntry.Next = newEntry.Schedule.Next(now)
				c.entries = append(c.entries, newEntry)
			case deleteID := <-c.delete:
				c.deleteEntry(deleteID)
			case <-c.stop:
				timer.Stop()
				return
			}
		}
	}
}

func (c *Cron) deleteEntry(id string) {
	newEntries := make([]*Entry, 0)
	for _, entry := range c.entries {
		if strings.Compare(entry.ID, id) != 0 {
			newEntries = append(newEntries, entry)
		}
	}
	c.entries = newEntries
}

func (c Cron) runWithRecovery(job Job) {
	defer func() {
		if r := recover(); r != nil {
			const size = 64 << 10 // 分配一个足够大的存储获取之后截断
			buf := make([]byte, size)
			buf = buf[:runtime.Stack(buf, false)]
		}
	}()
	job.Run()
}

func (c Cron) now() time.Time {
	return time.Now().In(c.location)
}

// New 返回以本地时区为基准的定时器
func New() *Cron {
	return NewWithLocation(time.Now().Location())
}

// NewWithLocation 返回以指定时区为基准的定时器
func NewWithLocation(location *time.Location) *Cron {
	return &Cron{
		entries:  make([]*Entry, 0),
		add:      make(chan *Entry),
		stop:     make(chan struct{}),
		delete:   make(chan string),
		running:  false,
		location: location,
	}
}
