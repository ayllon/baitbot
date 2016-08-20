package markov

// getPrefixId returns the prefix id. If it does not exist, and create is true, it will be
// inserted.
// Returns -1 if it does not exist and create is false
func (m *Markov) getPrefixId(prefix string, create bool) (int64, error) {
	rows, err := m.db.Query("SELECT rowid FROM markov_prefix WHERE prefix = ?", prefix)
	if err != nil {
		return -1, err
	}
	defer rows.Close()

	if rows.Next() {
		var rowid int64
		return rowid, rows.Scan(&rowid)
	} else if create {
		result, err := m.db.Exec("INSERT INTO markov_prefix (prefix) VALUES (?)", prefix)
		if err != nil {
			return -1, err
		}
		return result.LastInsertId()
	}
	return -1, nil
}

// insertEntry inserts into the DB a new link between a prefix and a token
func (m *Markov) insertEntry(prefix string, next string) error {
	prefixId, err := m.getPrefixId(prefix, true)
	if err != nil {
		return err
	}
	_, err = m.db.Exec("INSERT INTO markov_next (prefix_rowid, token) VALUES (?, ?)", prefixId, next)
	return err
}

// insertStart inserts a new prefix that can be considered the start of a sentence
func (m *Markov) insertStart(prefix string) error {
	_, err := m.db.Exec("INSERT INTO markov_start (prefix) VALUES (?)", prefix)
	return err
}
