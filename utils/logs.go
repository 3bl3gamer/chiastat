package utils

import "time"

type SyncInterval struct {
	minInterval   int64
	handler       func()
	lastCallStamp int64
}

func NewSyncInterval(minInterval time.Duration, handler func()) *SyncInterval {
	return &SyncInterval{minInterval: int64(minInterval / time.Nanosecond), handler: handler}
}

func (p *SyncInterval) Trigger() {
	now := time.Now().UnixNano()
	if now-p.lastCallStamp >= p.minInterval {
		p.handler()
		p.lastCallStamp = now
	}
}
