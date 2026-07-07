package ipaynow

import "fmt"

// ErrorCode is the stable, language-neutral error code carried by IpaynowError.
// See spec §7 for the meaning of each code.
type ErrorCode string

const (
	// CodeSystemError marks an internal SDK error (e.g. impossible state).
	CodeSystemError ErrorCode = "E0001"
	// CodeHTTPException marks a connection failure or a non-2xx HTTP response.
	CodeHTTPException ErrorCode = "E0002"
	// CodeConnectTimeout marks a timeout while establishing the TCP connection.
	CodeConnectTimeout ErrorCode = "E0003"
	// CodeSocketTimeout marks a timeout while waiting for / reading the response.
	CodeSocketTimeout ErrorCode = "E0004"
	// CodeSignError marks a missing or invalid signature on a response/notification.
	CodeSignError ErrorCode = "E0012"
)

// IpaynowError is the single error type returned by every SDK failure path.
// Callers only need to type-assert this one type and branch on Code.
type IpaynowError struct {
	Code    ErrorCode
	Message string
	Cause   error
}

func (e *IpaynowError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("ipaynow: [%s] %s: %v", e.Code, e.Message, e.Cause)
	}
	return fmt.Sprintf("ipaynow: [%s] %s", e.Code, e.Message)
}

func (e *IpaynowError) Unwrap() error { return e.Cause }
