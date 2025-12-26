package monitor

import (
	"regexp"
	"strings"
)

type Matcher struct {
	useRegex bool
	phrases  []string
	stop     []string
	regexps  []*regexp.Regexp
}

func NewMatcher(phrases []string, stopWords []string, useRegex bool) *Matcher {
	m := &Matcher{useRegex: useRegex, phrases: phrases}
	for _, sw := range stopWords {
		sw = strings.TrimSpace(sw)
		if sw != "" {
			m.stop = append(m.stop, strings.ToLower(sw))
		}
	}
	if useRegex {
		for _, p := range phrases {
			if re, err := regexp.Compile("(?i)" + p); err == nil {
				m.regexps = append(m.regexps, re)
			}
		}
	}
	return m
}

func (m *Matcher) Match(text string) bool {
	if text == "" {
		return false
	}
	tLower := strings.ToLower(text)
	for _, sw := range m.stop {
		if sw != "" && strings.Contains(tLower, sw) {
			return false
		}
	}
	if m.useRegex {
		for _, re := range m.regexps {
			if re.MatchString(text) {
				return true
			}
		}
		return false
	}
	for _, p := range m.phrases {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		if strings.Contains(tLower, strings.ToLower(p)) {
			return true
		}
	}
	return false
}


