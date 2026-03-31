package errors

type AppError struct {
	Status  int
	Code    string
	Message string
	Details interface{}
}

func (e *AppError) Error() string {
	if e == nil {
		return ""
	}
	return e.Message
}

func NewAppError(status int, code string, message string, details interface{}) *AppError {
	return &AppError{Status: status, Code: code, Message: message, Details: details}
}
