package mapset

import (
	"testing"

	"github.com/NordSecurity/nordvpn-linux/test/category"
	"github.com/stretchr/testify/assert"
)

func TestNewSet(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name string
		run  func(t *testing.T)
	}{
		{
			name: "int set is empty and not nil",
			run: func(t *testing.T) {
				s := NewSet[int]()
				assert.NotNil(t, s, "NewSet should not return nil")

				rawMap := s.(set[int])
				assert.Empty(t, rawMap, "expected empty set")
			},
		},
		{
			name: "string set is empty and not nil",
			run: func(t *testing.T) {
				s := NewSet[string]()
				assert.NotNil(t, s, "NewSet should not return nil")

				rawMap := s.(set[string])
				assert.Empty(t, rawMap, "expected empty set")
			},
		},
		{
			name: "set initialized with elements contains the elements",
			run: func(t *testing.T) {
				elements := []string{"a", "b", "c"}
				s := NewSet(elements...)

				rawMap := s.(set[string])
				assert.Len(t, rawMap, len(elements), "expected 3 element set")

				for _, element := range elements {
					assert.True(t, s.Contains(element), "expected element %s not found in the set", element)
				}
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, test.run)
	}
}

func TestAddInt(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name    string
		input   []int
		wantLen int
	}{
		{name: "single element", input: []int{42}, wantLen: 1},
		{name: "multiple distinct elements", input: []int{1, 2, 3}, wantLen: 3},
		{name: "duplicate elements", input: []int{42, 42, 42}, wantLen: 1},
		{name: "mixed distinct and duplicate", input: []int{1, 1, 2, 3, 3, 3}, wantLen: 3},
		{name: "no elements", input: []int{}, wantLen: 0},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			s := NewSet[int]()
			for _, e := range test.input {
				s.Add(e)
			}

			rawMap := s.(set[int])
			assert.Len(t, rawMap, test.wantLen)
		})
	}
}

func TestAddString(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name    string
		input   []string
		wantLen int
	}{
		{name: "distinct strings", input: []string{"hello", "world"}, wantLen: 2},
		{name: "duplicate strings", input: []string{"a", "a", "b"}, wantLen: 2},
		{name: "empty string counts", input: []string{"", ""}, wantLen: 1},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			s := NewSet[string]()
			for _, e := range test.input {
				s.Add(e)
			}

			rawMap := s.(set[string])
			assert.Len(t, rawMap, test.wantLen)
		})
	}
}

func TestContainsInt(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name  string
		added []int
		query int
		want  bool
	}{
		{name: "present element", added: []int{42}, query: 42, want: true},
		{name: "absent element", added: []int{42}, query: 99, want: false},
		{name: "empty set", added: []int{}, query: 1, want: false},
		{name: "one of many present", added: []int{10, 20, 30}, query: 20, want: true},
		{name: "none of many match", added: []int{10, 20, 30}, query: 99, want: false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			s := NewSet[int]()
			for _, e := range test.added {
				s.Add(e)
			}
			assert.Equal(t, test.want, s.Contains(test.query))
		})
	}
}

func TestContainsString(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name  string
		added []string
		query string
		want  bool
	}{
		{name: "present string", added: []string{"apple", "banana"}, query: "apple", want: true},
		{name: "another present string", added: []string{"apple", "banana"}, query: "banana", want: true},
		{name: "absent string", added: []string{"apple", "banana"}, query: "orange", want: false},
		{name: "empty set", added: []string{}, query: "apple", want: false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			s := NewSet[string]()
			for _, e := range test.added {
				s.Add(e)
			}
			assert.Equal(t, test.want, s.Contains(test.query))
		})
	}
}

func TestEqual(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name string
		a    []int
		b    []int
		want bool
	}{
		{name: "both empty", a: []int{}, b: []int{}, want: true},
		{name: "identical single element", a: []int{1}, b: []int{1}, want: true},
		{name: "identical multiple elements", a: []int{1, 2, 3}, b: []int{1, 2, 3}, want: true},
		{name: "same elements different order", a: []int{3, 1, 2}, b: []int{1, 2, 3}, want: true},
		{name: "same set with duplicate inputs", a: []int{1, 1, 2}, b: []int{2, 1}, want: true},
		{name: "different lengths", a: []int{1, 2}, b: []int{1, 2, 3}, want: false},
		{name: "same length different elements", a: []int{1, 2, 3}, b: []int{1, 2, 4}, want: false},
		{name: "one empty one non-empty", a: []int{}, b: []int{1}, want: false},
		{name: "disjoint sets", a: []int{1, 2}, b: []int{3, 4}, want: false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			a := NewSet(test.a...)
			b := NewSet(test.b...)
			assert.Equal(t, test.want, a.Equal(b))
		})
	}
}

func TestEqualSymmetry(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name string
		a    []int
		b    []int
	}{
		{name: "equal sets", a: []int{1, 2, 3}, b: []int{3, 2, 1}},
		{name: "unequal same length", a: []int{1, 2, 3}, b: []int{1, 2, 4}},
		{name: "unequal different length", a: []int{1}, b: []int{1, 2}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			a := NewSet(test.a...)
			b := NewSet(test.b...)
			assert.Equal(t, a.Equal(b), b.Equal(a), "Equal is not symmetric")
		})
	}
}

func TestEqualReflexive(t *testing.T) {
	category.Set(t, category.Unit)

	s := NewSet(1, 2, 3)
	assert.True(t, s.Equal(s), "a set should be equal to itself")
}

func TestEqualStrings(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name string
		a    []string
		b    []string
		want bool
	}{
		{name: "equal string sets", a: []string{"a", "b"}, b: []string{"b", "a"}, want: true},
		{name: "unequal string sets", a: []string{"a", "b"}, b: []string{"a", "c"}, want: false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			a := NewSet(test.a...)
			b := NewSet(test.b...)
			assert.Equal(t, test.want, a.Equal(b))
		})
	}
}

func TestIsSubset(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name string
		a    []string
		b    []string
		want bool
	}{
		{
			name: "b is a subset of a",
			a:    []string{"a", "b", "c", "d"},
			b:    []string{"c", "a"},
			want: true,
		},
		{
			name: "b is not a subset of a",
			a:    []string{"a", "b", "c", "d"},
			b:    []string{"e", "f"},
			want: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			a := NewSet(test.a...)
			b := NewSet(test.b...)
			assert.Equal(t, test.want, b.IsSubset(a))
		})
	}
}

func TestIsEmpty(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name string
		set  []string
		want bool
	}{
		{
			name: "non-empty set",
			set:  []string{"a", "b", "c", "d"},
			want: false,
		},
		{
			name: "empty set",
			set:  []string{},
			want: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			set := NewSet(test.set...)
			assert.Equal(t, test.want, set.IsEmpty())
		})
	}
}

func TestIsLength(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name string
		set  []string
		want int
	}{
		{
			name: "non-empty set",
			set:  []string{"a", "b", "c", "d"},
			want: 4,
		},
		{
			name: "empty set",
			set:  []string{},
			want: 0,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			set := NewSet(test.set...)
			assert.Equal(t, test.want, set.Length())
		})
	}
}

func TestString(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name string
		run  func(t *testing.T)
	}{
		{
			name: "set of strings",
			run: func(t *testing.T) {
				s := NewSet("a", "b", "c")
				setString := s.String()
				// check for each individual element, as map doesn't have any ordering
				assert.Contains(t, setString, "a")
				assert.Contains(t, setString, "b")
				assert.Contains(t, setString, "c")
				assert.Len(t, setString, len("[a b c]"))
			},
		},
		{
			name: "set of ints",
			run: func(t *testing.T) {
				s := NewSet(1, 2, 3)
				setString := s.String()
				// check for each individual element, as map  doesn't have any ordering
				assert.Contains(t, setString, "1")
				assert.Contains(t, setString, "2")
				assert.Contains(t, setString, "3")
				assert.Len(t, setString, len("[1 2 3]"))
			},
		},
		{
			name: "empty set",
			run: func(t *testing.T) {
				s := NewSet[int]()
				assert.Equal(t, "[]", s.String())
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, test.run)
	}
}
