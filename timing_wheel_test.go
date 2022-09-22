package timingwheel

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestNewTimingWheel(t *testing.T) {
	tw, err := NewTimingWheel(2*time.Second, 10)
	assert.Nil(t, err)
	assert.NotNil(t, tw)
}

func TestOnceTask(t *testing.T) {
	tw, err := NewTimingWheel(2*time.Second, 10)
	assert.Nil(t, err)
	assert.NotNil(t, tw)
	tw.Start()
	before := time.Now().Unix()
	err = tw.AddOnceTask(4*time.Second, "test1", false, func(key string) {
		println(time.Now().Unix() - before)
		println("test1")
	})
	assert.Nil(t, err)
	err = tw.AddOnceTask(7*time.Second, "test2", false, func(key string) {
		println(time.Now().Unix() - before)
		println("test2")
	})
	assert.Nil(t, err)
	tw.RemoveTask("test2")

	err = tw.AddOnceTask(32*time.Second, "test3", false, func(key string) {
		println(time.Now().Unix() - before)
		println("test3")
	})
	assert.Nil(t, err)
	<-make(chan struct{})
}

func TestPeriodTask(t *testing.T) {
	tw, err := NewTimingWheel(1*time.Second, 60)
	assert.Nil(t, err)
	assert.NotNil(t, tw)
	tw.Start()
	before := time.Now().Unix()
	err = tw.AddPeriodTask(2*time.Second, "test1", false, func(key string) {
		println(time.Now().Unix() - before)
		println("test1")
	})
	time.Sleep(6 * time.Second)
	tw.Stop()
	<-make(chan struct{})
}
