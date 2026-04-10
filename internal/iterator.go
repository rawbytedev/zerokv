package internal

// IteratorErrors holds a slice of errors for iterators.
type IteratorErrors struct {
	errs []error
}

// AddError adds an error to the slice.
func (ie *IteratorErrors) AddError(err error) {
	if err != nil {
		ie.errs = append(ie.errs, err)
	}
}

// Error returns the last error or nil.
func (ie *IteratorErrors) Error() error {
	if len(ie.errs) == 0 {
		return nil
	}
	return ie.errs[len(ie.errs)-1]
}
