package errs

import (
	"encoding/json"
	"errors"
	"fmt"
)

// MachineError is a structured error shape for automation.
type MachineError struct {
	Code         string `json:"code"`
	Message      string `json:"message"`
	Hint         string `json:"hint,omitempty"`
	Status       int    `json:"status,omitempty"`
	RetryAfter   int    `json:"retry_after_sec,omitempty"`
	RequestID    string `json:"request_id,omitempty"`
	WrappedError error  `json:"-"`
}

func (e *MachineError) Error() string {
	if e == nil {
		return ""
	}
	if e.Code == "" {
		return e.Message
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func (e *MachineError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.WrappedError
}

func New(code, message, hint string) *MachineError {
	return &MachineError{Code: code, Message: message, Hint: hint}
}

func Wrap(code, message, hint string, err error) *MachineError {
	return &MachineError{Code: code, Message: message, Hint: hint, WrappedError: err}
}

func AsMachine(err error) *MachineError {
	if err == nil {
		return nil
	}
	var m *MachineError
	if errors.As(err, &m) {
		return m
	}
	return &MachineError{Code: "internal_error", Message: err.Error(), WrappedError: err}
}

func ExitCode(err error) int {
	m := AsMachine(err)
	switch m.Code {
	case "invalid_argument", "invalid_config", "profile_not_found", "missing_secret":
		return 2
	case "auth_failed", "token_refresh_failed":
		return 3
	case "rate_limited":
		return 4
	case "api_not_found":
		return 5
	case "api_error":
		if m.Status == 403 || m.Status == 401 {
			return 3
		}
		return 6
	default:
		return 1
	}
}

func JSON(err error) string {
	m := AsMachine(err)
	b, marshalErr := json.Marshal(m)
	if marshalErr != nil {
		return fmt.Sprintf(`{"code":"internal_error","message":%q}`, err.Error())
	}
	return string(b)
}
