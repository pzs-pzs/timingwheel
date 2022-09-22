package timingwheel

import (
	"errors"
	"sync"
	"time"
)

var (
	ErrInvalidInterval = errors.New("invalid time wheel interval time")
	ErrInvalidSize     = errors.New("invalid time wheel size")
	ErrClosedTimeWheel = errors.New("time wheel is closed")
)

type Exec func(key string)

type task struct {
	key    string
	exec   Exec
	round  int
	remove bool
	pos    int
	circle bool
	gap    int
	async  bool
}

type TimingWheel struct {
	// precision
	interval    time.Duration
	ticker      *time.Ticker
	slots       []map[string]*task
	tasks       map[string]*task
	pos         int
	size        int
	addTaskCh   chan *task
	moveTaskCh  chan *task
	onceClose   sync.Once
	rwMutex     *sync.RWMutex
	closeSignal chan struct{}
}

func NewTimingWheel(interval time.Duration, size int) (*TimingWheel, error) {
	if interval < 0 {
		return nil, ErrInvalidInterval
	}
	if size < 0 {
		return nil, ErrInvalidSize
	}
	slots := make([]map[string]*task, size, size)
	for i := 0; i < size; i++ {
		slots[i] = make(map[string]*task)
	}
	return &TimingWheel{
		interval:    interval,
		ticker:      time.NewTicker(interval),
		slots:       slots,
		pos:         size - 1,
		size:        size,
		addTaskCh:   make(chan *task, 2),
		moveTaskCh:  make(chan *task, 2),
		closeSignal: make(chan struct{}),
		tasks:       make(map[string]*task),
		rwMutex:     &sync.RWMutex{},
	}, nil
}

func (tw *TimingWheel) Stop() {
	tw.onceClose.Do(func() {
		close(tw.closeSignal)
	})
}

func (tw *TimingWheel) Start() {
	go func() {
		defer tw.ticker.Stop()
		for {
			select {
			case <-tw.ticker.C:
				tw.onTick()
			case task := <-tw.addTaskCh:
				tw.addTask(task)
			case <-tw.closeSignal:
				tw.ticker.Stop()
			}
		}
	}()
}

func (tw *TimingWheel) RemoveTask(key string) {
	defer tw.rwMutex.RUnlock()
	tw.rwMutex.RLock()
	if task, ok := tw.tasks[key]; ok {
		task.remove = true
	}
}

func (tw *TimingWheel) AddOnceTask(delay time.Duration, key string, async bool, exec Exec) error {
	return tw.addInnerTask(delay, key, async, false, exec)
}

func (tw *TimingWheel) AddPeriodTask(delay time.Duration, key string, async bool, exec Exec) error {
	return tw.addInnerTask(delay, key, async, true, exec)
}

func (tw *TimingWheel) addInnerTask(delay time.Duration, key string, async, circle bool, exec Exec) error {
	if delay < tw.interval {
		delay = tw.interval
	}
	round := int(delay) / (tw.size * int(tw.interval))
	gap := int(delay / tw.interval)
	pos := (tw.pos + gap) % tw.size
	t := &task{
		key:    key,
		exec:   exec,
		round:  round,
		remove: false,
		circle: circle,
		pos:    pos,
		async:  async,
		gap:    gap,
	}
	select {
	case <-tw.closeSignal:
		return ErrClosedTimeWheel
	default:
		tw.addTaskCh <- t
	}
	func() {
		defer tw.rwMutex.RUnlock()
		tw.rwMutex.RLock()
		tw.tasks[key] = t
	}()
	return nil
}

func (tw *TimingWheel) onTick() {
	tw.pos = (tw.pos + 1) % tw.size
	for _, task := range tw.slots[tw.pos] {
		if task.remove {
			tw.removeTask(task)
			continue
		}
		if task.round > 0 {
			task.round--
			continue
		}
		if task.async {
			// TODO 控制并发数
			go task.exec(task.key)
		} else {
			task.exec(task.key)
		}
		if task.circle {
			tw.moveTask(task)
			continue
		}
		task.remove = true
	}
}

func (tw *TimingWheel) moveTask(t *task) {

	pos := (t.pos + t.gap) % tw.size
	round := t.gap / tw.size
	task := &task{
		key:    t.key,
		exec:   t.exec,
		round:  round,
		remove: t.remove,
		pos:    pos,
		circle: t.circle,
		gap:    t.gap,
	}
	t.remove = true
	tw.removeTask(t)
	tw.addTask(task)
}

func (tw *TimingWheel) removeTask(task *task) {
	defer tw.rwMutex.Unlock()
	tw.rwMutex.Lock()
	delete(tw.slots[task.pos], task.key)
}

func (tw *TimingWheel) addTask(t *task) {
	// immediately
	if t.round < 1 && t.pos < (tw.pos+1)%tw.size && !t.circle {
		t.exec(t.key)
		return
	}
	tw.slots[t.pos][t.key] = t
}
