package memo

import (
	"github.com/golang/glog"
	"github.com/shirou/gopsutil/mem"
	"sync"
)

var oldMemo float64

// return the percent of memory usage.
func memoPercent() (float64, error) {
	vm, err := mem.VirtualMemory()
	return vm.UsedPercent / 100, err
}

// return the difference between the old memory and the new one.
func MemoDiff() (float64, error) {
	newMemo, err := memoPercent()
	return newMemo - oldMemo, err
}

func init() {
	o := new(sync.Once)
	o.Do(func() {
		var err error
		oldMemo, err = memoPercent()
		if err != nil {
			glog.Error(err)
		}
	})
}
