package coroutine

import (
	"fmt"
	"sync"
	"time"

	"github.com/cloudlibraries/cast"
	uuid "github.com/satori/go.uuid"
)

type (
	// StatusType is an alias for string.
	StatusType = string

	// ID is the unique identifier for coroutine.
	ID = string

	// Coroutine is a simulator struct for coroutine.
	Coroutine struct {
		id          ID
		status      StatusType
		inCh        chan []interface{}
		outCh       chan []interface{}
		fn          func(id ID, args ...interface{}) error
		mutexStatus *sync.Mutex
		mutexResume *sync.Mutex
	}
)

const (
	// Created means ID is created and not started.
	Created = "Created"

	// Suspended means ID is started and yielded.
	Suspended = "Suspended"

	// Running means ID is started and running.
	Running = "Running"

	// Dead means ID not created or ended.
	Dead = "Dead"
)

var coroutines sync.Map

// Start wraps and starts a ID up.
// It is thread-safe, and it should be called before other funcs.
func Start(fn func(id ID) error) error {
	return Call(Wrap(func(id ID, args ...interface{}) error {
		return fn(id)
	}))
}

// Wrap wraps a ID and waits for a startup.
// It is thread-safe, and it should be called before other funcs.
// Call `Call` after `Wrap` to start up a ID.
func Wrap(fn func(id ID, args ...interface{}) error) ID {
	id := uuid.NewV4().String()

	c := &Coroutine{
		id:          id,
		status:      Created,
		inCh:        make(chan []interface{}, 1),
		outCh:       make(chan []interface{}, 1),
		fn:          fn,
		mutexStatus: &sync.Mutex{},
		mutexResume: &sync.Mutex{},
	}
	coroutines.Store(id, c)

	return id
}

// Call launch a ID that is already wrapped.
// It is not thread-safe, and it can only be called beside after Wrap.
// Call `Call` After `Wrap` to start up a ID.
func Call(id ID, args ...interface{}) error {
	c := findCoroutine(id)
	c.writeSyncStatus(Running)

	return func() (err error) {
		defer func() {
			if v := recover(); v != nil {
				err = cast.ToError(v)
			}
		}()
		defer func() {
			coroutines.Delete(id)
		}()

		return c.fn(id, args...)
	}()
}

// Create wraps and yields a ID with no args, waits for a resume.
// It is not thread-safe, and it should be called before other funcs.
// Call `Resume` after `Create` to start up a ID.
func Create(fn func(id ID, inData ...interface{}) error) ID {
	id := Wrap(func(id ID, args ...interface{}) error {
		inData := Yield(id)
		return fn(id, inData...)
	})

	// No error would be caused here.
	go Call(id)

	return id
}

// Resume continues a suspend ID, passing data in and out.
// It is thread-safe, and it can only be called in other Goroutine.
// Call `Resume` after `Create` to start up a ID.
// Call `Resume` after `Yield` to continue a ID.
func Resume(id ID, inData ...interface{}) ([]interface{}, bool) {
	c := findCoroutine(id)

	c.mutexResume.Lock()
	defer c.mutexResume.Unlock()

	if c.readSyncStatus() == Dead {
		return nil, false
	}
	outData := c.resume(inData)

	return outData, true
}

// TryResume likes Resume, but checks status instead of waiting for status.
// It is thread-safe, and it can only be called in other Goroutine.
// Call `TryResume` after `Create` to start up a ID.
// Call `TryResume` after `Yield` to continue a ID.
func TryResume(id ID, inData ...interface{}) ([]interface{}, bool) {
	c := findCoroutine(id)

	c.mutexResume.Lock()
	defer c.mutexResume.Unlock()

	if c.readSyncStatus() != Suspended {
		return nil, false
	}
	outData := c.resume(inData)

	return outData, true
}

// AsyncResume likes Resume, but works async.
// It is thread-safe, and it can only be called in other Goroutine.
// Call `AsyncResume` after `Create` to start up a ID.
// Call `AsyncResume` after `Yield` to continue a ID.
func AsyncResume(id ID, fn func(outData ...interface{}), inData ...interface{}) chan error {
	errCh := make(chan error, 1)

	go func() {
		defer func() {
			if v := recover(); v != nil {
				errCh <- cast.ToError(v)
			}
		}()

		co := findCoroutine(id)
		co.mutexResume.Lock()
		defer co.mutexResume.Unlock()

		if co.readSyncStatus() == Dead {
			panic(fmt.Errorf("coroutine is dead: %s", co))
		}
		outData := co.resume(inData)

		fn(outData...)
	}()

	return errCh
}

// Yield suspends a running coroutine, passing data in and out.
// It is not thread-safe, and it can only be called in coroutine.fn.
// Call `Resume`, `TryResume` or `AsyncResume`
// after `Yield` to continue a ID.
func Yield(id ID, outData ...interface{}) []interface{} {
	c := findCoroutine(id)
	c.writeSyncStatus(Suspended)
	inData := c.yield(outData)
	c.writeSyncStatus(Running)

	return inData
}

// Status shows the status of a ID.
// It is thread-safe, and it can be called in any Goroutine.
// Call `Status` anywhere you need.
func Status(id ID) StatusType {
	v, ok := coroutines.Load(id)
	if !ok {
		return Dead
	}
	c := v.(*Coroutine)

	return c.readSyncStatus()
}

func findCoroutine(id ID) *Coroutine {
	v, ok := coroutines.Load(id)
	if !ok {
		panic(fmt.Errorf("coroutine missing: [%s]", id))
	}

	return v.(*Coroutine)
}

func (c *Coroutine) String() string {
	return fmt.Sprintf("[%s]", c.id)
}

func (c *Coroutine) writeSyncStatus(status StatusType) {
	c.mutexStatus.Lock()
	defer c.mutexStatus.Unlock()
	c.status = status
}

func (c *Coroutine) readSyncStatus() StatusType {
	c.mutexStatus.Lock()
	defer c.mutexStatus.Unlock()

	return c.status
}

func (c *Coroutine) resume(inData []interface{}) []interface{} {
	var outData []interface{}

	select {
	case outData = <-c.outCh:
		break
	case <-time.After(time.Duration(expire) * time.Second):
		panic(fmt.Errorf("coroutine suspended timeout: %v", c))
	}

	select {
	case c.inCh <- inData:
		break
	case <-time.After(time.Duration(expire) * time.Second):
		panic(fmt.Errorf("coroutine suspended timeout: %v", c))
	}

	return outData
}

func (c *Coroutine) yield(outData []interface{}) []interface{} {
	var inData []interface{}

	select {
	case c.outCh <- outData:
		break
	case <-time.After(time.Duration(expire) * time.Second):
		c.writeSyncStatus(Dead)
		panic(fmt.Errorf("coroutine suspended timeout: %v", c))
	}

	select {
	case inData = <-c.inCh:
		break
	case <-time.After(time.Duration(expire) * time.Second):
		c.writeSyncStatus(Dead)
		panic(fmt.Errorf("coroutine suspended timeout: %v", c))
	}

	return inData
}
