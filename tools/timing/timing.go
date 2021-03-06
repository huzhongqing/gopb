package timing

import (
	"encoding/json"
	"sync"
	"time"
)

type Timer struct {
	start time.Time
}

func NewTimer() *Timer {
	timer := &Timer{}
	return timer
}

func (timer *Timer) Reset() {
	timer.start = time.Now().Local()
}

func (timer *Timer) Duration() int64 {
	now := time.Now().Local()
	nanos := now.Sub(timer.start).Nanoseconds()
	micros := nanos / 1000
	return micros
}

type Timing struct {
	Enable  bool
	mutex   sync.Mutex
	methods map[string]*Waiting
}

func NewTiming(enable bool) *Timing {
	t := &Timing{
		Enable:  enable,
		mutex:   sync.Mutex{},
		methods: make(map[string]*Waiting),
	}
	return t
}

type Waiting struct {
	Count int64
	Total int64
	Max   int64
	Min   int64
}

func (t *Timing) Do(method string, f func()) {
	if !t.enable() {
		f()
		return
	}
	start := time.Now()
	f()
	t.waiting(method, time.Now().Sub(start).Nanoseconds())
}

func (t *Timing) waiting(method string, nano int64) {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	w, ok := t.methods[method]
	if !ok {
		t.methods[method] = &Waiting{
			Count: 1,
			Total: nano,
			Max:   nano,
			Min:   nano,
		}
		return
	}
	w.Total += nano
	w.Count++

	if nano > w.Max {
		w.Max = nano
	}
	if nano < w.Min {
		w.Min = nano
	}
}

type MethodData struct {
	Method        string `json:"method"`
	DoCount       int64  `json:"do_count"`
	TimeNanoTotal int64  `json:"time_nano_total"`
	MaxTimeNano   int64  `json:"max_time_nano"`
	MinTimeNano   int64  `json:"min_time_nano"`
	AvgTimeNano   int64  `json:"avg_time_nano"`
}

func (d MethodData) String() string {
	v, _ := json.Marshal(d)
	return string(v)
}

func (t *Timing) GetMethodData() []*MethodData {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	data := make([]*MethodData, 0, len(t.methods))
	for k, v := range t.methods {
		d := &MethodData{
			Method:        k,
			DoCount:       v.Count,
			TimeNanoTotal: v.Total,
			MaxTimeNano:   v.Max,
			MinTimeNano:   v.Min,
			AvgTimeNano:   v.Total / v.Count,
		}
		data = append(data, d)
	}

	return data
}

func (t *Timing) SetEnable(enable bool) {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	t.Enable = enable
}

func (t *Timing) Clear() {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	t.methods = make(map[string]*Waiting)
	return
}

func (t *Timing) enable() bool {
	return t.Enable
}

var std = NewTiming(true)

func Do(method string, f func()) {
	if method == "" || f == nil {
		return
	}
	std.Do(method, f)
}

func SetEnable(enable bool) {
	std.SetEnable(enable)
}

func Clear() {
	std.Clear()
}

func GetMethodData() []*MethodData {
	return std.GetMethodData()
}
