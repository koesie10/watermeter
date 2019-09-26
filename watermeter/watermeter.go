package watermeter

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/warthog618/gpio"
)

type WaterMeter struct {
	numRises        int64
	onIncrement     []func(int64)
	onIncrementLock sync.Mutex
}

func New(p *gpio.Pin) (*WaterMeter, error) {
	wm := &WaterMeter{
		numRises: -1,
	}

	if err := p.Watch(gpio.EdgeBoth, wm.onRise); err != nil {
		return nil, fmt.Errorf("failed to watch pin: %w", err)
	}

	return wm, nil
}

func (wm *WaterMeter) RegisterWatcher(handler func(int64)) error {
	wm.onIncrementLock.Lock()
	defer wm.onIncrementLock.Unlock()
	wm.onIncrement = append(wm.onIncrement, handler)

	return nil
}

func (wm *WaterMeter) NumRises() int64 {
	return wm.numRises
}

func (wm *WaterMeter) onRise(p *gpio.Pin) {
	// We will always get called initially
	if wm.numRises == -1 {
		wm.numRises = 0
		return
	}

	time.Sleep(time.Second / 20) // 0.05 seconds

	// Filter out false positives of power fluctuation
	if p.Read() == gpio.Low {
		return
	}

	atomic.AddInt64(&wm.numRises, 1)

	wm.onIncrementLock.Lock()
	defer wm.onIncrementLock.Unlock()

	for _, v := range wm.onIncrement {
		v(wm.numRises)
	}
}
