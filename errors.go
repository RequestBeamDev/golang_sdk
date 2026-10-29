package requestbeam

// SDKError describes a locally detected validation, transport, or protocol error.
// Code is stable across languages; Message is localized.
type SDKError struct {
	Code    string
	Message string
	Cause   error
}

func (e *SDKError) Error() string { return e.Message }
func (e *SDKError) Unwrap() error { return e.Cause }

// APIError retains the server's original error and diagnostic metadata.
type APIError struct {
	StatusCode int
	Code       any
	Type       string
	RequestID  string
	RetryAfter string
	Body       map[string]any
	Message    string
}

func (e *APIError) Error() string { return e.Message }

// JobFailedError indicates a FAILED or CANCELLED remote job.
type JobFailedError struct {
	Message string
	Job     map[string]any
}

func (e *JobFailedError) Error() string { return e.Message }

// PollingTimeoutError means local waiting expired, not that the remote job stopped.
type PollingTimeoutError struct {
	Message string
	JobID   string
	Cause   error
}

func (e *PollingTimeoutError) Error() string { return e.Message }
func (e *PollingTimeoutError) Unwrap() error { return e.Cause }
