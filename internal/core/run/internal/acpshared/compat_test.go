package acpshared

import (
	"github.com/rodolfochicone/rc-project/internal/core/run/transcript"
)

func newSessionUpdateHandler(cfg SessionUpdateHandlerConfig) *SessionUpdateHandler {
	return NewSessionUpdateHandler(cfg)
}

const transcriptEntryAssistantMessage = transcript.EntryKindAssistantMessage
