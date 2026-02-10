package apiclient

import "errors"

var (
	// ErrUnexpectedHTTPStatus is returned by [Client] when it receives an unexpected HTTP status in HTTP response.
	ErrUnexpectedHTTPStatus = errors.New("unexpected HTTP response status")

	// ErrUnknownOrderStatus is returned by [Client] when it cannot
	// parse accrual system's response.
	ErrUnknownOrderStatus = errors.New("unknown order status")

	// ErrUnexpectedAPIResultType is returned by [Client] when it receives unexpected
	// API response, e.g. wrong data structure.
	ErrUnexpectedAPIResultType = errors.New("unexpected API result type")
)
