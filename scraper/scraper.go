package scraper

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/disksing/twiyou/store"
	"github.com/disksing/twiyou/twitter"
)

const (
	LErr  = "error"
	LInfo = "info"
)

const (
	InitialState        = "initial"
	FetchFollowersState = "fetch_followers"
	FetchFollowingState = "fetch_following"
	PullUsersState      = "pull_users"
	SumEventsState      = "sum_events"
	StashUsersState     = "stash_users"
	CleanUpState        = "cleanup"
	CompleteState       = "complete"
)

type Scraper struct {
	db   *store.DB
	self *twitter.User
}

func NewScraper() (*Scraper, error) {
	db, err := store.NewDB()
	if err != nil {
		return nil, err
	}
	self, err := twitter.LoadSelf()
	if err != nil {
		return nil, err
	}
	return &Scraper{
		db:   db,
		self: self,
	}, nil
}

func (s *Scraper) Close() {
	s.db.Close()
}

func (s *Scraper) Run() error {
	err := s.saveStats()
	if err != nil {
		return err
	}

	it, err := s.db.LoadLastIteration()
	if err != nil {
		s.db.Log(LErr, fmt.Sprintf("failed to load last iteration: %v", err))
		return err
	}

	for {
		switch it.State {
		case InitialState:
			it.State = FetchFollowersState

		case FetchFollowersState:
			if it.NextToken == "EOF" {
				it.State, it.NextToken = FetchFollowingState, ""
				it.CompleteFetchFollowersAt = sql.NullTime{Time: time.Now(), Valid: true}
				break
			}

			users, token, err := twitter.ListFriends(s.self.ID, "followers", it.NextToken)
			if err != nil {
				s.db.Log(LErr, fmt.Sprintf("failed to fetch followers: %v", err))
				return err
			}
			err = s.db.UpdateUserInfo("is_follower", users)
			if err != nil {
				s.db.Log(LErr, fmt.Sprintf("failed to update followers: %v", err))
				return err
			}
			if token == "" {
				token = "EOF"
			}
			it.NextToken = token

		case FetchFollowingState:
			if it.NextToken == "EOF" {
				it.State, it.NextToken = PullUsersState, ""
				it.CompleteFetchFollowingAt = sql.NullTime{Time: time.Now(), Valid: true}
				if err = s.db.SaveIteration(it); err != nil {
					return err
				}
				break
			}
			users, token, err := twitter.ListFriends(s.self.ID, "following", it.NextToken)
			if err != nil {
				s.db.Log(LErr, fmt.Sprintf("failed to fetch following: %v", err))
				return err
			}
			err = s.db.UpdateUserInfo("is_following", users)
			if err != nil {
				s.db.Log(LErr, fmt.Sprintf("failed to update following: %v", err))
				return err
			}
			if token == "" {
				token = "EOF"
			}
			it.NextToken = token

		case PullUsersState:
			userIDs, err := s.db.SelectUsersForPull(it.StartedAt, 100)
			if err != nil {
				s.db.Log(LErr, fmt.Sprintf("failed to pull users: %v", err))
				return err
			}
			if len(userIDs) > 0 {
				users, err := twitter.ListUsers(userIDs)
				if err != nil {
					s.db.Log(LErr, fmt.Sprintf("failed to list users: %v", err))
					return err
				}
				err = s.db.UpdateUserInfo("", users)
				if err != nil {
					s.db.Log(LErr, fmt.Sprintf("failed to update users: %v", err))
					return err
				}
				err = s.db.UpdateUserUpdateTime(userIDs)
				if err != nil {
					s.db.Log(LErr, fmt.Sprintf("failed to update user update time: %v", err))
					return err
				}
			}
			it.State = SumEventsState
			it.CompletePullUsersAt = sql.NullTime{Time: time.Now(), Valid: true}

		case SumEventsState:
			err = s.db.SumUserEvents(it.StartedAt)
			if err != nil {
				s.db.Log(LErr, fmt.Sprintf("failed to sum events: %v", err))
				return err
			}
			it.State = StashUsersState
			it.CompleteSumEventsAt = sql.NullTime{Time: time.Now(), Valid: true}

		case StashUsersState:
			err = s.db.StashUsers()
			if err != nil {
				s.db.Log(LErr, fmt.Sprintf("failed to stash users: %v", err))
				return err
			}
			it.State = CleanUpState
			it.CompleteStashUsersAt = sql.NullTime{Time: time.Now(), Valid: true}

		case CleanUpState:
			err = s.db.CleanUpCache()
			if err != nil {
				s.db.Log(LErr, fmt.Sprintf("failed to clean up: %v", err))
				return err
			}
			it.State = CompleteState
			it.CompletedAt = sql.NullTime{Time: time.Now(), Valid: true}
			if err = s.db.SaveIteration(it); err != nil {
				return err
			}
			s.db.Log(LInfo, fmt.Sprintf("iteration completed: %v", it.ID))
			return nil

		case CompleteState:
			it, err = s.db.CreateIteration()
			if err != nil {
				s.db.Log(LErr, fmt.Sprintf("failed to create iteration: %v", err))
				return err
			}
			s.db.Log(LInfo, fmt.Sprintf("new iteration: %v", it.ID))
			continue
		}

		if err = s.db.SaveIteration(it); err != nil {
			return err
		}
	}
}

func (s *Scraper) saveIteration(it *store.Iteration) error {
	err := s.db.SaveIteration(it)
	if err != nil {
		s.db.Log(LErr, fmt.Sprintf("failed to save iteration: %v", err))
		return err
	}
	return nil
}

func (s *Scraper) saveStats() error {
	err := s.db.InsertStats(s.self)
	if err != nil {
		s.db.Log(LErr, fmt.Sprintf("failed to save stats: %v", err))
		return err
	}
	s.db.Log(LInfo, fmt.Sprintf("saved stats %v", s.self.Metrics()))
	return nil
}
