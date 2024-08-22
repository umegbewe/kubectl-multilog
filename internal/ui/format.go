package ui

import (
	"strings"
)

func (t *LogExplorerTUI) FormatLogs(logs string) string {
	logLines := strings.Split(logs, "\n")
	formattedLogs := make([]string, 0, len(logLines))

	levelColors := map[string]string{
		"ERROR": "[red]",
		"WARN":  "[yellow]",
		"INFO":  "[green]",
		"DEBUG": "[blue]",
	}

	for i := len(logLines) -1; i >=0; i-- {
		line := logLines[i]
		if line == "" {
			continue
		}

		formattedLine := line
		for level, color := range levelColors {
			if strings.Contains(strings.ToUpper(line), level) {
				formattedLine = strings.ReplaceAll(line, level, color+level+"[-]")
				formattedLine = strings.ReplaceAll(formattedLine, strings.ToLower(level), color+strings.ToLower(level)+"[-]")
				formattedLine = strings.ReplaceAll(formattedLine, strings.Title(strings.ToLower(level)), color+strings.Title(strings.ToLower(level))+"[-]")
				break
			}
		}

		// Append the (possibly) formatted line to the result
		formattedLogs = append(formattedLogs, formattedLine)
	}

	return strings.Join(formattedLogs, "\n")
}
