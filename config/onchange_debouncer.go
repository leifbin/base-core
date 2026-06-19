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
		stop:   make(chan struct{}),
		done:   make(chan struct{}),
	}
	go d.run()
	return d
}

// Stop 停止防抖 goroutine，等待退出后返回。
func (d *Debouncer[T]) Stop() {
	d.once.Do(func() { // ← 确保只 close 一次
		close(d.stop)
		<-d.done
	})
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
	defer close(d.done) // ← 退出时通知
	var latestCfg T
	for {
		select {
		case cfg := <-d.ch:
			latestCfg = cfg
			timer := time.NewTimer(d.delay)
			// 内层 select 也要加上 stop 检测
			innerLoop := true
			for innerLoop {
				select {
				case newCfg := <-d.ch:
					latestCfg = newCfg
					if !timer.Stop() {
						<-timer.C
					}
					timer.Reset(d.delay)
				case <-timer.C:
					d.onFire(latestCfg)
					innerLoop = false
				case <-d.stop: // ← 收到停止信号
					timer.Stop()
					return
				}
			}
		case <-d.stop: // ← 收到停止信号
			return
		}
	}
}
