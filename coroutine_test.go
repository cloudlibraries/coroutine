package coroutine_test

import (
	"testing"

	"github.com/cloudlibraries/coroutine"
	"github.com/cloudlibraries/safe"
	qt "github.com/frankban/quicktest"
)

func TestCreate(t *testing.T) {
	c := qt.New(t)

	c.Run("Create", func(c *qt.C) {
		co, err := coroutine.Create(func(co *coroutine.Coroutine, args ...any) error {
			output, err := co.Yield("Hello")
			if err != nil {
				return err
			}
			c.Assert(output, qt.DeepEquals, []any{"World"})

			return nil
		})
		c.Assert(err, qt.IsNil)

		output, err := co.Resume()
		c.Assert(err, qt.IsNil)
		c.Assert(output, qt.DeepEquals, []any{})

		output, err = co.Resume("World")

		c.Assert(err, qt.IsNil)
		c.Assert(output, qt.DeepEquals, "Hello")
	})
}

func TestStart(t *testing.T) {
	c := qt.New(t)

	c.Run("Start", func(c *qt.C) {
		err := coroutine.Start(func(co *coroutine.Coroutine, args ...any) error {
			go safe.Do(func() error {
				output, err := co.Yield("Hello")
				if err != nil {
					return err
				}
				c.Assert(output, qt.DeepEquals, []any{"World"})

				return nil
			})

			output, err := co.Resume("World")

			c.Assert(err, qt.IsNil)
			c.Assert(output, qt.DeepEquals, "Hello")
			return nil
		})

		c.Assert(err, qt.IsNil)
	})
}
