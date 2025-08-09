package main

import (
	"reflect"
	"testing"
)


func TestCleanInput(t *testing.T) {
	cases := []struct {
		input 	 string
		expected []string
	}{
		{
			input: "  hello world  ",
			expected: []string {"hello", "world"},
		},
		{
			input:    "Charmander Bulbasaur PIKACHU",
			expected: []string{"charmander", "bulbasaur", "pikachu"},
		},
		{
			input:    "\t   Spaces\tand\nNewlines \t ",
			expected: []string{"spaces", "and", "newlines"},
		},
	}

	for _, c := range cases {
		actual := cleanInput(c.input)

		// Check if both slices are exactly the same
		if !reflect.DeepEqual(actual, c.expected) {
			t.Errorf("cleanInput(%q) == %v, expected %v", c.input, actual, c.expected)
		}
	}
}