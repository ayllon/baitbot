package markov

import (
	log "github.com/Sirupsen/logrus"
	"regexp"
	"strings"
)

var (
	splitRegex = regexp.MustCompile("\\s+")
)

// ProcessText extracts prefixes => next from the given text
func (m *Markov) ProcessText(text string) error {
	prefixes := make(map[string][]string)

	words := splitRegex.Split(text, -1)
	nwords := len(words)

	for i := 0; i+m.params.PrefixLen < nwords; i++ {
		end := i + m.params.PrefixLen
		prefix := strings.Join(words[i:end], " ")
		next := words[end]

		if _, ok := prefixes[prefix]; !ok {
			prefixes[prefix] = make([]string, 0, 1)
		}

		prefixes[prefix] = append(prefixes[prefix], next)
	}

	m.dbMutex.Lock()
	defer m.dbMutex.Unlock()

	for prefix, tokens := range prefixes {
		for _, token := range tokens {
			log.Debug(prefix, " => ", token)
			if !m.params.DryRun {
				if err := m.insertEntry(prefix, token); err != nil {
					return err
				}
			}
		}
	}

	if !m.params.DryRun && len(words) > m.params.PrefixLen {
		return m.insertStart(strings.Join(words[0:m.params.PrefixLen], " "))
	}
	return nil
}
