package analyzer

import (
	"reflect"
	"sort"
	"testing"
)

func TestExpandLanguages(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		expected []string
	}{
		{
			name:     "No expansion needed",
			input:    []string{"Go", "Python"},
			expected: []string{"Go", "Python"},
		},
		{
			name:     "Expand TypeScript to JavaScript",
			input:    []string{"TypeScript"},
			expected: []string{"TypeScript", "JavaScript"},
		},
		{
			name:     "Expand TSX to JavaScript",
			input:    []string{"TSX"},
			expected: []string{"TSX", "JavaScript"},
		},
		{
			name:     "Expand Vue to JavaScript",
			input:    []string{"Vue"},
			expected: []string{"Vue", "JavaScript"},
		},
		{
			name:     "Expand SCSS to CSS",
			input:    []string{"SCSS"},
			expected: []string{"SCSS", "CSS"},
		},
		{
			name:     "Expand Less to CSS",
			input:    []string{"Less"},
			expected: []string{"Less", "CSS"},
		},
		{
			name:     "Expand Kotlin to Java",
			input:    []string{"Kotlin"},
			expected: []string{"Kotlin", "Java"},
		},
		{
			name:     "Expand C++ to C",
			input:    []string{"C++"},
			expected: []string{"C++", "C"},
		},
		{
			name:     "Mixed expansion",
			input:    []string{"TypeScript", "SCSS", "Kotlin"},
			expected: []string{"TypeScript", "SCSS", "Kotlin", "JavaScript", "CSS", "Java"},
		},
		{
			name:     "Already present",
			input:    []string{"TypeScript", "JavaScript"},
			expected: []string{"TypeScript", "JavaScript"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := expandLanguages(tt.input)

			// Sort for comparison
			sort.Strings(got)
			sort.Strings(tt.expected)

			if !reflect.DeepEqual(got, tt.expected) {
				t.Errorf("expandLanguages() = %v, want %v", got, tt.expected)
			}
		})
	}
}
