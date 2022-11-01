package coroutine_test

import (
	"strings"
	"testing"

	"github.com/cloudlibraries/cast"
	"github.com/cloudlibraries/coroutine"
)

func TestCreate(t *testing.T) {
	exit := make(chan int)
	co := coroutine.Create(func(co coroutine.ID, args ...interface{}) error {
		strs := cast.ToStringSlice(args)
		out := strings.Join(strs, " ") // coroutine resume 1
		if out != "ID resume 1" {
			t.Error("ID flow error, should be ID resume 1")
		}
		t.Log(out)

		inData := coroutine.Yield(co, "ID", "yield", "2")
		strs = cast.ToStringSlice(inData)
		out = strings.Join(strs, " ") // coroutine resume 3
		if out != "ID resume 3" {
			t.Error("ID flow error, should be ID resume 3")
		}
		t.Log(out)

		_ = coroutine.Yield(co, "ID", "yield", "4")
		return nil
	})

	_, ok := coroutine.TryResume(co, "ID", "resume", "0")
	if !ok {
		t.Log("Try resume test result Correct")
	}

	_, ok = coroutine.Resume(co, "ID", "resume", "1")
	if !ok {
		t.Log("Dead ID")
	}

	outData, ok := coroutine.Resume(co, "ID", "resume", "3")
	if !ok {
		t.Log("Dead ID")
	}
	strs := cast.ToStringSlice(outData)
	out := strings.Join(strs, " ") // coroutine yield 2
	if out != "ID yield 2" {
		t.Error("ID flow error, should be ID yield 2")
	}
	t.Log(out)

	coroutine.AsyncResume(co, func(outData ...interface{}) {
		strs = cast.ToStringSlice(outData)
		out := strings.Join(strs, " ") // coroutine yield 4
		if out != "ID yield 4" {
			t.Error("ID flow error, should be ID yield 4")
		}
		t.Log(out)

		exit <- 1
	}, "ID", "resume", "3")

	<-exit
}

func TestStart(t *testing.T) {
	co := coroutine.Wrap(
		func(co coroutine.ID, args ...interface{}) error {
			strs := cast.ToStringSlice(args)
			out := strings.Join(strs, " ") // ID call 1
			if out != "ID call 1" {
				t.Error("ID flow error, should be ID call 1")
			}
			t.Log(out)

			inData := coroutine.Yield(co, "ID", "yield", "2")
			strs = cast.ToStringSlice(inData)
			out = strings.Join(strs, " ") // ID resume 3
			if out != "ID resume 3" {
				t.Error("ID flow error, should be ID resume 3")
			}
			t.Log(out)

			_ = coroutine.Yield(co, "ID", "yield", "4")
			return nil
		})

	go func() {
		if err := coroutine.Call(co, "ID", "call", "1"); err != nil {
			t.Error(err)
		}
	}()

	outData, ok := coroutine.Resume(co, "ID", "resume", "3")
	if !ok {
		t.Log("Dead ID")
	}
	strs := cast.ToStringSlice(outData)
	out := strings.Join(strs, " ") // ID yield 2
	if out != "ID yield 2" {
		t.Error("ID flow error, should be ID yield 2")
	}
	t.Log(out)

	outData, ok = coroutine.Resume(co, "ID", "resume", "5")
	if !ok {
		t.Log("Dead ID")
	}
	strs = cast.ToStringSlice(outData)
	out = strings.Join(strs, " ") // ID yield 4
	if out != "ID yield 4" {
		t.Error("ID flow error, should be ID yield 4")
	}
	t.Log(out)
}
