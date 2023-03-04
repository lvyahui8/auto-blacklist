package auto_blacklist

import (
	"sync/atomic"
)

/*
- 加入黑名单。连续出现10次，每次时间间隔固定，可以认为是定时资源，如果间隔时间小于5分钟，则拉黑。
*/

const historyLen = 5
const escapeInterval = 5 * 60

type Resource struct {
	key      string
	disabled bool
	last     int64
	interval int64
	times    int64
}

func NewResource(resourceKey string) *Resource {
	return &Resource{
		key:      resourceKey,
		disabled: false,
		last:     -1,
		interval: 0,
		times:    1,
	}
}

func (r *Resource) scroll(ts int64) {
	it := ts - r.last
	if it == r.interval {
		atomic.AddInt64(&r.times, 1)
		if r.times > historyLen && r.interval <= escapeInterval {
			r.disabled = true
		}
	} else {
		atomic.StoreInt64(&r.times, 1)
	}
	r.interval = it
	r.last = ts
}

type Sentinel struct {
	resourceMap *ConcurrentHashMap
}

func NewSentinel() *Sentinel {
	return &Sentinel{
		resourceMap: NewConcurrentHashMap(2048),
	}
}

func (s *Sentinel) getResource(resourceKey string) *Resource {
	var str = String(resourceKey)
	o := s.resourceMap.Get(str)
	var r *Resource
	if o != nil {
		r = o.(*Resource)
	} else {
		r = s.resourceMap.PutIfNotExists(str, NewResource(resourceKey)).(*Resource)
	}
	return r
}

func (s *Sentinel) pass(resourceKey string, ts int64) bool {
	r := s.getResource(resourceKey)
	if r.disabled {
		// 拉黑禁用
		return false
	}
	r.scroll(ts)
	return !r.disabled
}

func latin1StringIntHash(str string) int {
	h := 0
	for i := range str {
		ch := str[i]
		h = 31*h + int(ch&0xff)
	}
	return h
}

type String string

func (s String) Hash() int {
	return latin1StringIntHash(string(s))
}
