package store

import (
	"fmt"
	"time"

	"github.com/disksing/twiyou/twitter"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
)

func (db *DB) InsertStats(self *twitter.User) error {
	_, err := db.db.Exec(`
	INSERT INTO stats (
		id, name, user_name, profile_image,
		follower_count, following_count, tweet_count, listed_count)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`, self.ID, self.Name, self.UserName, self.ProfileImageURL,
		self.PublicMetrics.FollowersCount, self.PublicMetrics.FollowingCount, self.PublicMetrics.TweetCount, self.PublicMetrics.ListedCount)
	if err != nil {
		return errors.New("update self: " + err.Error())
	}
	return nil
}

func (db *DB) UpdateUserInfo(relation string, users []twitter.User) error {
	if len(users) == 0 {
		return nil
	}

	sql := `
	INSERT INTO users (
		id, name, user_name, profile_image,
		follower_count, following_count, tweet_count1, listed_count,
		updated_at %s)
	VALUES %%s
	ON DUPLICATE KEY UPDATE
		name = VALUES(name), user_name = VALUES(user_name), profile_image = VALUES(profile_image),
		follower_count = VALUES(follower_count), following_count = VALUES(following_count), tweet_count1 = VALUES(tweet_count1), listed_count = VALUES(listed_count),
		updated_at = VALUES(updated_at) %s
	`

	var param1, param2 string
	if relation != "" {
		param1 = fmt.Sprintf(",%s1", relation)                   // `,is_following1` / `,is_follower1`
		param2 = fmt.Sprintf(",%[1]s1=VALUES(%[1]s1)", relation) // `,is_following1=VALUES(is_following1)` / `,is_follower1=VALUES(is_follower1)`
	}
	sql = fmt.Sprintf(sql, param1, param2)

	var args []any
	for _, u := range users {
		args = append(args,
			u.ID, u.Name, u.UserName, u.ProfileImageURL,
			u.PublicMetrics.FollowersCount, u.PublicMetrics.FollowingCount, u.PublicMetrics.TweetCount, u.PublicMetrics.ListedCount,
			time.Now())
		if relation != "" {
			args = append(args, true)
		}
	}
	err := db.batchInsert(sql, 200, len(args)/len(users), args)
	if err != nil {
		return errors.New("update users: " + err.Error())
	}
	return nil
}

func (db *DB) SelectUsersForPull(t time.Time, limit int) ([]string, error) {
	var ids []string
	err := db.db.Select(&ids, `SELECT id FROM users WHERE updated_at < ? ORDER BY updated_at LIMIT ?`, t, limit)
	return ids, err
}

func (db *DB) UpdateUserUpdateTime(users []string) error {
	if len(users) == 0 {
		return nil
	}
	query, args, err := sqlx.In(`UPDATE users SET updated_at = ? WHERE id IN (?)`, time.Now(), users)
	if err != nil {
		return err
	}
	_, err = db.db.Exec(query, args...)
	return err
}

func (db *DB) SumUserEvents(t time.Time) error {
	sqls := []string{
		"DELETE FROM events WHERE created_at = ?",
		"UPDATE users SET last_active = ? WHERE tweet_count != tweet_count1",
		"INSERT INTO events SELECT id, ?, 'new_follower' FROM users WHERE is_follower = FALSE and is_follower1 = TRUE",
		"INSERT INTO events SELECT id, ?, 'lost_follower' FROM users WHERE is_follower = TRUE and is_follower1 = FALSE",
		"INSERT INTO events SELECT id, ?, 'new_following' FROM users WHERE is_following = FALSE and is_following1 = TRUE",
		"INSERT INTO events SELECT id, ?, 'cancel_following' FROM users WHERE is_following = TRUE and is_following1 = FALSE",
	}
	for _, sql := range sqls {
		_, err := db.db.Exec(sql, t)
		if err != nil {
			return errors.Wrap(err, "sum events failed")
		}
	}
	return nil
}

func (db *DB) StashUsers() error {
	_, err := db.db.Exec(`
	UPDATE users SET
		tweet_count = tweet_count1,
		is_following = is_following1,
		is_follower = is_follower1
			`)
	if err != nil {
		return errors.New("failed to stash user relationships:" + err.Error())
	}
	return nil
}

func (db *DB) CleanUpCache() error {
	_, err := db.db.Exec(`
	UPDATE users SET
		is_following1 = FALSE,
		is_follower1 = FALSE
			`)
	if err != nil {
		return errors.New("failed to clean up cache:" + err.Error())
	}
	return nil
}
