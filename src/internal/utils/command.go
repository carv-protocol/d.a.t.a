package utils

import (
	"log"
	"strings"
)

// ParseCommandArgs parses a command string into arguments, properly handling quoted strings
func ParseCommandArgs(content string) []string {
	// Add debug logging
	log.Printf("Parsing command: %s", content)

	// First split the command name
	parts := strings.Fields(content)
	if len(parts) == 0 {
		return nil
	}

	// Remove the command name (including the /)
	content = strings.TrimSpace(strings.TrimPrefix(content, parts[0]))

	var args []string
	var currentArg strings.Builder
	inQuotes := false

	// Parse the remaining arguments
	for _, char := range content {
		switch char {
		case '"':
			// If we're starting a quote, skip adding the quote character
			if !inQuotes {
				// If we have content before the quote, add it as a separate arg
				if currentArg.Len() > 0 {
					args = append(args, strings.TrimSpace(currentArg.String()))
					currentArg.Reset()
				}
				inQuotes = true
			} else {
				// End of quote, add the accumulated arg
				if currentArg.Len() > 0 {
					args = append(args, currentArg.String())
					currentArg.Reset()
				}
				inQuotes = false
			}
		case ' ':
			if !inQuotes {
				// If we have accumulated content, add it as an arg
				if currentArg.Len() > 0 {
					args = append(args, strings.TrimSpace(currentArg.String()))
					currentArg.Reset()
				}
			} else {
				currentArg.WriteRune(char)
			}
		default:
			currentArg.WriteRune(char)
		}
	}

	// Add any remaining content
	if currentArg.Len() > 0 {
		args = append(args, strings.TrimSpace(currentArg.String()))
	}

	// Remove any empty arguments
	var cleanArgs []string
	for _, arg := range args {
		if arg != "" {
			cleanArgs = append(cleanArgs, arg)
		}
	}

	// Add debug logging
	log.Printf("Parsed arguments: %#v", cleanArgs)

	return cleanArgs
}
