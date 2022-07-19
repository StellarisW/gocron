package gocron

import (
	"time"

	"github.com/emirpasic/gods/queues/priorityqueue"
	"github.com/emirpasic/gods/utils"
)

type Cron struct {
	entries *priorityqueue.Queue
	add     chan *Entry
	remove  chan EntryID
	stop    chan struct{}
	running bool // 引擎运行标志
}

type EntryID int
type Entry struct {
	*Schedule
	Next time.Time
	Job  Job
}

// Job 工作函数
type Job interface {
	Run()
}

func byTime(a, b interface{}) int {
	return utils.Int64Comparator(
		a.(*Entry).Next.Unix(),
		b.(*Entry).Next.Unix(),
	)
}

func New() *Cron {
	queue := priorityqueue.NewWith(byTime)
	return &Cron{
		entries: queue,
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
		entry.Next = time.Now().AddDate(5, 0, 0)
		c.entries.Enqueue(entry)
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
	size := c.entries.Size()
	for i := 1; i <= size; i++ {
		v, _ := c.entries.Dequeue()
		entry := v.(*Entry)
		entry.Next = entry.Schedule.Next(now)
		c.entries.Enqueue(entry)
	}
	for {

		var effective time.Time
		top, _ := c.entries.Peek()
		if c.entries.Size() == 0 || top.(*Entry).Next.IsZero() {
			// 如果工作池里没有入口函数,引擎直接休眠,但还是可以接受新的请求
			effective = now.AddDate(10, 0, 0)
		} else {
			effective = top.(*Entry).Next
		}

		select {
		case now = <-time.After(effective.Sub(now)):
			for {
				v, _ := c.entries.Dequeue()
				entry := v.(*Entry)
				if entry.Next != effective {
					break
				}
				go entry.Job.Run()
				entry.Next = entry.Schedule.Next(effective)
				c.entries.Enqueue(entry)
			}

		case newEntry := <-c.add:
			newEntry.Next = newEntry.Schedule.Next(now)
			c.entries.Enqueue(newEntry)

		case <-c.stop:
			return
		}
	}
}
