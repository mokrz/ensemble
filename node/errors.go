package node

type ErrNotFound struct {
	name string
	inner error
}

func (e ErrNotFound) Error() string {
	return e.name + " not found"
}

func (e ErrNotFound) Unwrap() error {
	return e.inner
}
