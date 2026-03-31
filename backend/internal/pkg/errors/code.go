package errors

const (
	CodeOK                       = "OK"
	CodeInvalidArgument          = "INVALID_ARGUMENT"
	CodeUnauthorized             = "UNAUTHORIZED"
	CodeForbidden                = "FORBIDDEN"
	CodeNotFound                 = "NOT_FOUND"
	CodeConflict                 = "CONFLICT"
	CodeInternal                 = "INTERNAL_ERROR"
	CodeFirstLoginPasswordChange = "FIRST_LOGIN_PASSWORD_CHANGE_REQUIRED"

	CodeInvalidCredentials         = "INVALID_CREDENTIALS"
	CodeResourceTypeMismatch       = "RESOURCE_TYPE_MISMATCH"
	CodeGrantPreconditionNotMet    = "GRANT_PRECONDITION_NOT_MET"
	CodeResourcePermissionDenied   = "RESOURCE_PERMISSION_DENIED"
	CodeCredentialPermissionDenied = "CREDENTIAL_PERMISSION_DENIED"
)
