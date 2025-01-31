package prototari

import "time"

const (
	defaultMaxPeers          int           = 64
	defaultBroadcastInterval time.Duration = 5 * time.Second

	BroadcastPort = 21451
	UnicastPort   = 21450
)

// Config is the set of parameters that modify the protocol's behaviour.
type Config struct {
	// MaxPeers is the maximum number of peers that the running protocol will
	// accept. Once the maximum number of peers is registered, no more peers
	// can be added.
	MaxPeers int
	// BroadcastInterval is the time between discovery broadcast messages.
	BroadcastInterval time.Duration
}

// MakeDefaultConfig returns a configuration whose parameters are adjusted using
// the protocol defined defaults.
func MakeDefaultConfig() Config {
	return Config{
		MaxPeers:          defaultMaxPeers,
		BroadcastInterval: time.Duration(defaultBroadcastInterval),
	}
}
