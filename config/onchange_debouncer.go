package config

import (
	"log/slog"
	"time"
)

// 新建 Debouncer
func NewDebouncer[T any](delay time.Duration, onFire func(T)) *Debouncer[T] {
	d := &Debouncer[T]{
		ch:     make(chan T, 1),
		delay:  delay,
		onFire: onFire,
	}
	go d.run()
	return d
}

// 提交一個新配置，會進入防抖流程
func (d *Debouncer[T]) Submit(cfg T) {
	select {
	case d.ch <- cfg:
	default:
		slog.Warn("防抖通道已满，丢弃旧配置", "channel_len", len(d.ch))
	}
}

// 防抖處理 goroutine
func (d *Debouncer[T]) run() {
	var latestCfg T
	for {
		cfg := <-d.ch
		latestCfg = cfg

		timer := time.NewTimer(d.delay)
		for {
			select {
			case newCfg := <-d.ch:
				latestCfg = newCfg
				if !timer.Stop() {
					<-timer.C
				}
				timer.Reset(d.delay)
			case <-timer.C:
				func() {
					defer func() {
						if r := recover(); r != nil {
							slog.Error("防抖回调 panic 已恢复", "recover", r)
						}
					}()
					d.onFire(latestCfg)
				}()
				goto NEXT
			}
		}
	NEXT:
	}
}
