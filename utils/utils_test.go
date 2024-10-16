package utils

import (
	"testing"
)

func TestClonePointer(t *testing.T) {
	t.Run("Clone *int", func(t *testing.T) {
		n := 123
		pointedN := &n
		clonedN := ClonePointer(pointedN)
		if clonedN == pointedN {
			t.Fatalf("Pointers are equal")
		}
		if *clonedN != *pointedN {
			t.Fatalf("Values are not equal")
		}
	})

	t.Run("Clone *string", func(t *testing.T) {
		s := "abc"
		pointedS := &s
		clonedS := ClonePointer(pointedS)
		if clonedS == pointedS {
			t.Fatalf("Pointers are equal")
		}
		if *clonedS != *pointedS {
			t.Fatalf("Values are not equal")
		}
	})

	t.Run("Clone *struct", func(t *testing.T) {
		m := struct {
			field int
		}{
			field: 123,
		}
		pointedM := &m
		clonedM := ClonePointer(pointedM)
		if clonedM == pointedM {
			t.Fatalf("Pointers are equal")
		}
		if *clonedM != *pointedM {
			t.Fatalf("Values are not equal")
		}
	})

	t.Run("Clone nil pointer", func(t *testing.T) {
		pointedM := struct {
			str *string
		}{}
		clonedStr := ClonePointer(pointedM.str)
		if clonedStr != nil {
			t.Fatalf("Cloned value should be nil")
		}
	})
}

func TestContainsDuplicates(t *testing.T) {
	tests := []struct {
		name string
		src  []int
		exp  bool
	}{
		{
			name: "no contains duplicates",
			src:  []int{1, 3, 4, 2, 5, 9, 7, 0},
		},
		{
			name: "contains duplicates",
			src:  []int{1, 3, 4, 2, 5, 9, 7, 9},
			exp:  true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ContainsDuplicates(tt.src); got != tt.exp {
				t.Fatalf("Test failed, expected %v, got %v", tt.exp, got)
			}
		})
	}
}
