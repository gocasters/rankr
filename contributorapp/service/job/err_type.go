package job

type RecordErrType int

const (
	ErrTypeValidation RecordErrType = iota + 1
	ErrTypeUnexpected
)

type RecordProcessError struct {
	Type RecordErrType
	Err  error
}

func (e RecordProcessError) Error() string {
	return e.Err.Error()
}
