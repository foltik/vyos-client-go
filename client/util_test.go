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

func TestUnit_Flatten_String(t *testing.T) {
	checkFlattenResult(t,
		map[string]any{
			"foo": "bar",
		},
		[][]string{
			{"foo", "bar"},
		},
	)
}

func TestUnit_Flatten_Map(t *testing.T) {
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
func TestUnit_Flatten_MapString(t *testing.T) {
	checkFlattenResult(t,
		map[string]any{
			"foo": map[string]string{
				"bar": "baz",
			},
		},
		[][]string{
			{"foo bar", "baz"},
		},
	)
}
func TestUnit_Flatten_MapNesting(t *testing.T) {
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
func TestUnit_Flatten_MapEmpty(t *testing.T) {
	checkFlattenResult(t,
		map[string]any{
			"foo": map[string]any{},
		},
		[][]string{
			{"foo", ""},
		},
	)
}
func TestUnit_Flatten_MapErrorWithAny(t *testing.T) {
	checkFlattenError(t,
		map[string]any{
			"foo": map[any]any{},
		},
		"foo: invalid type",
	)
}

func TestUnit_Flatten_Array(t *testing.T) {
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
func TestUnit_Flatten_ArrayMixed(t *testing.T) {
	checkFlattenResult(t,
		map[string]any{
			"test": []any{
				"foo",
				map[string]any{
					"bar": "baz",
				},
			},
		},
		[][]string{
			{"test", "foo"},
			{"test bar", "baz"},
		},
	)
}
func TestUnit_Flatten_ArrayEmpty(t *testing.T) {
	checkFlattenResult(t,
		map[string]any{
			"foo": []string{},
		},
		[][]string{},
	)
}
