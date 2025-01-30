package prototari

import "time"

const (
	defaultMaxPeers          int           = 64
	defaultBroadcastInterval time.Duration = 5 * time.Second
)

type Config struct {
	MaxPeers          int
	BroadcastInterval time.Duration
}

func MakeDefaultConfig() Config {
	return Config{
		MaxPeers:          defaultMaxPeers,
		BroadcastInterval: time.Duration(defaultBroadcastInterval),
	}
}
