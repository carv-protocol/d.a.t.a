package commands

import "strings"

// parseCommandArgs parses a command string into arguments, properly handling quoted strings
func ParseCommandArgs(content string) []string {
	var args []string
	var currentArg strings.Builder
	inQuotes := false

	// Skip the first character (/)
	for _, char := range content[1:] {
		switch char {
		case '"':
			inQuotes = !inQuotes
		case ' ':
			if !inQuotes {
				if currentArg.Len() > 0 {
					args = append(args, currentArg.String())
					currentArg.Reset()
				}
			} else {
				currentArg.WriteRune(char)
			}
		default:
			currentArg.WriteRune(char)
		}
	}

	if currentArg.Len() > 0 {
		args = append(args, currentArg.String())
	}

	return args
}
