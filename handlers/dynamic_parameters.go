package handlers

import (
	"fmt"

	"github.com/djatwood/formailer"
)

/*
ReplaceDynamic replaces a dynamic value with submission data

This is a {{dynamic}} value => This is a replaced value
*/
func ReplaceDynamic(str string, submission *formailer.Submission) string {
	openIndex := -1
	for i := range str[1:] {
		switch str[i : i+2] {
		case "{{":
			openIndex = i
		case "}}":
			if openIndex < 0 {
				continue
			}

			if v, ok := submission.Values[str[openIndex+2:i]]; ok {
				str = str[:openIndex] + fmt.Sprint(v) + str[i+2:]
			}
		}
	}

	return str
}
