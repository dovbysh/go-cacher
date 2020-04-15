package cacher

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestCache_GetOrProcess(t *testing.T) {
	c, e := New(5)
	assert.Empty(t, e)

	f := func() (interface{}, error) {
		return 10, nil
	}
	r, e := c.GetOrProcess("k", f)
	assert.Empty(t, e)
	assert.Equal(t, 10, r)
}

func TestCache_GetOrProcessFShouldCallOnce(t *testing.T) {
	c, e := New(5)
	assert.Empty(t, e)

	conter := 0
	f := func() (interface{}, error) {
		conter = conter + 1
		return 10, nil
	}
	r, e := c.GetOrProcess("k", f)
	assert.Empty(t, e)
	assert.Equal(t, 10, r)
	assert.Equal(t, 1, conter)
	r, e = c.GetOrProcess("k", f)
	assert.Empty(t, e)
	assert.Equal(t, 10, r)
	assert.Equal(t, 1, conter)

	c.Purge()
	r, e = c.GetOrProcess("k", f)
	assert.Empty(t, e)
	assert.Equal(t, 10, r)
	assert.Equal(t, 2, conter)
}

func TestCache_GetOrProcessRaceCondition(t *testing.T) {
	c, e := New(5)
	assert.Empty(t, e)

	var conter int32
	f := func() (interface{}, error) {
		atomic.AddInt32(&conter, 1)
		time.Sleep(time.Millisecond)
		return 10, nil
	}
	var wg sync.WaitGroup
	for i := 0; i < 100; i = i + 1 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			r, e := c.GetOrProcess("k", f)
			assert.Empty(t, e)
			assert.Equal(t, 10, r)
			assert.Equal(t, int32(1), atomic.LoadInt32(&conter))
		}()
	}
	for i := 0; i < 3; i = i + 1 {
		_, e = c.GetOrProcess("kf", func() (interface{}, error) {
			return nil, fmt.Errorf("e")
		})
		assert.Error(t, e)
	}
	wg.Wait()
}
