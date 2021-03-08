package multierror

import (
	"errors"
	"sync"
)

// MultiError type handles error accumulation from goroutines
type MultiError struct {
	errors []error
	mux    sync.Mutex
}

// Error turns the MultiError structure into a string
func (err *MultiError) Error() string {
	str := ""

	err.mux.Lock()
	for _, e := range err.errors {
		str += e.Error() + "\n"
	}
	err.mux.Unlock()

	return str[:len(str)-1]
}

// Add adds an error to the Multierror structure
func (err *MultiError) Add(e error) {
	if e == nil {
		return
	}

	err.mux.Lock()
	err.errors = append(err.errors, e)
	err.mux.Unlock()
}

// Return is used as a wrapper on return on whether to return the
// MultiError Structure if errors exist or nil instead of delivering an empty structure
func (err *MultiError) Return() error {
	if len(err.errors) > 0 {
		return err
	}

	return nil
}

func (err *MultiError) Is(other error) bool {
	err.mux.Lock()
	defer err.mux.Unlock()
	for _, e := range err.errors {
		if errors.Is(e, other) {
			return true
		}
	}
	return false
}

func (err *MultiError) As(other interface{}) bool {
	err.mux.Lock()
	defer err.mux.Unlock()
	for _, e := range err.errors {
		if errors.As(e, other) {
			return true
		}
	}
	return false
}
