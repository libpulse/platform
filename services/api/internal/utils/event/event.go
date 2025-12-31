package event

import (
	"time"

	"github.com/google/uuid"
	"github.com/libpulse/platform/services/api/internal/supabase"
)

// SDKInfo represents SDK metadata for internal processing
type SDKInfo struct {
	Name     *string
	Version  *string
	OS       *string
	Arch     *string
	Language *string
	Runtime  *string
}

// IngestEventData represents the event data for internal processing
type IngestEventData struct {
	EventType string
	EventName string
	Version   *string
	EventTS   string
	EventID   *string
	Props     map[string]interface{}
	SDK       *SDKInfo
}

// GenerateEventID generates or returns the event ID
func GenerateEventID(eventID *string) string {
	if eventID != nil && *eventID != "" {
		return *eventID
	}
	return uuid.New().String()
}

// ConvertEventToParams converts event data to database params
func ConvertEventToParams(projectID, eventID string, event *IngestEventData) (*supabase.InsertEventParams, error) {
	// Parse timestamp
	eventTS, err := time.Parse(time.RFC3339, event.EventTS)
	if err != nil {
		return nil, err
	}

	// Extract fields from props
	version := "unknown"
	if event.Version != nil {
		version = *event.Version
	}

	// Extract SDK info
	sdkName := "unknown"
	sdkVersion := "unknown"
	var sdkLanguage, sdkRuntime *string

	if event.SDK != nil {
		if event.SDK.Name != nil {
			sdkName = *event.SDK.Name
		}
		if event.SDK.Version != nil {
			sdkVersion = *event.SDK.Version
		}
		sdkLanguage = event.SDK.Language
		sdkRuntime = event.SDK.Runtime
	}

	// Extract props values
	var severity, code, message, stack, op, variant, surface *string
	var userID, sessionID, traceID *string
	var durationMS *int
	var argsSig *string
	var argsCount *int

	// Helper to get string from props
	getString := func(key string) *string {
		if v, ok := event.Props[key]; ok {
			if s, ok := v.(string); ok && s != "" {
				return &s
			}
		}
		return nil
	}

	// Helper to get int from props
	getInt := func(key string) *int {
		if v, ok := event.Props[key]; ok {
			if f, ok := v.(float64); ok {
				i := int(f)
				return &i
			}
		}
		return nil
	}

	severity = getString("severity")
	code = getString("code")
	message = getString("message")
	stack = getString("stack")
	op = getString("op")
	variant = getString("variant")
	surface = getString("surface")
	userID = getString("user_id")
	sessionID = getString("session_id")
	traceID = getString("trace_id")
	argsSig = getString("args_sig")
	argsCount = getInt("args_count")
	durationMS = getInt("duration_ms")

	// Set default op to event_name if not provided
	if op == nil {
		op = &event.EventName
	}

	// Set default user_id_h to "anonymous" if not provided
	userIDH := "anonymous"
	if userID != nil {
		userIDH = *userID
	}

	params := &supabase.InsertEventParams{
		ProjectID:   projectID,
		EventID:     eventID,
		EventType:   event.EventType,
		EventName:   event.EventName,
		EventTS:     eventTS,
		Op:          *op,
		Variant:     variant,
		Surface:     surface,
		Version:     version,
		ArgsSig:     argsSig,
		ArgsCount:   argsCount,
		Severity:    severity,
		Code:        code,
		Message:     message,
		Stack:       stack,
		DurationMS:  durationMS,
		UserIDH:     userIDH,
		SessionID:   sessionID,
		TraceID:     traceID,
		SDKName:     sdkName,
		SDKVersion:  sdkVersion,
		SDKLanguage: sdkLanguage,
		SDKRuntime:  sdkRuntime,
	}

	return params, nil
}
