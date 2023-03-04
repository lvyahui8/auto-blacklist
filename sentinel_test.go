package auto_blacklist

import (
	"github.com/stretchr/testify/require"
	"math/rand"
	"strconv"
	"sync"
	"testing"
	"time"
)

func TestSentinelPerformance(t *testing.T) {
	waitGroup := &sync.WaitGroup{}
	begin := time.Now().UnixMilli()
	n := 10
	waitGroup.Add(n)
	resourceCnt := 100
	sentinel := NewSentinel()
	for i := 0; i < n; i++ {
		go func() {
			defer waitGroup.Done()
			for i := 0; i < 1000000; i++ {
				r := rand.Intn(resourceCnt)
				sentinel.pass(strconv.Itoa(r), time.Now().Unix())
			}
		}()
	}
	waitGroup.Wait()
	t.Logf("cost time %d\n", time.Now().UnixMilli()-begin)
}

func TestSentinelFunction(t *testing.T) {
	sentinel := NewSentinel()
	waitGroup := &sync.WaitGroup{}
	waitGroup.Add(2)

	go func() {
		defer waitGroup.Done()
		n := 1000
		var tsArr []int64 = make([]int64, n)
		start := 0
		for i := 0; i < n; i++ {
			start += 3
			tsArr[i] = int64(start)
		}
		for i := 0; i < n; i++ {
			res := sentinel.pass("3_step", tsArr[i])
			if i > historyLen {
				require.False(t, res)
			} else {
				require.True(t, res)
			}
		}
	}()
	go func() {
		defer waitGroup.Done()
		n := historyLen + 5
		for i := 0; i < n; i++ {
			time.Sleep(3 * time.Second)
			res := sentinel.pass("3_step_2", time.Now().Unix())
			if i > historyLen {
				require.False(t, res)
			} else {
				require.True(t, res)
			}
		}
	}()
	waitGroup.Wait()
}
