package messages

// Message represents any message in the agent conversation
type Message interface {
	GetRole() string
	GetContent() []ContentBlock
	GetTimestamp() int64
}

type Role string

const (
	RoleUser       Role = "user"
	RoleAssistant  Role = "assistant"
	RoleToolResult Role = "toolResult"
	RoleSystem     Role = "system"
)

type ContentBlock interface {
	GetType() string
}

// Text content block
type TextBlock struct {
	Text string `json:"text"`
}

func (t *TextBlock) GetType() string { return "text" }

// Image content block
type ImageBlock struct {
	URL    string `json:"url,omitempty"`
	Base64 string `json:"base64,omitempty"`
	Alt    string `json:"alt,omitempty"`
}

func (i *ImageBlock) GetType() string { return "image" }

// ToolCall content block
type ToolCallBlock struct {
	ID        string         `json:"id"`
	Name      string         `json:"name"`
	Arguments map[string]any `json:"arguments"`
}

func (t *ToolCallBlock) GetType() string { return "toolCall" }

// ToolResult content block
type ToolResultBlock struct {
	ToolCallID string         `json:"toolCallId"`
	ToolName   string         `json:"toolName"`
	Content    []ContentBlock `json:"content"`
	IsError    bool           `json:"isError,omitempty"`
}

func (t *ToolResultBlock) GetType() string { return "toolResult" }

// UserMessage is a message from the user
type UserMessage struct {
	Role      Role           `json:"role"`
	Content   []ContentBlock `json:"content"`
	Timestamp int64          `json:"timestamp"`
}

func (u *UserMessage) GetRole() string            { return string(u.Role) }
func (u *UserMessage) GetContent() []ContentBlock { return u.Content }
func (u *UserMessage) GetTimestamp() int64        { return u.Timestamp }

// AssistantMessage is a message from the assistant (LLM)
type AssistantMessage struct {
	Role       Role           `json:"role"`
	Content    []ContentBlock `json:"content"`
	StopReason string         `json:"stopReason,omitempty"`
	Timestamp  int64          `json:"timestamp"`
}

func (a *AssistantMessage) GetRole() string            { return string(a.Role) }
func (a *AssistantMessage) GetContent() []ContentBlock { return a.Content }
func (a *AssistantMessage) GetTimestamp() int64        { return a.Timestamp }

// ToolResultMessage is a message containing tool execution results
type ToolResultMessage struct {
	Role       Role           `json:"role"`
	ToolCallID string         `json:"toolCallId"`
	ToolName   string         `json:"toolName"`
	Content    []ContentBlock `json:"content"`
	IsError    bool           `json:"isError,omitempty"`
	Details    any            `json:"details,omitempty"`
	Usage      *Usage         `json:"usage,omitempty"`
	Timestamp  int64          `json:"timestamp"`
}

func (t *ToolResultMessage) GetRole() string            { return string(t.Role) }
func (t *ToolResultMessage) GetContent() []ContentBlock { return t.Content }
func (t *ToolResultMessage) GetTimestamp() int64        { return t.Timestamp }

// SystemMessage is a system prompt message
type SystemMessage struct {
	Role      Role           `json:"role"`
	Content   []ContentBlock `json:"content"`
	Timestamp int64          `json:"timestamp"`
}

func (s *SystemMessage) GetRole() string            { return string(s.Role) }
func (s *SystemMessage) GetContent() []ContentBlock { return s.Content }
func (s *SystemMessage) GetTimestamp() int64        { return s.Timestamp }

// Usage tracks token usage
type Usage struct {
	InputTokens  int `json:"inputTokens,omitempty"`
	OutputTokens int `json:"outputTokens,omitempty"`
	TotalTokens  int `json:"totalTokens,omitempty"`
}
