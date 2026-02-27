package notice

import "strings"

func wrap(text string, width int) []string {
	var lines []string
	words := strings.Fields(text)

	line := ""
	for _, w := range words {
		if len(line)+len(w)+1 > width {
			lines = append(lines, line)
			line = w
		} else {
			if line == "" {
				line = w
			} else {
				line += " " + w
			}
		}
	}

	if line != "" {
		lines = append(lines, line)
	}

	return lines
}
