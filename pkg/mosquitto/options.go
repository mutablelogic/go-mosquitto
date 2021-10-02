package mosquitto

////////////////////////////////////////////////////////////////////////////////
// TYPES

type opts struct {
	qos    int
	retain bool
}

type ClientOpt func(opts *opts)

////////////////////////////////////////////////////////////////////////////////
// GLOBALS

var (
	defaultOpts = opts{
		qos:    0,
		retain: false,
	}
)

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Best efforts delivery
func OptAtMostOnce() ClientOpt {
	return func(opts *opts) {
		opts.qos = 0
	}
}

// Standard delivery
func OptAtLeastOnce() ClientOpt {
	return func(opts *opts) {
		opts.qos = 1
	}
}

// Guaranteed (?) delivery
func OptExactlyOnce() ClientOpt {
	return func(opts *opts) {
		opts.qos = 2
	}
}

// Custom QoS
func OptQoS(qos int) ClientOpt {
	return func(opts *opts) {
		opts.qos = qos
	}
}

// Make the message retained
func OptRetain() ClientOpt {
	return func(opts *opts) {
		opts.retain = true
	}
}
