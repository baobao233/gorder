package consts

const (
	ErrnoSuccess = 0
	ErrnoUnknown = 1

	// param error 1xxx  在接口层
	ErrnoBindRequest          = 1000
	ErrnoRequestValidateError = 1001
)

var ErrMsg = map[int]string{
	ErrnoSuccess: "success",
	ErrnoUnknown: "unknown error",

	ErrnoBindRequest:          "bind request error",
	ErrnoRequestValidateError: "validate request error",
}
