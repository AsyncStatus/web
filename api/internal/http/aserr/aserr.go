package aserr

import (
	"errors"
	"net/http"

	"github.com/go-playground/validator/v10"
)

type ASError struct {
	error          `json:"-"`
	ErrorsMap      validator.ValidationErrorsTranslations `json:"errors_map,omitempty"`
	Message        string                                 `json:"message" validate:"required"`
	Details        string                                 `json:"details,omitempty" validate:"required"`
	HTTPStatusCode int                                    `json:"-"`
	InternalStatus string                                 `json:"internal_status" validate:"required"`
} // @name ASError

func NewASError(err error, message string, details string, httpStatusCode int, internalStatus string) *ASError {
	return &ASError{error: err, Message: message, Details: details, HTTPStatusCode: httpStatusCode, InternalStatus: internalStatus}
}
func (e ASError) Error() string {
	return e.error.Error()
}

// NewWithError creates new ASError from error. Only use when you're sure you wouldn't expose sensitive information.
func (e ASError) NewWithError(err error) *ASError {
	return &ASError{error: err, Message: err.Error(), Details: e.Details, HTTPStatusCode: e.HTTPStatusCode, InternalStatus: e.InternalStatus}
}

// NewWithErrorMap creates new ASError from error with validation errors map. Only use when you're sure you wouldn't expose sensitive information.
func (e ASError) NewWithErrorMap(err error, errorsMap validator.ValidationErrorsTranslations) *ASError {
	return &ASError{error: err, Message: err.Error(), Details: e.Details, HTTPStatusCode: e.HTTPStatusCode, InternalStatus: e.InternalStatus, ErrorsMap: errorsMap}
}

var (
	ErrBadRequest = NewASError(
		errors.New("Bad request"),
		"Bad Request",
		"",
		http.StatusBadRequest, // 400
		"ASE-400",
	)

	ErrInvalidSignedValueFormat = NewASError(
		errors.New("Invalid signed value format"),
		"Invalid signed value format",
		"",
		http.StatusBadRequest, // 400
		"ASE-400/signed-value-format",
	)

	ErrInvalidSignedValueTimestamp = NewASError(
		errors.New("Invalid signed value timestamp"),
		"Invalid signed value timestamp",
		"",
		http.StatusBadRequest, // 400
		"ASE-400/signed-value-timestamp",
	)

	ErrInvalidSignedValueSignature = NewASError(
		errors.New("Invalid signed value signature"),
		"Invalid signed value signature",
		"",
		http.StatusBadRequest, // 400
		"ASE-400/signed-value-signature",
	)

	ErrUnauthorized = NewASError(
		errors.New("Unauthorized"),
		"Unauthorized",
		"",
		http.StatusUnauthorized, // 401
		"ASE-401",
	)

	ErrUnauthorizedAccountNotVerified = NewASError(
		errors.New("Account not verified"),
		"Account not verified",
		"Your account is not verified, please check your email.",
		http.StatusUnauthorized, // 401
		"ASE-401/account-not-verified",
	)

	ErrInvalidEmailOrPassword = NewASError(
		errors.New("Invalid email or password"),
		"Invalid email or password",
		"",
		http.StatusUnauthorized, // 401
		"ASE-401/invalid-email-or-password",
	)

	ErrPaymentRequired = NewASError(
		errors.New("Payment required"),
		"Payment required",
		"",
		http.StatusPaymentRequired, // 402
		"ASE-402",
	)

	ErrForbidden = NewASError(
		errors.New("Forbidden"),
		"Forbidden",
		"",
		http.StatusForbidden, // 403
		"ASE-403",
	)

	ErrNotFound = NewASError(
		errors.New("Not found"),
		"Not found",
		"",
		http.StatusNotFound, // 404
		"ASE-404",
	)

	ErrUserNotFound = NewASError(
		errors.New("User not found"),
		"User not found",
		"",
		http.StatusNotFound, // 404
		"ASE-404/user",
	)

	ErrProjectNotFound = NewASError(
		errors.New("Project not found"),
		"Project not found",
		"",
		http.StatusNotFound, // 404
		"ASE-404/project",
	)

	ErrProjectTeamNotFound = NewASError(
		errors.New("Project team not found"),
		"Project team not found",
		"",
		http.StatusNotFound, // 404
		"ASE-404/project-team",
	)

	ErrNoProjectMemberships = NewASError(
		errors.New("No project memberships"),
		"No project memberships",
		"You are not a member of any project.",
		http.StatusNotFound, // 404
		"ASE-404/project-memberships",
	)

	ErrProjectTeamMembershipNotFound = NewASError(
		errors.New("Project team membership not found"),
		"You're not a member of this team",
		"",
		http.StatusNotFound, // 404
		"ASE-404/project-team-membership",
	)

	ErrUserNotFoundSignIn = NewASError(
		errors.New("User not found"),
		"User not found",
		"",
		http.StatusNotFound, // 404
		"ASE-404/user-not-found-sign-in",
	)

	ErrMethodNotAllowed = NewASError(
		errors.New("Method not allowed"),
		"Method not allowed",
		"",
		http.StatusMethodNotAllowed, // 405
		"ASE-405",
	)

	ErrNotAcceptable = NewASError(
		errors.New("Not acceptable"),
		"Not acceptable",
		"",
		http.StatusNotAcceptable, // 406
		"ASE-406",
	)

	ErrProxyAuthRequired = NewASError(
		errors.New("Proxy authentication required"),
		"Proxy authentication required",
		"Authentication with the proxy is required.",
		http.StatusProxyAuthRequired, // 407
		"ASE-407",
	)

	ErrRequestTimeout = NewASError(
		errors.New("Request timeout"),
		"Request timeout",
		"The server timed out waiting for the request.",
		http.StatusRequestTimeout, // 408
		"ASE-408",
	)

	ErrConflict = NewASError(
		errors.New("Conflict"),
		"Conflict",
		"The request could not be completed due to a conflict with the current state of the resource.",
		http.StatusConflict, // 409
		"ASE-409",
	)

	ErrUserAlreadyExists = NewASError(
		errors.New("User already exists"),
		"User already exists",
		"User with this email already exists.",
		http.StatusConflict, // 409
		"ASE-409/user-already-exists",
	)

	ErrProjectConflictName = NewASError(
		errors.New("Project with name already exists"),
		"Project with name already exists",
		"Project with the given name already exists, please choose another name.",
		http.StatusInternalServerError, // 500
		"ASE-500/project-name-already-exists",
	)

	ErrProjectTeamConflictName = NewASError(
		errors.New("Project team with name already exists"),
		"Project team with name already exists",
		"Project team with the given name already exists, please choose another name.",
		http.StatusInternalServerError, // 500
		"ASE-500/project-team-name-already-exists",
	)

	ErrGone = NewASError(
		errors.New("Gone"),
		"Gone",
		"The requested resource is no longer available and will not be available again.",
		http.StatusGone, // 410
		"ASE-410",
	)

	ErrLengthRequired = NewASError(
		errors.New("Length required"),
		"Length required",
		"The request did not specify the length of its content, which is required.",
		http.StatusLengthRequired, // 411
		"ASE-411",
	)

	ErrPreconditionFailed = NewASError(
		errors.New("Precondition failed"),
		"Precondition failed",
		"One or more conditions given in the request header fields evaluated to false.",
		http.StatusPreconditionFailed, // 412
		"ASE-412",
	)

	ErrPayloadTooLarge = NewASError(
		errors.New("Payload too large"),
		"Payload too large",
		"The request is larger than the server is willing or able to process.",
		http.StatusRequestEntityTooLarge, // 413
		"ASE-413",
	)

	ErrURITooLong = NewASError(
		errors.New("URI too long"),
		"URI too long",
		"The URI provided was too long for the server to process.",
		http.StatusRequestURITooLong, // 414
		"ASE-414",
	)

	ErrUnsupportedMediaType = NewASError(
		errors.New("Unsupported media type"),
		"Unsupported media type",
		"The request entity has a media type which the server or resource does not support.",
		http.StatusUnsupportedMediaType, // 415
		"ASE-415",
	)

	ErrRangeNotSatisfiable = NewASError(
		errors.New("Range not satisfiable"),
		"Range not satisfiable",
		"None of the ranges in the request's Range header overlap the current extent of the resource.",
		http.StatusRequestedRangeNotSatisfiable, // 416
		"ASE-416",
	)

	ErrExpectationFailed = NewASError(
		errors.New("Expectation failed"),
		"Expectation failed",
		"The expectation given in the request's Expect header could not be met.",
		http.StatusExpectationFailed, // 417
		"ASE-417",
	)

	ErrImATeapot = NewASError(
		errors.New("I'm a teapot"),
		"I'm a teapot",
		"The server refuses to brew coffee because it is permanently a teapot.",
		http.StatusTeapot, // 418
		"ASE-418",
	)

	ErrUpgradeRequired = NewASError(
		errors.New("Upgrade required"),
		"Upgrade required",
		"The client should switch to a different protocol such as TLS/1.0.",
		http.StatusUpgradeRequired, // 426
		"ASE-426",
	)

	ErrPreconditionRequired = NewASError(
		errors.New("Precondition required"),
		"Precondition required",
		"The origin server requires the request to be conditional.",
		http.StatusPreconditionRequired, // 428
		"ASE-428",
	)

	ErrTooManyRequests = NewASError(
		errors.New("Too many requests"),
		"Too many requests",
		"You've sent too many requests.",
		http.StatusTooManyRequests, // 429
		"ASE-429",
	)

	ErrRequestHeaderFieldsTooLarge = NewASError(
		errors.New("Request header fields too large"),
		"Request header fields too large",
		"The server is unwilling to process the request because its header fields are too large.",
		http.StatusRequestHeaderFieldsTooLarge, // 431
		"ASE-431",
	)

	ErrUnavailableForLegalReasons = NewASError(
		errors.New("Unavailable for legal reasons"),
		"Unavailable for legal reasons",
		"The server is denying access to the resource as a consequence of a legal demand.",
		http.StatusUnavailableForLegalReasons, // 451
		"ASE-451",
	)

	ErrInternalServerError = NewASError(
		errors.New("Internal server error"),
		"Internal server error",
		"The server encountered an unexpected condition that prevented it from fulfilling the request.",
		http.StatusInternalServerError, // 500
		"ASE-500",
	)

	ErrNotImplemented = NewASError(
		errors.New("Not implemented"),
		"Not implemented",
		"The server does not support the functionality required to fulfill the request.",
		http.StatusNotImplemented, // 501
		"ASE-501",
	)

	ErrBadGateway = NewASError(
		errors.New("Bad gateway"),
		"Bad gateway",
		"The server received an invalid response from the upstream server.",
		http.StatusBadGateway, // 502
		"ASE-502",
	)

	ErrServiceUnavailable = NewASError(
		errors.New("Service unavailable"),
		"Service unavailable",
		"The server is currently unable to handle the request due to temporary overloading or maintenance.",
		http.StatusServiceUnavailable, // 503
		"ASE-503",
	)

	ErrPgUnavailable = NewASError(
		errors.New("Database unavailable"),
		"Database unavailable",
		"The server is unable to connect to the database.",
		http.StatusServiceUnavailable, // 503
		"ASE-503/pg",
	)

	ErrRedisUnavailable = NewASError(
		errors.New("Redis unavailable"),
		"Redis unavailable",
		"The server is unable to connect to Redis.",
		http.StatusServiceUnavailable, // 503
		"ASE-503/redis",
	)

	ErrGatewayTimeout = NewASError(
		errors.New("Gateway timeout"),
		"Gateway timeout",
		"The server did not receive a timely response from the upstream server.",
		http.StatusGatewayTimeout, // 504
		"ASE-504",
	)

	ErrHTTPVersionNotSupported = NewASError(
		errors.New("HTTP version not supported"),
		"HTTP version not supported",
		"The server does not support the HTTP protocol version used in the request.",
		http.StatusHTTPVersionNotSupported, // 505
		"ASE-505",
	)

	ErrVariantAlsoNegotiates = NewASError(
		errors.New("Variant also negotiates"),
		"Variant also negotiates",
		"Transparent content negotiation for the request results in a circular reference.",
		506,
		"ASE-506",
	)

	ErrInsufficientStorage = NewASError(
		errors.New("Insufficient storage"),
		"Insufficient storage",
		"The server is unable to store the representation needed to complete the request.",
		507,
		"ASE-507",
	)

	ErrLoopDetected = NewASError(
		errors.New("Loop detected"),
		"Loop detected",
		"The server detected an infinite loop while processing the request.",
		508,
		"ASE-508",
	)

	ErrNotExtended = NewASError(
		errors.New("Not extended"),
		"Not extended",
		"Further extensions to the request are required for the server to fulfill it.",
		510,
		"ASE-510",
	)

	ErrNetworkAuthenticationRequired = NewASError(
		errors.New("Network authentication required"),
		"Network authentication required",
		"The client needs to authenticate to gain network access.",
		511,
		"ASE-511",
	)
)
