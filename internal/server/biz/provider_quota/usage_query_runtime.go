package provider_quota

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"regexp"
	"strings"
	"time"

	"github.com/dop251/goja"
)

const (
	maxUsageQueryScriptBytes = 64 << 10
	usageQueryScriptTimeout  = 500 * time.Millisecond
)

var usageQueryProgressWindowID = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9_-]*$`)

// UsageQueryScriptRuntime is the replaceable boundary between quota business
// logic and the embedded JavaScript engine.
type UsageQueryScriptRuntime interface {
	ParseRequest(ctx context.Context, source string, variables map[string]string) (UsageQueryRequest, error)
	Extract(ctx context.Context, source string, response UsageQueryHTTPResponse, scriptContext UsageQueryScriptContext) (UsageQueryResult, error)
}

type UsageQueryRequest struct {
	URL     string            `json:"url"`
	Method  string            `json:"method"`
	Headers map[string]string `json:"headers"`
	Body    any               `json:"body,omitempty"`
}

// UsageQueryHTTPResponse is the HTTP response passed to new extractor scripts.
// Header names are lower-cased and multiple values are joined with ", ".
type UsageQueryHTTPResponse struct {
	Status  int               `json:"status"`
	Headers map[string]string `json:"headers"`
	Body    any               `json:"body"`
}

// UsageQueryScriptContext is host-provided, read-only data for extractors.
// Now always uses an ISO 8601 offset such as 2026-08-18T07:36:46+08:00.
type UsageQueryScriptContext struct {
	Now string `json:"now"`
}

type UsageQueryResult struct {
	// Balance is optional because not every provider represents its available
	// quota as money. When present, Remaining is required.
	Balance  *UsageQueryBalance  `json:"balance,omitempty"`
	Text     string              `json:"text,omitempty"`
	Tags     []string            `json:"tags,omitempty"`
	Progress *UsageQueryProgress `json:"progress,omitempty"`

	// Deprecated: return these fields directly to preserve compatibility with
	// existing usage-query scripts. New scripts should return text instead.
	IsValid        *bool    `json:"isValid,omitempty"`
	InvalidMessage string   `json:"invalidMessage,omitempty"`
	Remaining      *float64 `json:"remaining,omitempty"`
	Unit           string   `json:"unit,omitempty"`
	PlanName       string   `json:"planName,omitempty"`
	Total          *float64 `json:"total,omitempty"`
	Used           *float64 `json:"used,omitempty"`
	Extra          string   `json:"extra,omitempty"`
}

type UsageQueryBalance struct {
	Remaining float64 `json:"remaining"`
	Unit      string  `json:"unit,omitempty"`
}

// UsageQueryProgress groups independently rendered quota windows. A script can
// return any number of windows and each window may omit reset timestamps.
type UsageQueryProgress struct {
	Windows []UsageQueryProgressWindow `json:"windows,omitempty"`
}

type UsageQueryProgressWindow struct {
	ID               string   `json:"id"`
	DurationSeconds  *int     `json:"durationSeconds,omitempty"`
	RemainingPercent *float64 `json:"remainingPercent,omitempty"`
	ResetAt          string   `json:"resetAt,omitempty"`
}

type GojaUsageQueryRuntime struct{}

func NewGojaUsageQueryRuntime() *GojaUsageQueryRuntime {
	return &GojaUsageQueryRuntime{}
}

func (r *GojaUsageQueryRuntime) ParseRequest(
	ctx context.Context,
	source string,
	variables map[string]string,
) (UsageQueryRequest, error) {
	var request UsageQueryRequest
	err := r.run(ctx, source, func(_ *goja.Runtime, config *goja.Object) error {
		requestValue := config.Get("request")
		if goja.IsUndefined(requestValue) || goja.IsNull(requestValue) {
			return errors.New("script must define request")
		}
		if _, ok := goja.AssertFunction(config.Get("extractor")); !ok {
			return errors.New("script must define extractor as a function")
		}
		if err := validateUsageQueryResponseVersion(config); err != nil {
			return err
		}

		if err := exportJSON(requestValue, &request); err != nil {
			return fmt.Errorf("invalid request object: %w", err)
		}

		var err error
		request.URL, err = replaceUsageQueryVariables(request.URL, variables)
		if err != nil {
			return fmt.Errorf("invalid request URL: %w", err)
		}
		for key, value := range request.Headers {
			request.Headers[key], err = replaceUsageQueryVariables(value, variables)
			if err != nil {
				return fmt.Errorf("invalid request header %q: %w", key, err)
			}
		}
		request.Body, err = replaceUsageQueryValue(request.Body, variables)
		return err
	})
	if err != nil {
		return UsageQueryRequest{}, err
	}

	return request, nil
}

func (r *GojaUsageQueryRuntime) Extract(
	ctx context.Context,
	source string,
	response UsageQueryHTTPResponse,
	scriptContext UsageQueryScriptContext,
) (UsageQueryResult, error) {
	var result UsageQueryResult
	err := r.run(ctx, source, func(vm *goja.Runtime, config *goja.Object) error {
		extractor, ok := goja.AssertFunction(config.Get("extractor"))
		if !ok {
			return errors.New("script must define extractor as a function")
		}

		arguments := []goja.Value{vm.ToValue(response.Body)}
		if usageQueryUsesResponseEnvelope(config) {
			arguments = []goja.Value{
				vm.ToValue(map[string]any{
					"status":  response.Status,
					"headers": response.Headers,
					"body":    response.Body,
				}),
				vm.ToValue(map[string]any{"now": scriptContext.Now}),
			}
		}
		value, err := extractor(goja.Undefined(), arguments...)
		if err != nil {
			return fmt.Errorf("extractor failed: %w", err)
		}
		if promise, ok := value.Export().(*goja.Promise); ok {
			switch promise.State() {
			case goja.PromiseStateFulfilled:
				value = promise.Result()
			case goja.PromiseStateRejected:
				return fmt.Errorf("extractor promise rejected: %v", promise.Result())
			default:
				return errors.New("extractor returned a pending promise; asynchronous host APIs are not available")
			}
		}
		if goja.IsUndefined(value) || goja.IsNull(value) {
			return errors.New("extractor must return an object")
		}
		if err := exportJSON(value, &result); err != nil {
			return fmt.Errorf("invalid extractor result: %w", err)
		}
		return validateUsageQueryResult(result)
	})
	if err != nil {
		return UsageQueryResult{}, err
	}

	return result, nil
}

func (r *GojaUsageQueryRuntime) run(
	ctx context.Context,
	source string,
	fn func(vm *goja.Runtime, config *goja.Object) error,
) error {
	if len(source) == 0 {
		return errors.New("script is required")
	}
	if len(source) > maxUsageQueryScriptBytes {
		return fmt.Errorf("script exceeds %d bytes", maxUsageQueryScriptBytes)
	}

	program, err := goja.Compile("usage-query.js", source, false)
	if err != nil {
		return fmt.Errorf("failed to compile script: %w", err)
	}

	runCtx, cancel := context.WithTimeout(ctx, usageQueryScriptTimeout)
	defer cancel()

	vm := goja.New()
	stopInterrupt := context.AfterFunc(runCtx, func() {
		vm.Interrupt(runCtx.Err())
	})
	defer stopInterrupt()

	value, err := vm.RunProgram(program)
	if err != nil {
		var interrupted *goja.InterruptedError
		if errors.As(err, &interrupted) {
			return fmt.Errorf("script execution timed out: %w", runCtx.Err())
		}
		return fmt.Errorf("failed to evaluate script: %w", err)
	}
	config := value.ToObject(vm)
	if config == nil {
		return errors.New("script must evaluate to an object")
	}

	return fn(vm, config)
}

func exportJSON(value goja.Value, target any) error {
	data, err := json.Marshal(value.Export())
	if err != nil {
		return err
	}
	return json.Unmarshal(data, target)
}

func replaceUsageQueryVariables(value string, variables map[string]string) (string, error) {
	for key, replacement := range variables {
		value = strings.ReplaceAll(value, "{{"+key+"}}", replacement)
	}
	if start := strings.Index(value, "{{"); start >= 0 {
		if end := strings.Index(value[start+2:], "}}"); end >= 0 {
			return "", fmt.Errorf("unknown variable %q", value[start:start+end+4])
		}
	}
	return value, nil
}

func replaceUsageQueryValue(value any, variables map[string]string) (any, error) {
	switch typed := value.(type) {
	case nil:
		return nil, nil
	case string:
		return replaceUsageQueryVariables(typed, variables)
	case []any:
		result := make([]any, len(typed))
		for i, item := range typed {
			var err error
			result[i], err = replaceUsageQueryValue(item, variables)
			if err != nil {
				return nil, err
			}
		}
		return result, nil
	case map[string]any:
		result := make(map[string]any, len(typed))
		for key, item := range typed {
			var err error
			result[key], err = replaceUsageQueryValue(item, variables)
			if err != nil {
				return nil, err
			}
		}
		return result, nil
	default:
		return value, nil
	}
}

func validateUsageQueryResult(result UsageQueryResult) error {
	if result.Balance != nil && !isFiniteUsageQueryNumber(result.Balance.Remaining) {
		return errors.New("balance.remaining must be a finite number")
	}

	legacy := result.legacyFields()
	for name, value := range map[string]*float64{
		"remaining": legacy.Remaining,
		"total":     legacy.Total,
		"used":      legacy.Used,
	} {
		if value != nil && !isFiniteUsageQueryNumber(*value) {
			return fmt.Errorf("%s must be a finite number", name)
		}
	}
	if result.Progress == nil {
		return nil
	}

	for index, window := range result.Progress.Windows {
		if !usageQueryProgressWindowID.MatchString(window.ID) {
			return fmt.Errorf("progress.windows[%d].id must be an ASCII English identifier", index)
		}
		if window.DurationSeconds != nil && *window.DurationSeconds <= 0 {
			return fmt.Errorf("progress.windows[%d].durationSeconds must be greater than zero", index)
		}
		if window.RemainingPercent != nil && (!isFiniteUsageQueryNumber(*window.RemainingPercent) || *window.RemainingPercent < 0 || *window.RemainingPercent > 100) {
			return fmt.Errorf("progress.windows[%d].remainingPercent must be between 0 and 100", index)
		}
		if window.ResetAt != "" && !isUsageQueryOffsetTime(window.ResetAt) {
			return fmt.Errorf("progress.windows[%d].resetAt must use YYYY-MM-DDTHH:MM:SS±HH:MM", index)
		}
	}
	return nil
}

func isFiniteUsageQueryNumber(value float64) bool {
	return !math.IsNaN(value) && !math.IsInf(value, 0)
}

func isUsageQueryOffsetTime(value string) bool {
	if len(value) != len("2006-01-02T15:04:05+08:00") {
		return false
	}
	if _, err := time.Parse("2006-01-02T15:04:05-07:00", value); err != nil {
		return false
	}
	return value[19] == '+' || value[19] == '-'
}

type usageQueryLegacyFields struct {
	IsValid        *bool
	InvalidMessage string
	Remaining      *float64
	Unit           string
	PlanName       string
	Total          *float64
	Used           *float64
	Extra          string
}

func (r UsageQueryResult) legacyFields() usageQueryLegacyFields {
	return usageQueryLegacyFields{
		IsValid:        r.IsValid,
		InvalidMessage: r.InvalidMessage,
		Remaining:      r.Remaining,
		Unit:           r.Unit,
		PlanName:       r.PlanName,
		Total:          r.Total,
		Used:           r.Used,
		Extra:          r.Extra,
	}
}

func usageQueryUsesResponseEnvelope(config *goja.Object) bool {
	version, present, err := usageQueryResponseVersion(config)
	return err == nil && present && version == 2
}

func validateUsageQueryResponseVersion(config *goja.Object) error {
	version, present, err := usageQueryResponseVersion(config)
	if err != nil {
		return err
	}
	if !present {
		return nil
	}
	if version != 2 {
		return errors.New("responseVersion must be 2 when provided")
	}
	return nil
}

func usageQueryResponseVersion(config *goja.Object) (int, bool, error) {
	value := config.Get("responseVersion")
	if value == nil || goja.IsUndefined(value) || goja.IsNull(value) {
		return 0, false, nil
	}

	var version float64
	if err := exportJSON(value, &version); err != nil || !isFiniteUsageQueryNumber(version) || math.Trunc(version) != version {
		return 0, false, errors.New("responseVersion must be an integer")
	}
	return int(version), true, nil
}
