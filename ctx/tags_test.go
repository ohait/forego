package ctx_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/ohait/forego/ctx"
	"github.com/ohait/forego/test"
)

func TestTags(t *testing.T) {
	var c ctx.C
	c = context.Background()

	fetch := func(c ctx.C) []any {
		t.Helper()
		var list []any
		err := ctx.RangeTag(c, func(key string, j ctx.JSON) error {
			t.Logf("tag[%s] = %s", key, string(j))
			var v any
			err := json.Unmarshal(j, &v)
			if err != nil {
				return err
			}
			list = append(list, v)
			return nil
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		return list
	}

	c = ctx.WithTag(c, "a", "one")
	{
		list := fetch(c)
		test.EqualsJSON(t, []any{"one"}, list)
	}

	c = ctx.WithTag(c, "b", "two")
	{
		list := fetch(c)
		test.EqualsJSON(t, []any{"one", "two"}, list)
	}

	c = ctx.WithTag(c, "a", "typo")
	{
		list := fetch(c)
		test.EqualsJSON(t, []any{"one", "two", "typo"}, list) // NOTE(oha): we range over all the assignments, no check for duplications
	}
}
