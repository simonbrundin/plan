package metrics

import (
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

// Metrics holds all application metrics
type Metrics struct {
	// HTTP metrics
	httpRequestsTotal    map[string]*uint64 // path -> count
	httpRequestsInFlight uint64
	httpRequestDuration  *Histogram

	// Agent metrics
	agentTurnsTotal     uint64
	agentToolCallsTotal uint64
	agentErrorsTotal    uint64

	// Tool metrics
	toolExecutionsTotal map[string]*uint64 // tool_name -> count
	toolDuration        map[string]*Histogram

	// LLM metrics
	llmRequestsTotal   map[string]*uint64 // provider -> count
	llmTokensTotal     map[string]*uint64 // provider -> tokens
	llmErrorsTotal     map[string]*uint64 // provider -> errors
	llmRequestDuration map[string]*Histogram

	// mu for map access
	mu sync.RWMutex
}

// Histogram tracks a distribution of values
type Histogram struct {
	mu      sync.Mutex
	count   uint64
	sum     float64
	buckets map[float64]uint64 // bucket -> count
}

// DefaultBuckets returns common histogram buckets
func DefaultBuckets() []float64 {
	return []float64{0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10}
}

// NewHistogram creates a new histogram with buckets
func NewHistogram(buckets []float64) *Histogram {
	h := &Histogram{
		buckets: make(map[float64]uint64),
	}
	for _, b := range buckets {
		h.buckets[b] = 0
	}
	return h
}

// Observe adds an observation to the histogram
func (h *Histogram) Observe(v float64) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.count++
	h.sum += v
	for bucket := range h.buckets {
		if v <= bucket {
			h.buckets[bucket]++
		}
	}
}

// Summary returns histogram summary
func (h *Histogram) Summary() (count uint64, sum float64) {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.count, h.sum
}

// Buckets returns bucket counts
func (h *Histogram) Buckets() map[float64]uint64 {
	h.mu.Lock()
	defer h.mu.Unlock()
	result := make(map[float64]uint64)
	for k, v := range h.buckets {
		result[k] = v
	}
	return result
}

// Global metrics instance
var global *Metrics
var initOnce sync.Once

// Get returns the global metrics instance
func Get() *Metrics {
	initOnce.Do(func() {
		global = New()
	})
	return global
}

// New creates a new Metrics instance
func New() *Metrics {
	return &Metrics{
		httpRequestsTotal:    make(map[string]*uint64),
		httpRequestsInFlight: 0,
		httpRequestDuration:  NewHistogram(DefaultBuckets()),
		toolExecutionsTotal:  make(map[string]*uint64),
		toolDuration:         make(map[string]*Histogram),
		llmRequestsTotal:     make(map[string]*uint64),
		llmTokensTotal:       make(map[string]*uint64),
		llmErrorsTotal:       make(map[string]*uint64),
		llmRequestDuration:   make(map[string]*Histogram),
	}
}

// HTTP metrics

// HTTPRequestStart is called when an HTTP request starts
func (m *Metrics) HTTPRequestStart(path string) {
	atomic.AddUint64(&m.httpRequestsInFlight, 1)
}

// HTTPRequestEnd is called when an HTTP request ends
func (m *Metrics) HTTPRequestEnd(path string, duration time.Duration) {
	atomic.AddUint64(&m.httpRequestsInFlight, ^uint64(0))

	m.mu.Lock()
	if m.httpRequestsTotal[path] == nil {
		m.httpRequestsTotal[path] = new(uint64)
	}
	m.mu.Unlock()

	atomic.AddUint64(m.httpRequestsTotal[path], 1)
	m.httpRequestDuration.Observe(duration.Seconds())
}

// HTTPRequestsInFlight returns the number of requests being processed
func (m *Metrics) HTTPRequestsInFlight() uint64 {
	return atomic.LoadUint64(&m.httpRequestsInFlight)
}

// Agent metrics

// AgentTurn increments the agent turn counter
func (m *Metrics) AgentTurn() {
	atomic.AddUint64(&m.agentTurnsTotal, 1)
}

// AgentToolCall increments the tool call counter
func (m *Metrics) AgentToolCall() {
	atomic.AddUint64(&m.agentToolCallsTotal, 1)
}

// AgentError increments the agent error counter
func (m *Metrics) AgentError() {
	atomic.AddUint64(&m.agentErrorsTotal, 1)
}

// Tool metrics

// ToolExecution records a tool execution
func (m *Metrics) ToolExecution(toolName string, duration time.Duration) {
	m.mu.Lock()
	if m.toolExecutionsTotal[toolName] == nil {
		m.toolExecutionsTotal[toolName] = new(uint64)
		m.toolDuration[toolName] = NewHistogram([]float64{0.01, 0.05, 0.1, 0.5, 1, 5, 10, 30, 60})
	}
	m.mu.Unlock()

	atomic.AddUint64(m.toolExecutionsTotal[toolName], 1)
	m.toolDuration[toolName].Observe(duration.Seconds())
}

// LLM metrics

// LLMRequest records an LLM request
func (m *Metrics) LLMRequest(provider, model string, duration time.Duration, inputTokens, outputTokens int, err bool) {
	m.mu.Lock()
	if m.llmRequestsTotal[provider] == nil {
		m.llmRequestsTotal[provider] = new(uint64)
		m.llmTokensTotal[provider] = new(uint64)
		m.llmErrorsTotal[provider] = new(uint64)
		m.llmRequestDuration[provider] = NewHistogram([]float64{0.1, 0.25, 0.5, 1, 2, 5, 10, 30})
	}
	m.mu.Unlock()

	atomic.AddUint64(m.llmRequestsTotal[provider], 1)
	atomic.AddUint64(m.llmTokensTotal[provider], uint64(inputTokens+outputTokens))
	if err {
		atomic.AddUint64(m.llmErrorsTotal[provider], 1)
	}
	m.llmRequestDuration[provider].Observe(duration.Seconds())
}

// PrometheusOutput generates Prometheus-formatted metrics
func (m *Metrics) PrometheusOutput() string {
	var output string

	// HTTP metrics
	output += "# HELP http_requests_total Total HTTP requests\n"
	output += "# TYPE http_requests_total counter\n"
	m.mu.RLock()
	for path, count := range m.httpRequestsTotal {
		output += "http_requests_total{path=\"" + path + "\"} " + strconv.FormatUint(*count, 10) + "\n"
	}
	m.mu.RUnlock()

	output += "# HELP http_requests_in_flight HTTP requests currently being processed\n"
	output += "# TYPE http_requests_in_flight gauge\n"
	output += "http_requests_in_flight " + strconv.FormatUint(m.HTTPRequestsInFlight(), 10) + "\n"

	output += "# HELP http_request_duration_seconds HTTP request duration\n"
	output += "# TYPE http_request_duration_seconds histogram\n"
	count, sum := m.httpRequestDuration.Summary()
	output += "http_request_duration_seconds_count " + strconv.FormatUint(count, 10) + "\n"
	output += "http_request_duration_seconds_sum " + strconv.FormatFloat(sum, 'f', 6, 64) + "\n"
	for bucket, bc := range m.httpRequestDuration.Buckets() {
		output += "http_request_duration_seconds_bucket{le=\"" + strconv.FormatFloat(bucket, 'f', 6, 64) + "\"} " + strconv.FormatUint(bc, 10) + "\n"
	}

	// Agent metrics
	output += "# HELP agent_turns_total Total agent turns\n"
	output += "# TYPE agent_turns_total counter\n"
	output += "agent_turns_total " + strconv.FormatUint(atomic.LoadUint64(&m.agentTurnsTotal), 10) + "\n"

	output += "# HELP agent_tool_calls_total Total tool calls\n"
	output += "# TYPE agent_tool_calls_total counter\n"
	output += "agent_tool_calls_total " + strconv.FormatUint(atomic.LoadUint64(&m.agentToolCallsTotal), 10) + "\n"

	output += "# HELP agent_errors_total Total agent errors\n"
	output += "# TYPE agent_errors_total counter\n"
	output += "agent_errors_total " + strconv.FormatUint(atomic.LoadUint64(&m.agentErrorsTotal), 10) + "\n"

	// Tool metrics
	output += "# HELP tool_executions_total Total tool executions\n"
	output += "# TYPE tool_executions_total counter\n"
	m.mu.RLock()
	for tool, count := range m.toolExecutionsTotal {
		output += "tool_executions_total{tool=\"" + tool + "\"} " + strconv.FormatUint(*count, 10) + "\n"
	}
	m.mu.RUnlock()

	// LLM metrics
	output += "# HELP llm_requests_total Total LLM requests\n"
	output += "# TYPE llm_requests_total counter\n"
	output += "# HELP llm_tokens_total Total LLM tokens\n"
	output += "# TYPE llm_tokens_total counter\n"
	output += "# HELP llm_errors_total Total LLM errors\n"
	output += "# TYPE llm_errors_total counter\n"
	output += "# HELP llm_request_duration_seconds LLM request duration\n"
	output += "# TYPE llm_request_duration_seconds histogram\n"
	m.mu.RLock()
	for provider, count := range m.llmRequestsTotal {
		output += "llm_requests_total{provider=\"" + provider + "\"} " + strconv.FormatUint(*count, 10) + "\n"
		output += "llm_tokens_total{provider=\"" + provider + "\"} " + strconv.FormatUint(*m.llmTokensTotal[provider], 10) + "\n"
		output += "llm_errors_total{provider=\"" + provider + "\"} " + strconv.FormatUint(*m.llmErrorsTotal[provider], 10) + "\n"
		if d, ok := m.llmRequestDuration[provider]; ok {
			c, s := d.Summary()
			output += "llm_request_duration_seconds_count{provider=\"" + provider + "\"} " + strconv.FormatUint(c, 10) + "\n"
			output += "llm_request_duration_seconds_sum{provider=\"" + provider + "\"} " + strconv.FormatFloat(s, 'f', 6, 64) + "\n"
		}
	}
	m.mu.RUnlock()

	return output
}
