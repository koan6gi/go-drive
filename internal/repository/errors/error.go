package errors

type PathError struct {
	Err     error
	Content string
}

func (e *PathError) Error() string { return e.Content }

type SystemError struct {
	Err     error
	Content string
}

func (e *SystemError) Error() string { return e.Content }
