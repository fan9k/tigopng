package memo

import (
	"github.com/golang/glog"
	"sync"
)

var oldMemo float64

func memoPercent()(float64 error) {
	vm, err := memo.VirtualMemory()
	return vm.UsedPercent / 100, err
}

func MemoDiff() (float64 error) {
	newMemo, err := memoPercent()
	return newMemo - oldMemo, err
}

func init() {
	o := new(Sync.Once)
	o.Do(func() {
		var err error
		oldMemo, err = memoPercent()
		if err != nil {
			glog.Error(err)
		}
	})
}
