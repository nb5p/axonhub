package provider_quota

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/dop251/goja"
)

const (
	maxUsageQueryScriptBytes = 64 << 10
	usageQueryScriptTimeout  = 500 * time.Millisecond
)

// UsageQueryScriptRuntime is the replaceable boundary between quota business
// logic and the embedded JavaScript engine.
type UsageQueryScriptRuntime interface {
	ParseRequest(ctx context.Context, source string, variables map[string]string) (UsageQueryRequest, error)
	Extract(ctx context.Context, source string, response any) (UsageQueryResult, error)
}

type UsageQueryRequest struct {
	URL     string            `json:"url"`
	Method  string            `json:"method"`
	Headers map[string]string `json:"headers"`
	Body    any               `json:"body,omitempty"`
}

type UsageQueryResult struct {
	IsValid        *bool    `json:"isValid,omitempty"`
	InvalidMessage string   `json:"invalidMessage,omitempty"`
	Remaining      *float64 `json:"remaining,omitempty"`
	Unit           string   `json:"unit,omitempty"`
	PlanName       string   `json:"planName,omitempty"`
	Total          *float64 `json:"total,omitempty"`
	Used           *float64 `json:"used,omitempty"`
	Extra          string   `json:"extra,omitempty"`
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

func (r *GojaUsageQueryRuntime) Extract(ctx context.Context, source string, response any) (UsageQueryResult, error) {
	var result UsageQueryResult
	err := r.run(ctx, source, func(vm *goja.Runtime, config *goja.Object) error {
		extractor, ok := goja.AssertFunction(config.Get("extractor"))
		if !ok {
			return errors.New("script must define extractor as a function")
		}

		value, err := extractor(goja.Undefined(), vm.ToValue(response))
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
	for name, value := range map[string]*float64{
		"remaining": result.Remaining,
		"total":     result.Total,
		"used":      result.Used,
	} {
		if value != nil && (math.IsNaN(*value) || math.IsInf(*value, 0)) {
			return fmt.Errorf("%s must be a finite number", name)
		}
	}
	return nil
}
