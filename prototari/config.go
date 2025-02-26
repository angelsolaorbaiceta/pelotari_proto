package prototari

import "time"

const (
	defaultMaxPeers          int           = 64
	defaultBroadcastInterval time.Duration = 5 * time.Second

	BroadcastPort = 21451
	UnicastPort   = 21450

	connReadTimeout time.Duration = 200 * time.Millisecond

	defaultInactivePeerTime               = 10 * time.Second
	heartbeatInterval       time.Duration = 10 * time.Second
)

// Config is the set of parameters that modify the protocol's behaviour.
type Config struct {
	// MaxPeers is the maximum number of peers that the running protocol will
	// accept. Once the maximum number of peers is registered, no more peers
	// can be added.
	MaxPeers int
	// BroadcastInterval is the time between discovery broadcast messages.
	BroadcastInterval time.Duration

	// InactivePeerTime is the maximum allowed time elapsed before a heartbeat
	// message is sent to check if the peer is still alive.
	InactivePeerTime time.Duration
}

// MakeDefaultConfig returns a configuration whose parameters are adjusted using
// the protocol defined defaults.
func MakeDefaultConfig() Config {
	return Config{
		MaxPeers:          defaultMaxPeers,
		BroadcastInterval: time.Duration(defaultBroadcastInterval),
		InactivePeerTime:  defaultInactivePeerTime,
	}
}

func makeTestingConfig() Config {
	// This is a time longer than a single test debug session should take
	// so that events using this timeout never happen inside tests.
	longTime := time.Duration(10 * time.Minute)

	return Config{
		MaxPeers:          1,
		BroadcastInterval: longTime,
		InactivePeerTime:  longTime,
	}
}
