package markov

import (
	"bytes"
	log "github.com/Sirupsen/logrus"
	"math/rand"
	"strings"
	"errors"
)

var (
	// ErrEmptySeed is returned if there is no seed set
	ErrEmptySeed = errors.New("Empty seed")
)

// getSeed returns a random seed from the database
func (m *Markov) GetSeed() (string, error) {
	maxRowId, err := m.db.Query("SELECT MAX(rowid) FROM markov_start")
	if err != nil {
		return "", err
	}
	defer maxRowId.Close()

	var maxid int64
	maxRowId.Next()
	maxRowId.Scan(&maxid)

	rowid := rand.Int63n(maxid)
	seedRows, err := m.db.Query("SELECT prefix FROM markov_start WHERE rowid = ?", rowid)
	if err != nil {
		return "", err
	}
	defer seedRows.Close()

	var seed string
	seedRows.Next()
	seedRows.Scan(&seed)
	return seed, nil
}

// getNextToken returns the next token starting with the given prefix
func (m *Markov) getNextToken(prefix string) (string, error) {
	prefixId, err := m.getPrefixId(prefix, false)
	if err != nil {
		return "", err
	}
	if prefixId < 0 {
		return "", nil
	}

	tokenRows, err := m.db.Query("SELECT token FROM markov_next WHERE prefix_rowid = ?", prefixId)
	if err != nil {
		return "", err
	}
	defer tokenRows.Close()

	possibilities := []string{}
	for tokenRows.Next() {
		var token string
		tokenRows.Scan(&token)
		possibilities = append(possibilities, token)
	}
	log.Debug(len(possibilities), " possibilities: ", possibilities)
	i := rand.Int31n(int32(len(possibilities)))
	return possibilities[i], nil
}

// Generate generates a new text with the given length starting with seed.
// If seed is empty, it will pick a random one from the db.
func (m *Markov) Generate(seed string, len int) (string, error) {
	if seed == "" {
		return "", ErrEmptySeed
	}

	buffer := bytes.NewBufferString(seed)
	count := 0
	window := strings.Split(seed, " ")

	for {
		log.Debug(window)
		next, err := m.getNextToken(strings.Join(window, " "))
		if err != nil {
			return "", err
		} else if next == "" {
			log.Debug("End of chain")
			break
		}

		buffer.WriteString(" ")
		buffer.WriteString(next)

		window = append(window[1:m.params.PrefixLen], next)
		count++
		if count > len {
			log.Debug("Length reached")
			break
		}
	}

	return buffer.String(), nil
}
