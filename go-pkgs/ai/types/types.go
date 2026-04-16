package types

// Provider represents an AI provider type
type Provider string

const (
	ProviderOpenAI    Provider = "openai"
	ProviderDeepSeek  Provider = "deepseek"
	ProviderAnthropic Provider = "anthropic"
	ProviderMoonshot  Provider = "moonshot"
)

// Scenario represents different AI usage scenarios
type Scenario string

const (
	ScenarioDefault    Scenario = "default"    // Default scenario for general use
	ScenarioFast       Scenario = "fast"       // Fast response, simpler tasks
	ScenarioSuggestion Scenario = "suggestion" // Quick suggestions like alias generation
	ScenarioReasoning  Scenario = "reasoning"  // Complex reasoning tasks
)

// ChunkType represents the type of a streamed chunk
type ChunkType string

const (
	ChunkTypeThinking ChunkType = "thinking"
	ChunkTypeContent  ChunkType = "content"
	ChunkTypeDone     ChunkType = "done"
	ChunkTypeError    ChunkType = "error"
)

// Message represents a chat message
type Message struct {
	Role    string `json:"role"`    // "user", "assistant", "system"
	Content string `json:"content"` // The message content
}

// ChatRequest represents a chat completion request
type ChatRequest struct {
	Messages []Message   `json:"messages"`
	Context  ChatContext `json:"context,omitempty"`
	// Provider allows explicit provider selection (optional)
	Provider string `json:"provider,omitempty"`
	// Model allows explicit model selection (optional)
	Model string `json:"model,omitempty"`
}

// ChatContext provides additional context for the chat
type ChatContext struct {
	Type    string   `json:"type,omitempty"`     // e.g., "pick_something"
	TaskIDs []string `json:"task_ids,omitempty"` // Optional task IDs for context
}

// StreamChunk represents a chunk of streamed AI response
type StreamChunk struct {
	Type       ChunkType   `json:"type"`                 // "thinking", "content", "done", "error"
	Content    string      `json:"content,omitempty"`    // The actual content
	Error      string      `json:"error,omitempty"`      // Error message if type is "error"
	TokenUsage *TokenUsage `json:"tokenUsage,omitempty"` // Token usage (sent with done chunk)
}

// StreamCallback is called for each chunk of the streamed response
type StreamCallback func(chunk StreamChunk) error

// Tool represents a function tool the AI can call
type Tool struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Parameters  map[string]any `json:"parameters"`
}

// ToolCall represents a tool call made by the AI
type ToolCall struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

// ToolResult represents the result of a tool call
type ToolResult struct {
	ToolCallID string `json:"tool_call_id"`
	Name       string `json:"name"`
	Content    string `json:"content"`
}

// ChatOptions holds optional parameters for chat completion
type ChatOptions struct {
	Tools []Tool `json:"tools,omitempty"`
}

// ChatResponse represents a non-streaming chat completion response
type ChatResponse struct {
	Content   string     `json:"content,omitempty"`
	ToolCalls []ToolCall `json:"tool_calls,omitempty"`
}

// ToolMessage represents a message with tool call or tool result context
type ToolMessage struct {
	Message
	ToolCalls  []ToolCall  `json:"tool_calls,omitempty"`
	ToolResult *ToolResult `json:"tool_result,omitempty"`
}

// Config holds AI provider configuration
type Config struct {
	Provider     Provider `json:"provider"`
	APIKey       string   `json:"api_key"`
	BaseURL      string   `json:"base_url,omitempty"`
	Model        string   `json:"model,omitempty"`
	MaxTokens    int      `json:"max_tokens,omitempty"`
	ThinkingMode bool     `json:"thinking_mode,omitempty"`
}

// PickSomethingRequest represents a request to get task suggestions
type PickSomethingRequest struct {
	// No parameters needed - backend will fetch recent tasks automatically
}

// TaskSuggestion represents a suggested task with explanation
type TaskSuggestion struct {
	TaskID       string `json:"taskId"`
	Title        string `json:"title"`
	IsPicked     bool   `json:"isPicked"`               // Whether this task is recommended
	Reason       string `json:"reason"`                 // Why this task is/isn't recommended
	Priority     int    `json:"priority"`               // 1-5, higher is more important
	Urgency      string `json:"urgency"`                // "high", "medium", "low"
	TimeEstimate string `json:"timeEstimate,omitempty"` // e.g., "30 min", "1 hour"
}

// PickSomethingResponse represents the AI's task suggestions
type PickSomethingResponse struct {
	Summary     string           `json:"summary"`           // Overall summary/advice
	Suggestions []TaskSuggestion `json:"suggestions"`       // List of task suggestions
	TopPick     *TaskSuggestion  `json:"topPick,omitempty"` // The most recommended task
}

// PickChunkType represents the type of a pick something stream chunk
type PickChunkType string

const (
	PickChunkTypeStatus   PickChunkType = "status"   // Status update (e.g., "Fetching tasks...")
	PickChunkTypeThinking PickChunkType = "thinking" // AI thinking process
	PickChunkTypeContent  PickChunkType = "content"  // AI response content
	PickChunkTypeResult   PickChunkType = "result"   // Final parsed result
	PickChunkTypeDone     PickChunkType = "done"     // Stream complete
	PickChunkTypeError    PickChunkType = "error"    // Error occurred
)

// TokenUsage represents token usage statistics
type TokenUsage struct {
	PromptTokens     int `json:"promptTokens,omitempty"`
	CompletionTokens int `json:"completionTokens,omitempty"`
	TotalTokens      int `json:"totalTokens,omitempty"`
}

// PickSomethingStreamChunk represents a chunk in the pick something stream
type PickSomethingStreamChunk struct {
	Type       PickChunkType          `json:"type"`
	Status     string                 `json:"status,omitempty"`     // For status type
	Content    string                 `json:"content,omitempty"`    // For thinking/content type
	Result     *PickSomethingResponse `json:"result,omitempty"`     // For result type
	Error      string                 `json:"error,omitempty"`      // For error type
	TokenUsage *TokenUsage            `json:"tokenUsage,omitempty"` // Token usage stats
	Cost       float64                `json:"cost,omitempty"`       // Cost in USD
}

// PickSomethingStreamCallback is called for each chunk
type PickSomethingStreamCallback func(chunk PickSomethingStreamChunk) error
