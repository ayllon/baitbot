package main

import (
	"bufio"
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/Sirupsen/logrus"
	_ "github.com/mattn/go-sqlite3"
	"github.com/spf13/cobra"
	"io"
	"math/rand"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"
)

var (
	dbPath     string
	prefixLen  int
	messageLen int
	db         *sql.DB
	dbMutex    sync.Mutex

	splitRegex = regexp.MustCompile("\\s")
)

func createSchema() {
	_, err := db.Exec(`
CREATE TABLE IF NOT EXISTS markov_prefix (prefix TEXT)`)
	if err != nil {
		logrus.Panic(err)
	}
	_, err = db.Exec(`
CREATE UNIQUE INDEX IF NOT EXISTS markov_prefix_index ON markov_prefix(prefix)`)
	if err != nil {
		logrus.Panic(err)
	}

	_, err = db.Exec(`
CREATE TABLE IF NOT EXISTS markov_next
	(prefix_rowid INTEGER NOT NULL, token TEXT,
	 FOREIGN KEY (prefix_rowid) REFERENCES markov_prefix(rowid))`)
	if err != nil {
		logrus.Panic(err)
	}

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS markov_start (prefix TEXT)`)
	if err != nil {
		logrus.Panic(err)
	}
}

var MarkovCmd = &cobra.Command{
	Use: "markov",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		RootCmd.PersistentPreRun(cmd, args)

		var err error
		logrus.Debug("Opening sqlite ", dbPath)
		db, err = sql.Open("sqlite3", dbPath)
		if err != nil {
			logrus.Fatal(err)
		}
		createSchema()
	},
	PersistentPostRun: func(cmd *cobra.Command, args []string) {
		db.Close()
	},
}

var MarkovClearCmd = &cobra.Command{
	Use: "clear",
	Run: func(cmd *cobra.Command, args []string) {
		logrus.Info("Clearing the database")
		if _, err := db.Exec("DELETE FROM markov"); err != nil {
			logrus.Fatal(err)
		}
	},
}

func getPrefixId(tr *sql.Tx, prefix string) int64 {
	rows, err := tr.Query("SELECT rowid FROM markov_prefix WHERE prefix = ?", prefix)
	if err != nil {
		logrus.Panic(err)
	}
	defer rows.Close()
	var rowid int64
	if rows.Next() {
		err = rows.Scan(&rowid)
		if err != nil {
			logrus.Panic(err)
		}
	} else {
		result, err := tr.Exec("INSERT INTO markov_prefix (prefix) VALUES (?)", prefix)
		if err != nil {
			logrus.Panic(err)
		}
		rowid, err = result.LastInsertId()
		if err != nil {
			logrus.Panic(err)
		}
	}
	return rowid
}

func insertEntry(prefix string, next string) {
	dbMutex.Lock()
	defer dbMutex.Unlock()

	tr, err := db.Begin()
	if err != nil {
		logrus.Panic(err)
	}
	defer tr.Commit()

	prefixId := getPrefixId(tr, prefix)
	_, err = tr.Exec("INSERT INTO markov_next (prefix_rowid, token) VALUES (?, ?)", prefixId, next)
	if err != nil {
		logrus.Panic(err)
	}
	logrus.Debug(prefixId, " => ", next)
}

func insertStart(prefix string) {
	dbMutex.Lock()
	defer dbMutex.Unlock()

	_, err := db.Exec("INSERT INTO markov_start (prefix) VALUES (?)", prefix)
	if err != nil {
		logrus.Panic(err)
	}
}

func processText(text string) {
	words := splitRegex.Split(text, -1)
	nwords := len(words)
	for i := 0; i+prefixLen < nwords; i++ {
		prefix := words[i : i+prefixLen]
		next := words[i+prefixLen]
		logrus.Debug(strings.Join(prefix, "|"), " => ", next)
		insertEntry(strings.Join(prefix, " "), next)
	}
	if len(words) > prefixLen {
		insertStart(strings.Join(words[0:prefixLen], " "))
	}
}

func processInput(path string) (int, error) {
	fd, err := os.Open(path)
	if err != nil {
		return 0, err
	}
	defer fd.Close()

	reader := bufio.NewReader(fd)

	var line string
	var lines int

	for line, err = reader.ReadString('\n'); err == nil; line, err = reader.ReadString('\n') {
		lines++
		var str string
		if err := json.Unmarshal([]byte(line), &str); err != nil {
			return 0, err
		}
		processText(str)
	}
	if err != io.EOF {
		return 0, err
	}
	return lines, nil
}

var MarkovImportCmd = &cobra.Command{
	Use: "import file1 [file2 [file3]...]",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 1 {
			fmt.Println("Need at least one source")
			cmd.Usage()
			return
		}

		count := 0
		for _, path := range args {
			entries, err := processInput(path)
			if err != nil {
				logrus.Error(err)
				continue
			}
			count += entries
		}
		logrus.Info("Imported ", count, " texts")
	},
}

func getSeed() string {
	maxRowId, err := db.Query("SELECT MAX(rowid) FROM markov_start")
	defer maxRowId.Close()
	if err != nil {
		logrus.Panic(err)
	}
	var maxid int64
	maxRowId.Next()
	maxRowId.Scan(&maxid)

	rowid := rand.Int63n(maxid)
	seedRows, err := db.Query("SELECT prefix FROM markov_start WHERE rowid = ?", rowid)
	defer seedRows.Close()
	seedRows.Next()
	var seed string
	seedRows.Scan(&seed)
	return seed
}

func getNextToken(prefix []string) string {
	prefixStr := strings.Join(prefix, " ")
	rowIds, err := db.Query("SELECT rowid from markov_prefix WHERE prefix = ?", prefixStr)
	if err != nil {
		logrus.Panic(err)
	}
	defer rowIds.Close()

	if !rowIds.Next() {
		return ""
	}
	var rowid int64
	rowIds.Scan(&rowid)

	tokenRows, err := db.Query("SELECT token FROM markov_next WHERE prefix_rowid = ?", rowid)
	if err != nil {
		logrus.Panic(err)
	}
	defer tokenRows.Close()

	possibilities := []string{}
	for tokenRows.Next() {
		var token string
		tokenRows.Scan(&token)
		possibilities = append(possibilities, token)
	}
	logrus.Debug(len(possibilities), " possibilities: ", possibilities)
	i := rand.Int31n(int32(len(possibilities)))
	return possibilities[i]
}

func generate(seed string) string {
	buffer := bytes.NewBufferString(seed)
	count := 0
	window := strings.Split(seed, " ")

	for {
		logrus.Debug(window)
		next := getNextToken(window)
		if next == "" {
			logrus.Debug("End of chain")
			break
		}

		buffer.WriteString(" ")
		buffer.WriteString(next)

		window = append(window[1:prefixLen], next)
		count++
		if count > messageLen {
			break
		}
	}

	return buffer.String()
}

var MarkovGenerateCmd = &cobra.Command{
	Use: "generate",
	Run: func(cmd *cobra.Command, args []string) {
		rand.Seed(time.Now().Unix())

		seed := getSeed()
		logrus.Info("Using as seed ", seed)

		text := generate(seed)
		fmt.Println(text)
	},
}

func init() {
	MarkovCmd.PersistentFlags().StringVar(&dbPath, "db", "/tmp/baitbot.db", "Chain database")
	MarkovCmd.PersistentFlags().IntVar(&prefixLen, "prefix-len", 3, "Prefix length")
	MarkovGenerateCmd.PersistentFlags().IntVar(&messageLen, "len", 200, "Message length")

	MarkovCmd.AddCommand(MarkovClearCmd)
	MarkovCmd.AddCommand(MarkovImportCmd)
	MarkovCmd.AddCommand(MarkovGenerateCmd)
	RootCmd.AddCommand(MarkovCmd)
}
