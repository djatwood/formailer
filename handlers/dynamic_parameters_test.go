package handlers

import (
	"fmt"
	"testing"

	"github.com/djatwood/formailer"
)

func TestReplace(t *testing.T) {
	submission := formailer.Submission{
		Values: map[string]interface{}{
			"name": "First Last",
			"map": map[string]interface{}{
				"kind": "map",
			},
		},
	}

	tests := map[string]interface{}{
		"Prefix {{name}} Suffix": "Prefix First Last Suffix",
		"{{map}}":                submission.Values["map"],
		"name}}":                 nil,
		"{{name":                 nil,
		"{{notexist}}":           nil,
	}

	for test, expected := range tests {
		if expected == nil {
			expected = test
		}
		result := ReplaceDynamic(test, &submission)
		if result != fmt.Sprint(expected) {
			t.Errorf("unexpected result from replace\nExpected: %s\nGot: %s", expected, result)
		}
	}
}
