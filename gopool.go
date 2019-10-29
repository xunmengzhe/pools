package pools

import (
	"container/list"
	"fmt"
	"runtime"
	"sync"
	"time"
)

type task struct {
	hander func(data interface{})
	data   interface{}
}

type GoPool struct {
	inited       bool
	destroying   bool
	goroutineNum int
	runningNum   int
	waitingNum   int
	cond         *sync.Cond
	tasks        *list.List
}

func NewGoPool(num int) *GoPool {
	if num <= 0 {
		numCPUs := runtime.NumCPU()
		runtime.GOMAXPROCS(numCPUs)
		num = numCPUs
	}
	pool := &GoPool{
		inited:       true,
		destroying:   false,
		goroutineNum: num,
		runningNum:   0,
		waitingNum:   0,
		cond:         sync.NewCond(new(sync.Mutex)),
		tasks:        list.New(),
	}
	for i := 0; i < num; i++ {
		go pool.worker()
	}
	return pool
}

func (g *GoPool) worker() {
	g.runningNum++
	for !g.destroying {
		g.cond.L.Lock()
		if g.tasks.Len() <= 0 {
			g.waitingNum++
			g.cond.Wait()
			g.waitingNum--
			g.cond.L.Unlock()
			continue
		}
		t := g.tasks.Remove(g.tasks.Front())
		g.cond.L.Unlock()
		tk := t.(*task)
		tk.hander(tk.data)
	}
	g.runningNum--
}

func (g *GoPool) Destroy() {
	if !g.inited {
		return
	}
	g.inited = false
	g.destroying = true
	g.cond.L.Lock()
	g.cond.Broadcast()
	g.cond.L.Unlock()
	timer := time.NewTimer(time.Second)
	for {
		<-timer.C
		if g.runningNum > 0 {
			timer.Reset(time.Second)
		} else {
			break
		}
	}
	timer.Stop()
	for g.tasks.Len() > 0 {
		g.tasks.Remove(g.tasks.Front())
	}
}

func (g *GoPool) AddWorker(data interface{}, hander func(data interface{})) error {
	if !g.inited {
		return fmt.Errorf("Invalid GoPool object.")
	}
	tk := &task{
		data:   data,
		hander: hander,
	}
	g.cond.L.Lock()
	g.tasks.PushBack(tk)
	if g.waitingNum > 0 {
		g.cond.Signal()
	}
	g.cond.L.Unlock()
	return nil
}
