package conc

import (
	"context"
	"sync"
	"time"

	"github.com/tnnmigga/core/algorithm"
	"github.com/tnnmigga/core/infra/zlog"
	"github.com/tnnmigga/core/utils"
)

var (
	rootCtx, cancelGo = context.WithCancel(context.Background())
	wg                = &sync.WaitGroup{}
	wkg               = newWorkerGroup()
	running           = algorithm.NewCounter[string]()
)

func newWorkerGroup() *workerGroup {
	return &workerGroup{
		workerPool: sync.Pool{
			New: func() any {
				return &worker{
					pending: make(chan func(), 256),
				}
			},
		},
	}
}

type gocall interface {
	func(context.Context) | func()
}

type workerGroup struct {
	group      sync.Map
	workerPool sync.Pool
	mu         sync.Mutex
}

func (wkg *workerGroup) run(name string, fn func()) {
	wkg.mu.Lock()
	var w *worker
	value, ok := wkg.group.Load(name)
	if !ok {
		w = wkg.workerPool.Get().(*worker)
		w.name = name
		wkg.group.Store(name, w)
	} else {
		w = value.(*worker)
	}
	w.count++
	pending := w.count
	wkg.mu.Unlock()
	w.pending <- fn
	if pending == 1 {
		Go(w.work)
	}
}

type worker struct {
	name    string
	pending chan func()
	count   int32
}

func (w *worker) work() {
	for {
		select {
		case fn := <-w.pending:
			utils.ExecAndRecover(fn)
			w.count--
		default:
			wkg.mu.Lock()
			var empty bool
			if w.count == 0 {
				wkg.group.Delete(w.name)
				wkg.workerPool.Put(w)
				empty = true
			}
			wkg.mu.Unlock()
			if empty {
				return
			}
		}
	}
}

// 开启一个受到一定监督的协程
// 若fn参数包含context.Context类型, 当系统准备退出时, 此ctx会Done, 此时必须退出协程
// 系统会等候所有由Go开辟的协程退出后再退出
func Go[T gocall](fn T) {
	switch f := any(fn).(type) {
	case func(context.Context):
		wg.Add(1)
		go func() {
			// GoRunMark(utils.FuncName(fn))
			// defer GoDoneMark(utils.FuncName(fn))
			defer utils.RecoverPanic()
			defer wg.Done()
			f(rootCtx)
		}()
	case func():
		wg.Add(1)
		go func() {
			// GoRunMark(utils.FuncName(fn))
			// defer GoDoneMark(utils.FuncName(fn))
			defer utils.RecoverPanic()
			defer wg.Done()
			f()
		}()
	}
}

// 规则同Go, 但是可以通过name参数对协程进行分组
// 同分组下的任务会等候上一个执行完毕后再执行
func GoWithGroup(name string, fn func()) {
	wkg.run(name, fn)
}

// 等候所有由Go开辟的协程退出
func WaitGoDone(maxWaitTime time.Duration) {
	cancelGo()
	c := make(chan struct{}, 1)
	timer := time.After(maxWaitTime)
	go func() {
		defer utils.RecoverPanic()
		wg.Wait()
		c <- struct{}{}
	}()
	select {
	case <-c:
		return
	case <-timer:
		// PrintCurrentGo()
		zlog.Errorf("wait goroutine exit timeout")
	}
}

func GoRunMark(key string) {
	running.Change(key, 1)
}

func GoDoneMark(key string) {
	running.Change(key, -1)
}

func PrintCurrentGo() {
	running.Range(func(s string, i int) {
		if i > 0 {
			zlog.Debugf("Go %s: %d", s, i)
		}
	})
}
