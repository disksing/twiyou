package store

import (
	"database/sql"
	"time"
)

type Iteration struct {
	ID                       int64        `db:"id"`
	State                    string       `db:"state"`
	StartedAt                time.Time    `db:"started_at"`
	CompleteFetchFollowersAt sql.NullTime `db:"complete_fetch_followers_at"`
	CompleteFetchFollowingAt sql.NullTime `db:"complete_fetch_following_at"`
	CompletePullUsersAt      sql.NullTime `db:"complete_pull_users_at"`
	CompleteSumEventsAt      sql.NullTime `db:"complete_sum_events_at"`
	CompleteStashUsersAt     sql.NullTime `db:"complete_stash_users_at"`
	CompletedAt              sql.NullTime `db:"completed_at"`
	NextToken                string       `db:"next_token"`
}

func (db *DB) LoadLastIteration() (*Iteration, error) {
	var it Iteration
	err := db.db.Get(&it, "SELECT * FROM iterations ORDER BY id DESC LIMIT 1")
	if err != nil {
		return nil, err
	}
	return &it, nil
}

func (db *DB) SaveIteration(it *Iteration) error {
	_, err := db.db.NamedExec(`
		UPDATE iterations SET
			state = :state,
			started_at = :started_at,
			complete_fetch_followers_at = :complete_fetch_followers_at,
			complete_fetch_following_at = :complete_fetch_following_at,
			complete_pull_users_at = :complete_pull_users_at,
			complete_sum_events_at = :complete_sum_events_at,
			complete_stash_users_at = :complete_stash_users_at,
			completed_at = :completed_at
		WHERE id = :id
	`, it)
	return err
}

func (db *DB) CreateIteration() (*Iteration, error) {
	_, err := db.db.Exec(`INSERT INTO iterations (state) VALUES (?)`, "initial")
	if err != nil {
		return nil, err
	}
	return db.LoadLastIteration()
}
