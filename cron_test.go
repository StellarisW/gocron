package gocron

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

const OneSecond = 1*time.Second + 10*time.Millisecond

func wait(wg *sync.WaitGroup) chan bool {
	ch := make(chan bool)
	go func() {
		wg.Wait()
		ch <- true
	}()
	return ch
}

func stop(cron *Cron) chan bool {
	ch := make(chan bool)
	go func() {
		cron.Stop()
		ch <- true
	}()
	return ch
}

func TestStopCausesJobsToNotRun(t *testing.T) {
	wg := &sync.WaitGroup{}
	wg.Add(1)

	cron := New()
	err := cron.AddFunc("*/5 * * * * ?", func() { wg.Done() })
	if err != nil {
		fmt.Println(err)
	}
	cron.Start()
	time.Sleep(time.Minute * 60)
	cron.Stop()

	select {
	case <-time.After(OneSecond):
		// No job ran!
	case <-wait(wg):
		t.FailNow()
	}
}
