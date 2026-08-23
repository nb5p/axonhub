package llm

// RequestClient identifies the application that sent an inbound request.
//
// The value is derived once from the inbound User-Agent and then reused for
// routing and request persistence. Unknown deliberately means that no User-
// Agent was supplied; other means an unrecognised non-empty User-Agent.
type RequestClient string

const (
	RequestClientUnknown RequestClient = "unknown"
	RequestClientOther   RequestClient = "other"
	RequestClientCodex   RequestClient = "codex"
)

// NormalizeRequestClient converts an empty value to unknown and recognises the
// stable client categories accepted by channel settings.
func NormalizeRequestClient(client RequestClient) RequestClient {
	switch client {
	case RequestClientCodex, RequestClientOther, RequestClientUnknown:
		return client
	default:
		return RequestClientUnknown
	}
}

// IsKnownRequestClient reports whether client is a supported persisted client
// classification. Unlike NormalizeRequestClient, it does not accept empty
// values because configuration must be explicit.
func IsKnownRequestClient(client RequestClient) bool {
	switch client {
	case RequestClientCodex, RequestClientOther, RequestClientUnknown:
		return true
	default:
		return false
	}
}
