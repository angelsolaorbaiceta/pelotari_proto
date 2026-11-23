package prototari

const (
	// The discoveryMessage is sent by the broadcaster to every peer in the
	// private network to discover who's available.
	discoveryMessage    string = "pelotari?"
	discoveryMessageLen        = len(discoveryMessage)

	// The response message is sent by the responder to signal the broadcaster
	// it is interested in joining it.
	responseMessage    string = "aupa!"
	responseMessageLen        = len(responseMessage)

	// The confirmation message is sent by the broadcaster to the peer that sent
	// it a response message, to confirm that they can talk to each other.
	confirmationMessage    string = "dale!"
	confirmationMessageLen        = len(confirmationMessage)

	// A heartbeatMessage is sent to every peer to make sure they are still online.
	// If a few of these are sent without answer from a peer, the peer is removed
	// from the list of peers.
	heartbeatMessage    string = "hor?"
	heartbeatMessageLen        = len(heartbeatMessage)

	// A heartbeatResMessage is a response to the heartbeat message.
	heartbeatResMessage    string = "hemen nago!"
	heartbeatResMessageLen        = len(heartbeatResMessage)
)
