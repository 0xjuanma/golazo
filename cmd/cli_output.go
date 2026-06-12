package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"sort"

	"github.com/0xjuanma/golazo/internal/api"
)

// ErrorCode is a typed, machine-readable error category for CLI consumers.
type ErrorCode string

const (
	ErrCodeInvalidArgs   ErrorCode = "invalid_args"
	ErrCodeNotFound      ErrorCode = "not_found"
	ErrCodeUpstreamError ErrorCode = "upstream_error"
	ErrCodeTimeout       ErrorCode = "timeout"
	ErrCodeOffline       ErrorCode = "offline"
)

// Exit codes mapped from ErrorCode. Documented in docs/cli.md.
const (
	ExitOK           = 0
	ExitUpstream     = 1
	ExitInvalidArgs  = 2
	ExitNotFound     = 3
	ExitTimeout      = 4
	ExitOffline      = 5
)

// ExitCodeFor returns the documented exit code for a given ErrorCode.
func ExitCodeFor(code ErrorCode) int {
	switch code {
	case ErrCodeInvalidArgs:
		return ExitInvalidArgs
	case ErrCodeNotFound:
		return ExitNotFound
	case ErrCodeTimeout:
		return ExitTimeout
	case ErrCodeOffline:
		return ExitOffline
	case ErrCodeUpstreamError:
		return ExitUpstream
	default:
		return ExitUpstream
	}
}

// okEnvelope is the shape returned on success.
type okEnvelope struct {
	Status        string `json:"status"`
	Degraded      bool   `json:"degraded,omitempty"`
	FailedDates   []string `json:"failed_dates,omitempty"`
	Count         int    `json:"count"`
	Data          any    `json:"data"`
}

// errEnvelope is the shape returned on failure.
type errEnvelope struct {
	Status  string    `json:"status"`
	Code    ErrorCode `json:"code"`
	Message string    `json:"message"`
}

// Pretty controls whether JSON is indented. Default is compact (agent-friendly).
var Pretty bool

// WriteJSON writes a successful envelope to w. count is set from data length
// when data is a slice; for single-object responses callers should wrap in
// a one-element slice so count stays consistent.
func WriteJSON(w io.Writer, data any) error {
	env := okEnvelope{
		Status: "ok",
		Count:  sliceLen(data),
		Data:   nonNilSlice(data),
	}
	return encode(w, env)
}

// WriteDegraded writes a successful envelope flagged as degraded. failedDates
// is the list of date strings (YYYY-MM-DD) whose upstream fetch failed but
// other days succeeded.
func WriteDegraded(w io.Writer, data any, failedDates []string) error {
	env := okEnvelope{
		Status:      "ok",
		Degraded:    true,
		FailedDates: failedDates,
		Count:       sliceLen(data),
		Data:        nonNilSlice(data),
	}
	return encode(w, env)
}

// WriteError writes the error envelope to w (typically stderr) and returns
// the documented exit code for the given ErrorCode.
func WriteError(w io.Writer, code ErrorCode, err error) int {
	msg := ""
	if err != nil {
		msg = err.Error()
	}
	env := errEnvelope{
		Status:  "error",
		Code:    code,
		Message: msg,
	}
	_ = encode(w, env)
	return ExitCodeFor(code)
}

// ClassifyClientError maps a transport/client error to an ErrorCode.
// Callers pass an error from a fotmob.Client call and a flag indicating
// whether the context deadline was exceeded.
func ClassifyClientError(err error, timedOut bool) ErrorCode {
	if err == nil {
		return ErrCodeUpstreamError
	}
	if timedOut {
		return ErrCodeTimeout
	}
	return ErrCodeUpstreamError
}

// SortMatches sorts matches deterministically by MatchTime (nils last) then ID.
// In-place sort.
func SortMatches(matches []api.Match) {
	sort.SliceStable(matches, func(i, j int) bool {
		a, b := matches[i], matches[j]
		switch {
		case a.MatchTime != nil && b.MatchTime != nil:
			if !a.MatchTime.Equal(*b.MatchTime) {
				return a.MatchTime.Before(*b.MatchTime)
			}
		case a.MatchTime != nil && b.MatchTime == nil:
			return true
		case a.MatchTime == nil && b.MatchTime != nil:
			return false
		}
		return a.ID < b.ID
	})
}

func encode(w io.Writer, v any) error {
	enc := json.NewEncoder(w)
	if Pretty {
		enc.SetIndent("", "  ")
	}
	enc.SetEscapeHTML(false)
	return enc.Encode(v)
}

// sliceLen returns the length when v is a slice; otherwise 0 if nil, else 1.
func sliceLen(v any) int {
	switch s := v.(type) {
	case nil:
		return 0
	case []api.Match:
		return len(s)
	case []api.MatchDetails:
		return len(s)
	case []api.League:
		return len(s)
	case []api.LeagueTableEntry:
		return len(s)
	case []any:
		return len(s)
	}
	// Single-object fallback. Callers are expected to wrap singles in []T.
	// Slices of an unrecognized type fall through here too and report 1; if
	// you add a new agent-facing list type, register it above.
	return 1
}

// nonNilSlice returns v unless v is a nil slice; in that case it returns an
// empty slice of the same element type so JSON encodes `[]` instead of `null`.
func nonNilSlice(v any) any {
	switch s := v.(type) {
	case []api.Match:
		if s == nil {
			return []api.Match{}
		}
	case []api.MatchDetails:
		if s == nil {
			return []api.MatchDetails{}
		}
	case []api.League:
		if s == nil {
			return []api.League{}
		}
	case []api.LeagueTableEntry:
		if s == nil {
			return []api.LeagueTableEntry{}
		}
	case []any:
		if s == nil {
			return []any{}
		}
	}
	return v
}

// errInvalidArg is a sentinel for invalid argument errors so callers can build
// a precise message without re-classifying.
var errInvalidArg = errors.New("invalid argument")

// NewInvalidArg returns an error wrapping errInvalidArg with a formatted msg.
func NewInvalidArg(format string, args ...any) error {
	return fmt.Errorf("%w: "+format, append([]any{errInvalidArg}, args...)...)
}
