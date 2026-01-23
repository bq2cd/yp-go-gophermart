package apiclient

import "errors"

var (
	// ErrUnexpectedHTTPStatus is returned by [Client] when it receives an unexpected HTTP status in HTTP response.
	ErrUnexpectedHTTPStatus = errors.New("unexpected HTTP response status")
)
