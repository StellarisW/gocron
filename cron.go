package gocron

import (
	"sort"
	"time"
)

type Cron struct {
	entries []*Entry
	add     chan *Entry
	stop    chan struct{}
	running bool // 引擎运行标志
}

type Entry struct {
	*Schedule
	Next time.Time
	Prev time.Time
	Job  Job
}

type byTime []*Entry

// sort的自定义函数

func (s byTime) Len() int      { return len(s) }
func (s byTime) Swap(i, j int) { s[i], s[j] = s[j], s[i] }
func (s byTime) Less(i, j int) bool {
	if s[i].Next.IsZero() {
		return false
	}
	if s[j].Next.IsZero() {
		return true
	}
	return s[i].Next.Before(s[j].Next)
}

// Job 工作函数
type Job interface {
	Run()
}

func New() *Cron {
	return &Cron{
		entries: nil,
		add:     make(chan *Entry),
		stop:    make(chan struct{}),
		running: false,
	}
}

// 包裹了需要执行的函数
type jobAdapter func()

func (r jobAdapter) Run() { r() }

func (c *Cron) AddFunc(pattern string, cmd func()) (err error) {
	err = c.AddJob(pattern, jobAdapter(cmd))
	return
}

func (c *Cron) AddJob(pattern string, cmd Job) (err error) {
	entry := &Entry{
		Job: cmd,
	}

	entry.Schedule, err = newSchedule(pattern)
	if err != nil {
		return
	}

	if !c.running { // 引擎未运行，添加到工作池里面
		c.entries = append(c.entries, entry)
		return
	}
	c.add <- entry // 引擎运行，直接添加到运行队列
	return nil
}

func (c *Cron) Start() {
	c.running = true
	go c.run()
}

func (c *Cron) Stop() {
	c.stop <- struct{}{}
	c.running = false
}

func (c *Cron) run() {
	now := time.Now().Local()
	for _, entry := range c.entries {
		entry.Next = entry.Schedule.Next(now)
	}
	for {
		// 决定下一个运行的入口函数
		sort.Sort(byTime(c.entries)) //TODO: 可以写个优先队列来调度,暂时用排序来代替

		var effective time.Time
		if len(c.entries) == 0 || c.entries[0].Next.IsZero() {
			// 如果工作池里没有入口函数,引擎直接休眠,但还是可以接受新的请求
			effective = now.AddDate(10, 0, 0)
		} else {
			effective = c.entries[0].Next
		}

		select {
		case now = <-time.After(effective.Sub(now)):
			for _, e := range c.entries {
				if e.Next != effective {
					break
				}
				go e.Job.Run()
				e.Prev = e.Next
				e.Next = e.Schedule.Next(effective)
			}

		case newEntry := <-c.add:
			c.entries = append(c.entries, newEntry)
			newEntry.Next = newEntry.Schedule.Next(now)

		case <-c.stop:
			return
		}
	}
}
