package markov

import (
	"testing"
)

var (
	db     = "/tmp/markov-test.db"
	params = Parameters{PrefixLen: 3}
)

func getMarkov() *Markov {
	m, err := New(db, params)
	if err != nil {
		panic(err)
	}
	if err = m.Clear(); err != nil {
		panic(err)
	}
	return m
}

// Train with a simple test, start with the same prefix, should get it back
func TestSimple(t *testing.T) {
	markov := getMarkov()
	defer markov.Close()

	input := "This is a simple sample text"
	if err := markov.ProcessText(input); err != nil {
		t.Fatal(err)
	}

	output, err := markov.Generate("This is a", 6)
	if err != nil {
		t.Fatal(err)
	}

	if output != input {
		t.Error("Expected '", input, "' got '", output, "'")
	}
}
