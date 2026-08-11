package eventbus

// DefaultPublishPort is the conventional local hub publish HTTP port.
const DefaultPublishPort = 23891

// Locked v1 event type strings (wire vocabulary).
const (
	TypeSeatalkMessageReceived = "seatalk.message.received"
	TypeSeatalkSessionOpened   = "seatalk.session.opened"
	TypeAgentTTYStarted        = "agent.tty.started"
	// TypeAgentTTYRestarted is emitted when a session already has a TTY path
	// and activity warrants local re-attach (follow-up send or resume).
	TypeAgentTTYRestarted = "agent.tty.restarted"
)

// Locked payload "reason" values for agent.tty.started / agent.tty.restarted.
const (
	ReasonTTYNew      = "new"      // first live TTY (agent.tty.started)
	ReasonTTYFollowup = "followup" // live send into existing TTY
	ReasonTTYResume   = "resume"   // resume / reclaim after runner exited
)

// Locked v1 source identifiers (wire vocabulary).
const (
	SourceSeatalkLocalBot = "seatalk.local-bot"
	SourceAgentRun        = "agent-run"
)
