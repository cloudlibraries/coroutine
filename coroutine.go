package coroutine

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/cloudlibraries/safe"
)

var (
	ErrCoroutineIsClosed = errors.New("coroutine is closed")
	ErrInvalidFunction   = errors.New("invalid function")
)

type (
	// Status is the status of a Coroutine
	Status int

	// Coroutine is a coroutine
	Coroutine struct {
		status   Status
		inputCh  chan []any
		outputCh chan []any
		function func(*Coroutine, ...any) error
	}
)

const (
	// CREATED means coroutine is created and not started.
	CREATED Status = iota

	// RUNNING means coroutine is started and running.
	RUNNING

	// SUSPENDED means coroutine is started and yielded.
	SUSPENDED

	// CLOSED means coroutine not created or ended.
	CLOSED
)

var statusStringMap = map[Status]string{
	CREATED:   "Created",
	SUSPENDED: "Suspended",
	RUNNING:   "Running",
	CLOSED:    "Closed",
}

func (s Status) String() string {
	if v, ok := statusStringMap[s]; ok {
		return v
	}
	return fmt.Sprintf("Unknown Status: %d", s)
}

// Start starts a coroutine.
func Start(v any) error {
	c, err := Create(v)
	if err != nil {
		return err
	}

	_, err = c.Resume(nil)
	return err
}

// Create wraps starts a coroutine up.
func Create(v any) (*Coroutine, error) {
	c := &Coroutine{
		inputCh:  make(chan []any, 1),
		outputCh: make(chan []any, 1),
	}

	switch v := v.(type) {
	case func(_ *Coroutine, _ ...any) error:
		c.function = v
	case func(*Coroutine, ...any):
		c.function = func(c *Coroutine, args ...any) error {
			v(c, args...)
			return nil
		}
	case func(*Coroutine) error:
		c.function = func(c *Coroutine, _ ...any) error {
			return v(c)
		}
	case func(*Coroutine):
		c.function = func(c *Coroutine, _ ...any) error {
			v(c)
			return nil
		}
	case func() error:
		c.function = func(*Coroutine, ...any) error {
			return v()
		}
	case func():
		c.function = func(*Coroutine, ...any) error {
			v()
			return nil
		}
	default:
		return nil, ErrInvalidFunction
	}

	c.status = CREATED

	go safe.Do(func() error {
		defer func() {
			c.status = CLOSED
			close(c.inputCh)
			close(c.outputCh)
		}()

		input, err := c.Yield()
		if err != nil {
			return err
		}

		return c.function(c, input...)
	})

	return c, nil
}

// Resume continues a suspend ID, passing data in and out.
func Resume(c *Coroutine, input ...any) (output []any, err error) {
	return c.Resume(input...)
}

// ResumeWithContext resumes a coroutine with context.
func ResumeWithContext(ctx context.Context, c *Coroutine, input ...any) (output []any, err error) {
	return c.ResumeWithContext(ctx, input...)
}

// ResumeWithTimeout resumes a coroutine with timeout.
func ResumeWithTimeout(c *Coroutine, timeout time.Duration, input ...any) (output []any, err error) {
	return c.ResumeWithTimeout(timeout, input...)
}

// Resume continues a suspend ID, passing data in and out.
func (c *Coroutine) Resume(input ...any) (output []any, err error) {
	err = safe.Do(func() error {
		switch c.status {
		case CLOSED:
			return ErrCoroutineIsClosed

		default:
			output = <-c.outputCh
			c.inputCh <- input
		}

		return nil
	})

	return
}

// ResumeWithContext resumes a coroutine with context.
func (c *Coroutine) ResumeWithContext(ctx context.Context, input ...any) (output []any, err error) {
	err = safe.DoWithContext(ctx, func() error {
		switch c.status {
		case CLOSED:
			return ErrCoroutineIsClosed

		default:
			output = <-c.outputCh
			c.inputCh <- input
		}

		return nil
	})

	return
}

// ResumeWithTimeout resumes a coroutine with timeout.
func (c *Coroutine) ResumeWithTimeout(timeout time.Duration, input ...any) (output []any, err error) {
	err = safe.DoWithTimeout(timeout, func() error {
		switch c.status {
		case CLOSED:
			return ErrCoroutineIsClosed

		default:
			output = <-c.outputCh
			c.inputCh <- input
		}

		return nil
	})

	return
}

// Yield suspends a running coroutine, passing data in and out.
func Yield(c *Coroutine, input ...any) (output []any, err error) {
	return c.Yield(input...)
}

// YieldWithContext suspends a running coroutine with context, passing data in and out.
func YieldWithContext(ctx context.Context, c *Coroutine, input ...any) (output []any, err error) {
	return c.YieldWithContext(ctx, input...)
}

// YieldWithTimeout suspends a running coroutine with timeout, passing data in and out.
func YieldWithTimeout(c *Coroutine, timeout time.Duration, input ...any) (output []any, err error) {
	return c.YieldWithTimeout(timeout, input...)
}

// Yield suspends a running coroutine, passing data in and out.
func (c *Coroutine) Yield(output ...any) (input []any, err error) {
	err = safe.Do(func() error {
		switch c.status {
		case CLOSED:
			return ErrCoroutineIsClosed

		case CREATED:
			c.outputCh <- output
			input = <-c.inputCh

			c.status = RUNNING

		default:
			c.status = SUSPENDED

			c.outputCh <- output
			input = <-c.inputCh

			c.status = RUNNING
		}

		return nil
	})

	return
}

// YieldWithContext suspends a running coroutine with context, passing data in and out.
func (c *Coroutine) YieldWithContext(ctx context.Context, output ...any) (input []any, err error) {
	err = safe.DoWithContext(ctx, func() error {
		switch c.status {
		case CLOSED:
			return ErrCoroutineIsClosed

		case CREATED:
			c.outputCh <- output
			input = <-c.inputCh

			c.status = RUNNING

		default:
			c.status = SUSPENDED

			c.outputCh <- output
			input = <-c.inputCh

			c.status = RUNNING
		}

		return nil
	})

	return
}

// YieldWithTimeout suspends a running coroutine with timeout, passing data in and out.
func (c *Coroutine) YieldWithTimeout(timeout time.Duration, output ...any) (input []any, err error) {
	err = safe.DoWithTimeout(timeout, func() error {
		switch c.status {
		case CLOSED:
			return ErrCoroutineIsClosed

		case CREATED:
			c.outputCh <- output
			input = <-c.inputCh

			c.status = RUNNING

		default:
			c.status = SUSPENDED

			c.outputCh <- output
			input = <-c.inputCh

			c.status = RUNNING
		}

		return nil
	})

	return
}
