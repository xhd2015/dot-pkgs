package eventbus

// DefaultPublishPort is the conventional local hub publish HTTP port.
const DefaultPublishPort = 23891

// Locked v1 event type strings (wire vocabulary).
const (
	TypeSeatalkMessageReceived = "seatalk.message.received"
	TypeSeatalkSessionOpened   = "seatalk.session.opened"
	TypeAgentTTYStarted        = "agent.tty.started"
)

// Locked v1 source identifiers (wire vocabulary).
const (
	SourceSeatalkLocalBot = "seatalk.local-bot"
	SourceAgentRun        = "agent-run"
)
