package client

import (
	"reflect"
	"strings"
	"testing"
)

func checkFlattenResult(t *testing.T, tree map[string]any, expected [][]string) {
	flat, err := Flatten(tree)
	if err != nil {
		t.Errorf("unexpected error: '%s'", err.Error())
	} else if !reflect.DeepEqual(flat, expected) {
		t.Errorf("unexpected result: %v, expected: %v", flat, expected)
	}
}
func checkFlattenError(t *testing.T, tree map[string]any, substr string) {
	flat, err := Flatten(tree)
	if err == nil {
		t.Errorf("unexpected result: %v, expected error: '%s'", flat, substr)
	} else if !strings.Contains(err.Error(), substr) {
		t.Errorf("unexpected error: '%s', expected '%s'", err.Error(), substr)
	}
}

func TestFlatten_String(t *testing.T) {
	checkFlattenResult(t,
		map[string]any{
			"foo": "bar",
		},
		[][]string{
			{"foo", "bar"},
		},
	)
}

func TestFlatten_Map(t *testing.T) {
	checkFlattenResult(t,
		map[string]any{
			"foo": map[string]any{
				"bar": "baz",
			},
		},
		[][]string{
			{"foo bar", "baz"},
		},
	)
}
func TestFlatten_MapNesting(t *testing.T) {
	checkFlattenResult(t,
		map[string]any{
			"foo": map[string]any{
				"bar": map[string]any{
					"baz": map[string]any{
						"qux": "quo",
					},
				},
			},
		},
		[][]string{
			{"foo bar baz qux", "quo"},
		},
	)
}
func TestFlatten_MapEmpty(t *testing.T) {
	checkFlattenResult(t,
		map[string]any{
			"foo": map[string]any{},
		},
		[][]string{
			{"foo", ""},
		},
	)
}
func TestFlatten_MapErrorWithAny(t *testing.T) {
	checkFlattenError(t,
		map[string]any{
			"foo": map[any]any{},
		},
		"foo: invalid type",
	)
}

func TestFlatten_Array(t *testing.T) {
	checkFlattenResult(t,
		map[string]any{
			"test": []string{
				"foo",
				"bar",
				"baz",
			},
		},
		[][]string{
			{"test", "foo"},
			{"test", "bar"},
			{"test", "baz"},
		},
	)
}
func TestFlatten_ArrayEmpty(t *testing.T) {
	checkFlattenResult(t,
		map[string]any{
			"foo": []string{},
		},
		[][]string{},
	)
}
func TestFlatten_ArrayErrorWithAny(t *testing.T) {
	checkFlattenError(t,
		map[string]any{
			"foo": []any{},
		},
		"foo: invalid type",
	)
}
