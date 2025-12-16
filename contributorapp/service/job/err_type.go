package job

type RecordErrType int

const (
	ErrTypeValidation RecordErrType = iota + 1
	ErrTypeUnexpect
)

type RecordErr struct {
	ErrType RecordErrType
	err     error
}

func (e RecordErr) Error() string {
	return e.err.Error()
}
