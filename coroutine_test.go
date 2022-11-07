package coroutine_test

import (
	"testing"

	"github.com/frankban/quicktest"
	"github.com/golibraries/coroutine"
	"github.com/golibraries/safe"
)

func TestCreate(t *testing.T) {
	c := quicktest.New(t)

	c.Run("Create", func(c *quicktest.C) {
		co, err := coroutine.Create(func(co *coroutine.Coroutine, args ...any) error {
			output, err := co.Yield("Hello")
			if err != nil {
				return err
			}
			c.Assert(output, quicktest.DeepEquals, []any{"World"})

			return nil
		})
		c.Assert(err, quicktest.IsNil)

		output, err := co.Resume()
		c.Assert(err, quicktest.IsNil)
		c.Assert(output, quicktest.DeepEquals, []any{})

		output, err = co.Resume("World")

		c.Assert(err, quicktest.IsNil)
		c.Assert(output, quicktest.DeepEquals, "Hello")
	})
}

func TestStart(t *testing.T) {
	c := quicktest.New(t)

	c.Run("Start", func(c *quicktest.C) {
		err := coroutine.Start(func(co *coroutine.Coroutine, args ...any) error {
			go safe.Do(func() error {
				output, err := co.Yield("Hello")
				if err != nil {
					return err
				}
				c.Assert(output, quicktest.DeepEquals, []any{"World"})

				return nil
			})

			output, err := co.Resume("World")

			c.Assert(err, quicktest.IsNil)
			c.Assert(output, quicktest.DeepEquals, "Hello")
			return nil
		})

		c.Assert(err, quicktest.IsNil)
	})
}
