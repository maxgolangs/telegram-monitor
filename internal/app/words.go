package app

import (
	"os"
	"strings"
)

func mustReadWords(keywordsFile, stopFile string) (keywords []string, stopwords []string) {
	keywords = readLines(keywordsFile)
	stopwords = readLines(stopFile)
	return
}

func readLines(path string) []string {
	path = strings.TrimSpace(path)
	if path == "" {
		return nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	var out []string
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		out = append(out, line)
	}
	return out
}



