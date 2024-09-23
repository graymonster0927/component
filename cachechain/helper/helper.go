package helper

type ErrHelper struct {
	Err error
}

func (e *ErrHelper) IsSuccess() bool {
	return e.Err == nil
}
