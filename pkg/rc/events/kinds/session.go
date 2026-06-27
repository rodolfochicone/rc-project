package kinds

// SessionStatus describes the lifecycle state of a streamed session update.
type SessionStatus string

const (
	// StatusRunning marks an in-flight session update.
	StatusRunning SessionStatus = "running"
	// StatusCompleted marks a completed session.
	StatusCompleted SessionStatus = "completed"
	// StatusFailed marks a failed or canceled session.
	StatusFailed SessionStatus = "failed"
)

// SessionUpdateKind identifies the semantic variant of a session update.
type SessionUpdateKind string

const (
	// UpdateKindUnknown marks an update with no additional semantic classification.
	UpdateKindUnknown SessionUpdateKind = ""
	// UpdateKindUserMessageChunk marks a streamed user message chunk.
	UpdateKindUserMessageChunk SessionUpdateKind = "user_message_chunk"
	// UpdateKindAgentMessageChunk marks a streamed agent message chunk.
	UpdateKindAgentMessageChunk SessionUpdateKind = "agent_message_chunk"
	// UpdateKindAgentThoughtChunk marks a streamed agent thought chunk.
	UpdateKindAgentThoughtChunk SessionUpdateKind = "agent_thought_chunk"
	// UpdateKindToolCallStarted marks the start of a tool call lifecycle.
	UpdateKindToolCallStarted SessionUpdateKind = "tool_call_started"
	// UpdateKindToolCallUpdated marks an update to an existing tool call lifecycle.
	UpdateKindToolCallUpdated SessionUpdateKind = "tool_call_updated"
	// UpdateKindPlanUpdated marks a plan update.
	UpdateKindPlanUpdated SessionUpdateKind = "plan_updated"
	// UpdateKindAvailableCommandsUpdated marks an available commands update.
	UpdateKindAvailableCommandsUpdated SessionUpdateKind = "available_commands_updated"
	// UpdateKindCurrentModeUpdated marks a current mode update.
	UpdateKindCurrentModeUpdated SessionUpdateKind = "current_mode_updated"
)

// ToolCallState describes the lifecycle state of a tool call entry.
type ToolCallState string

const (
	// ToolCallStateUnknown marks a tool call without an explicit lifecycle state.
	ToolCallStateUnknown ToolCallState = ""
	// ToolCallStatePending marks a pending tool call.
	ToolCallStatePending ToolCallState = "pending"
	// ToolCallStateInProgress marks an in-flight tool call.
	ToolCallStateInProgress ToolCallState = "in_progress"
	// ToolCallStateCompleted marks a completed tool call.
	ToolCallStateCompleted ToolCallState = "completed"
	// ToolCallStateFailed marks a failed tool call.
	ToolCallStateFailed ToolCallState = "failed"
	// ToolCallStateWaitingForConfirmation is reserved for future permission-aware UX.
	ToolCallStateWaitingForConfirmation ToolCallState = "waiting_for_confirmation"
)

// SessionPlanEntry describes one plan entry.
type SessionPlanEntry struct {
	Content  string `json:"content"`
	Priority string `json:"priority"`
	Status   string `json:"status"`
}

// SessionAvailableCommand describes one slash-command style action.
type SessionAvailableCommand struct {
	Name         string `json:"name"`
	Description  string `json:"description,omitempty"`
	ArgumentHint string `json:"argument_hint,omitempty"`
}

// SessionUpdate is the public view of one streamed ACP update.
type SessionUpdate struct {
	Kind              SessionUpdateKind         `json:"kind,omitempty"`
	ToolCallID        string                    `json:"tool_call_id,omitempty"`
	ToolCallState     ToolCallState             `json:"tool_call_state,omitempty"`
	Blocks            []ContentBlock            `json:"blocks,omitempty"`
	ThoughtBlocks     []ContentBlock            `json:"thought_blocks,omitempty"`
	PlanEntries       []SessionPlanEntry        `json:"plan_entries,omitempty"`
	AvailableCommands []SessionAvailableCommand `json:"available_commands,omitempty"`
	CurrentModeID     string                    `json:"current_mode_id,omitempty"`
	Usage             Usage                     `json:"usage,omitempty"`
	Status            SessionStatus             `json:"status"`
}

// SessionStartedPayload describes a new attached session.
type SessionStartedPayload struct {
	Index          int    `json:"index"`
	ACPSessionID   string `json:"acp_session_id"`
	AgentSessionID string `json:"agent_session_id,omitempty"`
	Resumed        bool   `json:"resumed,omitempty"`
}

// SessionUpdatePayload carries one streamed session update.
type SessionUpdatePayload struct {
	Index  int           `json:"index"`
	Update SessionUpdate `json:"update"`
}

// Session awaiting-input kinds distinguish why a run paused for the user.
const (
	// AwaitingInputKindPermission marks a pause for an ACP permission request,
	// where the user selects one of Options.
	AwaitingInputKindPermission = "permission"
	// AwaitingInputKindQuestion marks a pause for a free-text skill question,
	// where the user answers with text (Options is empty).
	AwaitingInputKindQuestion = "question"
)

// SessionInputOption is one selectable choice offered to the user for a
// permission request, mapped from an ACP permission option.
type SessionInputOption struct {
	OptionID string `json:"option_id"`
	Label    string `json:"label,omitempty"`
}

// SessionAwaitingInputPayload describes a run paused waiting for user input,
// either an ACP permission request or a free-text skill question. The pause is
// resolved when a response carrying the matching PromptID is delivered.
type SessionAwaitingInputPayload struct {
	Index        int    `json:"index"`
	ACPSessionID string `json:"acp_session_id,omitempty"`
	// Kind is one of the AwaitingInputKind* values.
	Kind string `json:"kind"`
	// PromptID correlates this pause with the response that resolves it.
	PromptID string `json:"prompt_id"`
	// Text is the permission or question prompt shown to the user.
	Text string `json:"text,omitempty"`
	// Options are the selectable choices for a permission request; it is empty
	// for a free-text question.
	Options []SessionInputOption `json:"options,omitempty"`
}

// SessionCompletedPayload describes a completed session.
type SessionCompletedPayload struct {
	Index int   `json:"index"`
	Usage Usage `json:"usage,omitempty"`
}

// SessionFailedPayload describes a failed session.
type SessionFailedPayload struct {
	Index int    `json:"index"`
	Error string `json:"error,omitempty"`
	Usage Usage  `json:"usage,omitempty"`
}
