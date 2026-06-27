package contract

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/rodolfochicone/rc-project/internal/contentblock"
	"github.com/rodolfochicone/rc-project/pkg/rc/events"
	"github.com/rodolfochicone/rc-project/pkg/rc/events/kinds"
)

type MutationAcceptedResponse struct {
	Accepted bool `json:"accepted"`
}

type WorkspaceRegisterRequest struct {
	Path string `json:"path"`
	Name string `json:"name,omitempty"`
}

type WorkspaceUpdateRequest struct {
	Name string `json:"name"`
}

type WorkspaceResolveRequest struct {
	Path string `json:"path"`
}

type WorkflowRefRequest struct {
	Workspace string `json:"workspace"`
}

type WorkflowArchiveRequest struct {
	Workspace string `json:"workspace"`
	Force     bool   `json:"force,omitempty"`
}

type TaskRunRequest struct {
	Workspace        string          `json:"workspace"`
	PresentationMode string          `json:"presentation_mode,omitempty"`
	RuntimeOverrides json.RawMessage `json:"runtime_overrides,omitempty"`
}

type ReviewFetchRequest struct {
	Workspace string `json:"workspace"`
	Provider  string `json:"provider,omitempty"`
	PRRef     string `json:"pr_ref,omitempty"`
	Round     *int   `json:"round,omitempty"`
}

type ReviewRunRequest struct {
	Workspace        string          `json:"workspace"`
	PresentationMode string          `json:"presentation_mode,omitempty"`
	RuntimeOverrides json.RawMessage `json:"runtime_overrides,omitempty"`
	Batching         json.RawMessage `json:"batching,omitempty"`
}

type ReviewWatchRequest struct {
	Workspace        string          `json:"workspace"`
	Provider         string          `json:"provider,omitempty"`
	PRRef            string          `json:"pr_ref"`
	UntilClean       bool            `json:"until_clean,omitempty"`
	MaxRounds        int             `json:"max_rounds,omitempty"`
	AutoPush         bool            `json:"auto_push,omitempty"`
	PushRemote       string          `json:"push_remote,omitempty"`
	PushBranch       string          `json:"push_branch,omitempty"`
	PollInterval     string          `json:"poll_interval,omitempty"`
	ReviewTimeout    string          `json:"review_timeout,omitempty"`
	QuietPeriod      string          `json:"quiet_period,omitempty"`
	RuntimeOverrides json.RawMessage `json:"runtime_overrides,omitempty"`
	Batching         json.RawMessage `json:"batching,omitempty"`
}

type SyncRequest struct {
	Workspace    string `json:"workspace,omitempty"`
	Path         string `json:"path,omitempty"`
	WorkflowSlug string `json:"workflow_slug,omitempty"`
}

type ExecRequest struct {
	WorkspacePath    string          `json:"workspace_path"`
	Prompt           string          `json:"prompt"`
	PresentationMode string          `json:"presentation_mode,omitempty"`
	RuntimeOverrides json.RawMessage `json:"runtime_overrides,omitempty"`
	// Interactive opts the run into pausing for user input on permission
	// requests and skill questions instead of auto-approving and finalizing.
	Interactive bool `json:"interactive,omitempty"`
}

type DaemonStatus struct {
	PID            int       `json:"pid"`
	Version        string    `json:"version,omitempty"`
	StartedAt      time.Time `json:"started_at"`
	SocketPath     string    `json:"socket_path,omitempty"`
	HTTPPort       int       `json:"http_port,omitempty"`
	ActiveRunCount int       `json:"active_run_count"`
	WorkspaceCount int       `json:"workspace_count"`
}

type DaemonHealth struct {
	Ready               bool                       `json:"ready"`
	Degraded            bool                       `json:"degraded,omitempty"`
	StartedAt           time.Time                  `json:"started_at,omitempty"`
	UptimeSeconds       int64                      `json:"uptime_seconds,omitempty"`
	ActiveRunCount      int                        `json:"active_run_count,omitempty"`
	ActiveRunsByMode    []DaemonModeCount          `json:"active_runs_by_mode,omitempty"`
	WorkspaceCount      int                        `json:"workspace_count,omitempty"`
	IntegrityIssueCount int                        `json:"integrity_issue_count,omitempty"`
	Databases           DaemonDatabaseDiagnostics  `json:"databases,omitempty"`
	Reconcile           DaemonReconcileDiagnostics `json:"reconcile,omitempty"`
	Details             []HealthDetail             `json:"details,omitempty"`
}

type HealthDetail struct {
	Code     string `json:"code"`
	Message  string `json:"message"`
	Severity string `json:"severity,omitempty"`
}

type DaemonModeCount struct {
	Mode  string `json:"mode"`
	Count int    `json:"count"`
}

type DaemonDatabaseDiagnostics struct {
	GlobalBytes int64 `json:"global_bytes,omitempty"`
	RunDBBytes  int64 `json:"run_db_bytes,omitempty"`
}

type DaemonReconcileDiagnostics struct {
	ReconciledRuns     int    `json:"reconciled_runs,omitempty"`
	CrashEventAppended int    `json:"crash_event_appended,omitempty"`
	CrashEventMissing  int    `json:"crash_event_missing,omitempty"`
	LastRunID          string `json:"last_run_id,omitempty"`
}

type Workspace struct {
	ID              string     `json:"id"`
	RootDir         string     `json:"root_dir"`
	Name            string     `json:"name"`
	FilesystemState string     `json:"filesystem_state"`
	ReadOnly        bool       `json:"read_only"`
	HasCatalogData  bool       `json:"has_catalog_data"`
	WorkflowCount   int        `json:"workflow_count"`
	RunCount        int        `json:"run_count"`
	LastCheckedAt   *time.Time `json:"last_checked_at,omitempty"`
	LastSyncedAt    *time.Time `json:"last_sync_at,omitempty"`
	LastSyncError   string     `json:"last_sync_error,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

type WorkspaceRegisterResult struct {
	Workspace Workspace
	Created   bool
}

type WorkspaceUpdateInput struct {
	Name string `json:"name,omitempty"`
}

type WorkspaceSyncResult struct {
	Checked              int      `json:"checked"`
	Removed              int      `json:"removed"`
	Missing              int      `json:"missing"`
	Synced               int      `json:"synced"`
	WorkflowsPruned      int      `json:"workflows_pruned,omitempty"`
	SnapshotsUpserted    int      `json:"snapshots_upserted"`
	TaskItemsUpserted    int      `json:"task_items_upserted"`
	ReviewRoundsUpserted int      `json:"review_rounds_upserted"`
	ReviewIssuesUpserted int      `json:"review_issues_upserted"`
	Warnings             []string `json:"warnings,omitempty"`
}

type WorkflowSummary struct {
	ID               string              `json:"id"`
	WorkspaceID      string              `json:"workspace_id"`
	Slug             string              `json:"slug"`
	ArchivedAt       *time.Time          `json:"archived_at,omitempty"`
	LastSyncedAt     *time.Time          `json:"last_synced_at,omitempty"`
	TaskCounts       *WorkflowTaskCounts `json:"task_counts,omitempty"`
	CanStartRun      *bool               `json:"can_start_run,omitempty"`
	StartBlockReason string              `json:"start_block_reason,omitempty"`
	ArchiveEligible  *bool               `json:"archive_eligible,omitempty"`
	ArchiveReason    string              `json:"archive_reason,omitempty"`
}

type WorkflowTaskCounts struct {
	Total     int `json:"total"`
	Completed int `json:"completed"`
	Pending   int `json:"pending"`
}

type TaskItem struct {
	ID         string    `json:"id"`
	TaskNumber int       `json:"task_number"`
	TaskID     string    `json:"task_id"`
	Title      string    `json:"title"`
	Status     string    `json:"status"`
	Type       string    `json:"type"`
	DependsOn  []string  `json:"depends_on,omitempty"`
	SourcePath string    `json:"source_path"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type ValidationSuccess struct {
	Valid     bool      `json:"valid"`
	CheckedAt time.Time `json:"checked_at,omitempty"`
}

type ArchiveResult struct {
	Archived             bool       `json:"archived"`
	ArchivedAt           *time.Time `json:"archived_at,omitempty"`
	Forced               bool       `json:"forced,omitempty"`
	CompletedTasks       int        `json:"completed_tasks,omitempty"`
	ResolvedReviewIssues int        `json:"resolved_review_issues,omitempty"`
}

type ReviewFetchResult struct {
	Summary ReviewSummary
	Created bool
}

type ReviewSummary struct {
	WorkflowSlug    string    `json:"workflow_slug"`
	RoundNumber     int       `json:"round_number"`
	Provider        string    `json:"provider,omitempty"`
	PRRef           string    `json:"pr_ref,omitempty"`
	ResolvedCount   int       `json:"resolved_count"`
	UnresolvedCount int       `json:"unresolved_count"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type ReviewRound struct {
	ID              string    `json:"id"`
	WorkflowSlug    string    `json:"workflow_slug"`
	RoundNumber     int       `json:"round_number"`
	Provider        string    `json:"provider,omitempty"`
	PRRef           string    `json:"pr_ref,omitempty"`
	ResolvedCount   int       `json:"resolved_count"`
	UnresolvedCount int       `json:"unresolved_count"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type ReviewIssue struct {
	ID          string    `json:"id"`
	IssueNumber int       `json:"issue_number"`
	Severity    string    `json:"severity"`
	Status      string    `json:"status"`
	SourcePath  string    `json:"source_path"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type SessionEntryKind string

const (
	SessionEntryKindAssistantMessage  SessionEntryKind = "assistant_message"
	SessionEntryKindAssistantThinking SessionEntryKind = "assistant_thinking"
	SessionEntryKindToolCall          SessionEntryKind = "tool_call"
	SessionEntryKindStderrEvent       SessionEntryKind = "stderr_event"
	SessionEntryKindRuntimeNotice     SessionEntryKind = "runtime_notice"
)

type SessionStatus string

type ToolCallState string

type ContentBlockType string

type ContentBlock struct {
	Type ContentBlockType `json:"type"`
	Data json.RawMessage  `json:"-"`
}

func (b ContentBlock) MarshalJSON() ([]byte, error) {
	return contentblock.MarshalEnvelopeJSON(b.Type, b.Data)
}

func (b *ContentBlock) UnmarshalJSON(data []byte) error {
	envelope, err := contentblock.UnmarshalEnvelopeJSON(data, func(ContentBlockType, []byte) error {
		return nil
	})
	if err != nil {
		return err
	}
	b.Type = envelope.Type
	b.Data = envelope.Data
	return nil
}

type SessionPlanEntry struct {
	Content  string `json:"content"`
	Priority string `json:"priority"`
	Status   string `json:"status"`
}

type SessionAvailableCommand struct {
	Name         string `json:"name"`
	Description  string `json:"description,omitempty"`
	ArgumentHint string `json:"argumentHint,omitempty"`
}

type SessionEntry struct {
	ID            string           `json:"id"`
	Kind          SessionEntryKind `json:"kind"`
	Title         string           `json:"title,omitempty"`
	Preview       string           `json:"preview,omitempty"`
	ToolCallID    string           `json:"tool_call_id,omitempty"`
	ToolCallState ToolCallState    `json:"tool_call_state,omitempty"`
	Blocks        []ContentBlock   `json:"blocks,omitempty"`
}

type SessionPlanState struct {
	Entries      []SessionPlanEntry `json:"entries,omitempty"`
	PendingCount int                `json:"pending_count,omitempty"`
	RunningCount int                `json:"running_count,omitempty"`
	DoneCount    int                `json:"done_count,omitempty"`
}

type SessionMetaState struct {
	CurrentModeID     string                    `json:"current_mode_id,omitempty"`
	AvailableCommands []SessionAvailableCommand `json:"available_commands,omitempty"`
	Status            SessionStatus             `json:"status,omitempty"`
}

type SessionViewSnapshot struct {
	Revision int              `json:"revision"`
	Entries  []SessionEntry   `json:"entries,omitempty"`
	Plan     SessionPlanState `json:"plan,omitempty"`
	Session  SessionMetaState `json:"session,omitempty"`
}

type Run struct {
	RunID            string     `json:"run_id"`
	WorkspaceID      string     `json:"workspace_id"`
	WorkflowID       *string    `json:"workflow_id,omitempty"`
	WorkflowSlug     string     `json:"workflow_slug,omitempty"`
	ParentRunID      string     `json:"parent_run_id,omitempty"`
	Mode             string     `json:"mode"`
	Status           string     `json:"status"`
	PresentationMode string     `json:"presentation_mode"`
	StartedAt        time.Time  `json:"started_at"`
	EndedAt          *time.Time `json:"ended_at,omitempty"`
	ErrorText        string     `json:"error_text,omitempty"`
	RequestID        string     `json:"request_id,omitempty"`
}

type RunJobSummary struct {
	Index           int                 `json:"index"`
	CodeFile        string              `json:"code_file,omitempty"`
	CodeFiles       []string            `json:"code_files,omitempty"`
	Issues          int                 `json:"issues,omitempty"`
	TaskTitle       string              `json:"task_title,omitempty"`
	TaskType        string              `json:"task_type,omitempty"`
	SafeName        string              `json:"safe_name,omitempty"`
	IDE             string              `json:"ide,omitempty"`
	Model           string              `json:"model,omitempty"`
	ReasoningEffort string              `json:"reasoning_effort,omitempty"`
	AccessMode      string              `json:"access_mode,omitempty"`
	OutLog          string              `json:"out_log,omitempty"`
	ErrLog          string              `json:"err_log,omitempty"`
	Attempt         int                 `json:"attempt,omitempty"`
	MaxAttempts     int                 `json:"max_attempts,omitempty"`
	RetryReason     string              `json:"retry_reason,omitempty"`
	ExitCode        int                 `json:"exit_code,omitempty"`
	ErrorText       string              `json:"error_text,omitempty"`
	Session         SessionViewSnapshot `json:"session,omitempty"`
	Usage           kinds.Usage         `json:"usage,omitempty"`
}

type RunJobState struct {
	Index     int            `json:"index"`
	JobID     string         `json:"job_id"`
	TaskID    string         `json:"task_id,omitempty"`
	Status    string         `json:"status"`
	AgentName string         `json:"agent_name,omitempty"`
	Summary   *RunJobSummary `json:"summary,omitempty"`
	UpdatedAt time.Time      `json:"updated_at"`
}

type RunTranscriptMessage struct {
	Sequence    uint64          `json:"sequence"`
	Stream      string          `json:"stream"`
	Role        string          `json:"role"`
	Content     string          `json:"content"`
	MetadataRaw json.RawMessage `json:"metadata,omitempty"`
	Timestamp   time.Time       `json:"timestamp"`
}

type RunUIMessageRole string

const (
	RunUIMessageRoleSystem    RunUIMessageRole = "system"
	RunUIMessageRoleUser      RunUIMessageRole = "user"
	RunUIMessageRoleAssistant RunUIMessageRole = "assistant"
)

type RunUIMessagePartType string

const (
	RunUIMessagePartText        RunUIMessagePartType = "text"
	RunUIMessagePartReasoning   RunUIMessagePartType = "reasoning"
	RunUIMessagePartDynamicTool RunUIMessagePartType = "dynamic-tool"
	RunUIMessagePartRcEvent     RunUIMessagePartType = "data-rc-event"
	RunUIMessagePartRcBlock     RunUIMessagePartType = "data-rc-block"
)

type RunUIMessagePartState string

const (
	RunUIMessagePartStateStreaming RunUIMessagePartState = "streaming"
	RunUIMessagePartStateDone      RunUIMessagePartState = "done"
)

type RunUIToolPartState string

const (
	RunUIToolPartStateInputStreaming    RunUIToolPartState = "input-streaming"
	RunUIToolPartStateInputAvailable    RunUIToolPartState = "input-available"
	RunUIToolPartStateApprovalRequested RunUIToolPartState = "approval-requested"
	RunUIToolPartStateOutputAvailable   RunUIToolPartState = "output-available"
	RunUIToolPartStateOutputError       RunUIToolPartState = "output-error"
)

type RunUIMessage struct {
	ID          string             `json:"id"`
	Role        RunUIMessageRole   `json:"role"`
	MetadataRaw json.RawMessage    `json:"metadata,omitempty"`
	Parts       []RunUIMessagePart `json:"parts"`
}

type RunUIMessagePart struct {
	Type        RunUIMessagePartType `json:"type"`
	ID          string               `json:"id,omitempty"`
	Text        string               `json:"text,omitempty"`
	State       string               `json:"state,omitempty"`
	ToolName    string               `json:"toolName,omitempty"`
	ToolCallID  string               `json:"toolCallId,omitempty"`
	Title       string               `json:"title,omitempty"`
	Input       json.RawMessage      `json:"input,omitempty"`
	RawInput    json.RawMessage      `json:"rawInput,omitempty"`
	Output      json.RawMessage      `json:"output,omitempty"`
	ErrorText   string               `json:"errorText,omitempty"`
	Data        json.RawMessage      `json:"data,omitempty"`
	Preliminary bool                 `json:"preliminary,omitempty"`
}

type RunTranscript struct {
	RunID             string              `json:"run_id"`
	Messages          []RunUIMessage      `json:"messages"`
	Session           SessionViewSnapshot `json:"session,omitempty"`
	Incomplete        bool                `json:"incomplete,omitempty"`
	IncompleteReasons []string            `json:"incomplete_reasons,omitempty"`
	NextCursor        *StreamCursor       `json:"-"`
}

type RunShutdownState struct {
	Phase       string    `json:"phase,omitempty"`
	Source      string    `json:"source,omitempty"`
	RequestedAt time.Time `json:"requested_at,omitempty"`
	DeadlineAt  time.Time `json:"deadline_at,omitempty"`
}

// RunPendingInput describes a prompt a run is currently waiting on the user to
// answer: an ACP permission request or a free-text skill question.
type RunPendingInput struct {
	PromptID string           `json:"prompt_id"`
	Kind     string           `json:"kind"`
	Text     string           `json:"text,omitempty"`
	Options  []RunInputOption `json:"options,omitempty"`
}

// RunInputOption is one selectable choice offered to the user for a permission
// request; it is empty for a free-text question.
type RunInputOption struct {
	OptionID string `json:"option_id"`
	Label    string `json:"label,omitempty"`
}

// RunInput is a user's answer submitted to a run that is awaiting input.
type RunInput struct {
	PromptID string `json:"prompt_id"`
	OptionID string `json:"option_id,omitempty"`
	Text     string `json:"text,omitempty"`
	Canceled bool   `json:"canceled,omitempty"`
}

type RunSnapshot struct {
	Run               Run                    `json:"run"`
	Jobs              []RunJobState          `json:"jobs,omitempty"`
	Transcript        []RunTranscriptMessage `json:"transcript,omitempty"`
	Usage             kinds.Usage            `json:"usage,omitempty"`
	Shutdown          *RunShutdownState      `json:"shutdown,omitempty"`
	PendingInput      *RunPendingInput       `json:"pending_input,omitempty"`
	Incomplete        bool                   `json:"incomplete,omitempty"`
	IncompleteReasons []string               `json:"incomplete_reasons,omitempty"`
	NextCursor        *StreamCursor          `json:"-"`
}

type RunListQuery struct {
	Workspace string
	Status    string
	Mode      string
	Limit     int
}

type RunEventPageQuery struct {
	After StreamCursor
	Limit int
}

type RunEventPage struct {
	Events     []events.Event
	NextCursor *StreamCursor
	HasMore    bool
}

type SyncResult struct {
	WorkspaceID            string     `json:"workspace_id,omitempty"`
	WorkflowSlug           string     `json:"workflow_slug,omitempty"`
	SyncedAt               *time.Time `json:"synced_at,omitempty"`
	Target                 string     `json:"target,omitempty"`
	WorkflowsScanned       int        `json:"workflows_scanned,omitempty"`
	WorkflowsPruned        int        `json:"workflows_pruned,omitempty"`
	SnapshotsUpserted      int        `json:"snapshots_upserted,omitempty"`
	TaskItemsUpserted      int        `json:"task_items_upserted,omitempty"`
	ReviewRoundsUpserted   int        `json:"review_rounds_upserted,omitempty"`
	ReviewIssuesUpserted   int        `json:"review_issues_upserted,omitempty"`
	CheckpointsUpdated     int        `json:"checkpoints_updated,omitempty"`
	LegacyArtifactsRemoved int        `json:"legacy_artifacts_removed,omitempty"`
	SyncedPaths            []string   `json:"synced_paths,omitempty"`
	PrunedWorkflows        []string   `json:"pruned_workflows,omitempty"`
	Warnings               []string   `json:"warnings,omitempty"`
}

type DaemonStatusResponse struct {
	Daemon DaemonStatus `json:"daemon"`
}

type DaemonHealthResponse struct {
	Health DaemonHealth `json:"health"`
}

type WorkspaceResponse struct {
	Workspace Workspace `json:"workspace"`
}

type WorkspaceListResponse struct {
	Workspaces []Workspace `json:"workspaces"`
}

type TaskWorkflowListResponse struct {
	Workflows []WorkflowSummary `json:"workflows"`
}

type TaskWorkflowResponse struct {
	Workflow WorkflowSummary `json:"workflow"`
}

type TaskItemsResponse struct {
	Items []TaskItem `json:"items"`
}

type ValidationResponse = ValidationSuccess

type ArchiveResponse = ArchiveResult

type ReviewFetchResponse struct {
	Review ReviewSummary `json:"review"`
}

type ReviewSummaryResponse struct {
	Review ReviewSummary `json:"review"`
}

type ReviewRoundResponse struct {
	Round ReviewRound `json:"round"`
}

type ReviewIssuesResponse struct {
	Issues []ReviewIssue `json:"issues"`
}

type RunListResponse struct {
	Runs []Run `json:"runs"`
}

type RunResponse struct {
	Run Run `json:"run"`
}

type RunSnapshotResponse struct {
	Run               Run                    `json:"run"`
	Jobs              []RunJobState          `json:"jobs,omitempty"`
	Transcript        []RunTranscriptMessage `json:"transcript,omitempty"`
	Usage             kinds.Usage            `json:"usage,omitempty"`
	Shutdown          *RunShutdownState      `json:"shutdown,omitempty"`
	Incomplete        bool                   `json:"incomplete,omitempty"`
	IncompleteReasons []string               `json:"incomplete_reasons,omitempty"`
	NextCursor        string                 `json:"next_cursor,omitempty"`
}

type RunTranscriptResponse struct {
	RunID             string              `json:"run_id"`
	Messages          []RunUIMessage      `json:"messages"`
	Session           SessionViewSnapshot `json:"session,omitempty"`
	Incomplete        bool                `json:"incomplete,omitempty"`
	IncompleteReasons []string            `json:"incomplete_reasons,omitempty"`
	NextCursor        string              `json:"next_cursor,omitempty"`
}

type RunEventPageResponse struct {
	Events     []events.Event `json:"events"`
	NextCursor string         `json:"next_cursor,omitempty"`
	HasMore    bool           `json:"has_more"`
}

type SyncResponse = SyncResult

func RunSnapshotResponseFromSnapshot(snapshot RunSnapshot) RunSnapshotResponse {
	return RunSnapshotResponse{
		Run:               snapshot.Run,
		Jobs:              snapshot.Jobs,
		Transcript:        snapshot.Transcript,
		Usage:             snapshot.Usage,
		Shutdown:          snapshot.Shutdown,
		Incomplete:        snapshot.Incomplete,
		IncompleteReasons: append([]string(nil), snapshot.IncompleteReasons...),
		NextCursor:        FormatCursorPointer(snapshot.NextCursor),
	}
}

func (r RunSnapshotResponse) Decode() (RunSnapshot, error) {
	nextCursor, err := ParseCursor(r.NextCursor)
	if err != nil {
		return RunSnapshot{}, fmt.Errorf("decode snapshot cursor: %w", err)
	}

	snapshot := RunSnapshot{
		Run:               r.Run,
		Jobs:              r.Jobs,
		Transcript:        r.Transcript,
		Usage:             r.Usage,
		Shutdown:          r.Shutdown,
		Incomplete:        r.Incomplete,
		IncompleteReasons: append([]string(nil), r.IncompleteReasons...),
	}
	if nextCursor.Sequence > 0 {
		snapshot.NextCursor = &nextCursor
	}
	return snapshot, nil
}

func RunTranscriptResponseFromTranscript(transcript RunTranscript) RunTranscriptResponse {
	messages := transcript.Messages
	if messages == nil {
		messages = []RunUIMessage{}
	}
	return RunTranscriptResponse{
		RunID:             transcript.RunID,
		Messages:          messages,
		Session:           transcript.Session,
		Incomplete:        transcript.Incomplete,
		IncompleteReasons: append([]string(nil), transcript.IncompleteReasons...),
		NextCursor:        FormatCursorPointer(transcript.NextCursor),
	}
}

func (r RunTranscriptResponse) Decode() (RunTranscript, error) {
	nextCursor, err := ParseCursor(r.NextCursor)
	if err != nil {
		return RunTranscript{}, fmt.Errorf("decode transcript cursor: %w", err)
	}
	transcript := RunTranscript{
		RunID:             r.RunID,
		Messages:          append([]RunUIMessage(nil), r.Messages...),
		Session:           r.Session,
		Incomplete:        r.Incomplete,
		IncompleteReasons: append([]string(nil), r.IncompleteReasons...),
	}
	if nextCursor.Sequence > 0 {
		transcript.NextCursor = &nextCursor
	}
	return transcript, nil
}

func RunEventPageResponseFromPage(page RunEventPage) RunEventPageResponse {
	return RunEventPageResponse{
		Events:     page.Events,
		NextCursor: FormatCursorPointer(page.NextCursor),
		HasMore:    page.HasMore,
	}
}

func (r RunEventPageResponse) Decode() (RunEventPage, error) {
	nextCursor, err := ParseCursor(r.NextCursor)
	if err != nil {
		return RunEventPage{}, fmt.Errorf("decode events cursor: %w", err)
	}
	if r.HasMore && nextCursor.Sequence == 0 {
		return RunEventPage{}, fmt.Errorf("decode events cursor: missing next_cursor when has_more=true")
	}

	page := RunEventPage{
		Events:  r.Events,
		HasMore: r.HasMore,
	}
	if nextCursor.Sequence > 0 {
		page.NextCursor = &nextCursor
	}
	return page, nil
}

// ConfigDocument is the JSON-facing representation of ProjectConfig.
// All fields are pointer-optional so JSON null / absent means "unset".
// Round-trip: JSON → ProjectConfig → TOML (write) and TOML → ProjectConfig → JSON (read).
type ConfigDocument struct {
	Defaults     *ConfigDefaults     `json:"defaults,omitempty"`
	Tasks        *ConfigTasks        `json:"tasks,omitempty"`
	FixReviews   *ConfigFixReviews   `json:"fix_reviews,omitempty"`
	FetchReviews *ConfigFetchReviews `json:"fetch_reviews,omitempty"`
	WatchReviews *ConfigWatchReviews `json:"watch_reviews,omitempty"`
	Exec         *ConfigExec         `json:"exec,omitempty"`
	Runs         *ConfigRuns         `json:"runs,omitempty"`
	Sound        *ConfigSound        `json:"sound,omitempty"`
}

// ConfigDefaults mirrors workspace.DefaultsConfig.
type ConfigDefaults struct {
	IDE                    *string   `json:"ide,omitempty"`
	Model                  *string   `json:"model,omitempty"`
	OutputFormat           *string   `json:"output_format,omitempty"`
	ReasoningEffort        *string   `json:"reasoning_effort,omitempty"`
	AccessMode             *string   `json:"access_mode,omitempty"`
	Timeout                *string   `json:"timeout,omitempty"`
	TailLines              *int      `json:"tail_lines,omitempty"`
	AddDirs                *[]string `json:"add_dirs,omitempty"`
	AutoCommit             *bool     `json:"auto_commit,omitempty"`
	MaxRetries             *int      `json:"max_retries,omitempty"`
	RetryBackoffMultiplier *float64  `json:"retry_backoff_multiplier,omitempty"`
}

// ConfigTasks mirrors workspace.TasksConfig.
type ConfigTasks struct {
	Types *[]string      `json:"types,omitempty"`
	Run   *ConfigTaskRun `json:"run,omitempty"`
}

// ConfigTaskRuntimeRule mirrors model.TaskRuntimeRule.
type ConfigTaskRuntimeRule struct {
	ID              *string `json:"id,omitempty"`
	Type            *string `json:"type,omitempty"`
	IDE             *string `json:"ide,omitempty"`
	Model           *string `json:"model,omitempty"`
	ReasoningEffort *string `json:"reasoning_effort,omitempty"`
}

// ConfigTaskRun mirrors workspace.TaskRunConfig.
type ConfigTaskRun struct {
	IncludeCompleted *bool                    `json:"include_completed,omitempty"`
	OutputFormat     *string                  `json:"output_format,omitempty"`
	TUI              *bool                    `json:"tui,omitempty"`
	TaskRuntimeRules *[]ConfigTaskRuntimeRule `json:"task_runtime_rules,omitempty"`
}

// ConfigFixReviews mirrors workspace.FixReviewsConfig.
type ConfigFixReviews struct {
	Concurrent      *int    `json:"concurrent,omitempty"`
	BatchSize       *int    `json:"batch_size,omitempty"`
	IncludeResolved *bool   `json:"include_resolved,omitempty"`
	OutputFormat    *string `json:"output_format,omitempty"`
	TUI             *bool   `json:"tui,omitempty"`
}

// ConfigFetchReviews mirrors workspace.FetchReviewsConfig.
type ConfigFetchReviews struct {
	Provider *string `json:"provider,omitempty"`
	Nitpicks *bool   `json:"nitpicks,omitempty"`
}

// ConfigWatchReviews mirrors workspace.WatchReviewsConfig.
type ConfigWatchReviews struct {
	MaxRounds     *int    `json:"max_rounds,omitempty"`
	PollInterval  *string `json:"poll_interval,omitempty"`
	ReviewTimeout *string `json:"review_timeout,omitempty"`
	QuietPeriod   *string `json:"quiet_period,omitempty"`
	AutoPush      *bool   `json:"auto_push,omitempty"`
	UntilClean    *bool   `json:"until_clean,omitempty"`
	PushRemote    *string `json:"push_remote,omitempty"`
	PushBranch    *string `json:"push_branch,omitempty"`
}

// ConfigExec mirrors workspace.ExecConfig.
type ConfigExec struct {
	IDE                    *string   `json:"ide,omitempty"`
	Model                  *string   `json:"model,omitempty"`
	OutputFormat           *string   `json:"output_format,omitempty"`
	ReasoningEffort        *string   `json:"reasoning_effort,omitempty"`
	AccessMode             *string   `json:"access_mode,omitempty"`
	Timeout                *string   `json:"timeout,omitempty"`
	TailLines              *int      `json:"tail_lines,omitempty"`
	AddDirs                *[]string `json:"add_dirs,omitempty"`
	AutoCommit             *bool     `json:"auto_commit,omitempty"`
	MaxRetries             *int      `json:"max_retries,omitempty"`
	RetryBackoffMultiplier *float64  `json:"retry_backoff_multiplier,omitempty"`
	Verbose                *bool     `json:"verbose,omitempty"`
	TUI                    *bool     `json:"tui,omitempty"`
	Persist                *bool     `json:"persist,omitempty"`
}

// ConfigRuns mirrors workspace.RunsConfig.
type ConfigRuns struct {
	DefaultAttachMode    *string `json:"default_attach_mode,omitempty"`
	KeepTerminalDays     *int    `json:"keep_terminal_days,omitempty"`
	KeepMax              *int    `json:"keep_max,omitempty"`
	ShutdownDrainTimeout *string `json:"shutdown_drain_timeout,omitempty"`
}

// ConfigSound mirrors workspace.SoundConfig.
type ConfigSound struct {
	Enabled     *bool   `json:"enabled,omitempty"`
	OnCompleted *string `json:"on_completed,omitempty"`
	OnFailed    *string `json:"on_failed,omitempty"`
}

// ConfigDocumentResponse wraps ConfigDocument for API responses.
type ConfigDocumentResponse struct {
	Config ConfigDocument `json:"config"`
}

// ExtensionItem describes one discovered extension for the catalog API.
type ExtensionItem struct {
	Name        string `json:"name"`
	Source      string `json:"source"`
	Enabled     bool   `json:"enabled"`
	Description string `json:"description,omitempty"`
	Version     string `json:"version,omitempty"`
}

// ExtensionListResponse wraps the extension catalog for API responses.
type ExtensionListResponse struct {
	Extensions    []ExtensionItem `json:"extensions"`
	FailureCount  int             `json:"failure_count,omitempty"`
	OverrideCount int             `json:"override_count,omitempty"`
}

// AgentItem describes one resolved reusable agent for the catalog API.
type AgentItem struct {
	Name        string   `json:"name"`
	Scope       string   `json:"scope"`
	Description string   `json:"description,omitempty"`
	Warnings    []string `json:"warnings,omitempty"`
}

// AgentListResponse wraps the agent catalog for API responses.
type AgentListResponse struct {
	Agents []AgentItem `json:"agents"`
}

// SetupAgent describes one agent destination available for skill installation.
type SetupAgent struct {
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
	Detected    bool   `json:"detected"`
}

// SetupSkill describes one bundled rc skill available for installation.
type SetupSkill struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

// SetupOptionsResponse lists the agents and bundled skills a project can
// install, plus whether the project already has the rc skills set up.
type SetupOptionsResponse struct {
	Agents     []SetupAgent `json:"agents"`
	Skills     []SetupSkill `json:"skills"`
	Configured bool         `json:"configured"`
}

// SetupInstallRequest selects which bundled skills install into which agents.
type SetupInstallRequest struct {
	Agents []string `json:"agents"`
	Skills []string `json:"skills"`
}

// SetupInstalledItem reports one successful skill/agent installation mapping.
type SetupInstalledItem struct {
	Skill string `json:"skill"`
	Agent string `json:"agent"`
	Path  string `json:"path"`
}

// SetupFailedItem reports one failed skill/agent installation mapping.
type SetupFailedItem struct {
	Skill string `json:"skill"`
	Agent string `json:"agent"`
	Error string `json:"error"`
}

// SetupInstallResponse summarizes one project-scoped setup installation run.
type SetupInstallResponse struct {
	Installed []SetupInstalledItem `json:"installed"`
	Failed    []SetupFailedItem    `json:"failed"`
}
