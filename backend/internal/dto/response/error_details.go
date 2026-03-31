package dtoresp

type ParamErrorDetails struct {
	Param string `json:"param"`
}

type FieldErrorItem struct {
	Field string `json:"field"`
	Tag   string `json:"tag"`
	Param string `json:"param,omitempty"`
}

type ValidationErrorDetails struct {
	Errors []FieldErrorItem `json:"errors"`
}
