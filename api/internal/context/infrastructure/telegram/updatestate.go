package telegram

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/gotd/td/telegram/updates"
)

type updateState struct {
	path string

	mu    sync.Mutex
	users map[int64]*userState
}

type userState struct {
	State    *updates.State  `json:"state,omitempty"`
	Channels map[int64]int   `json:"channels,omitempty"`
	Hashes   map[int64]int64 `json:"hashes,omitempty"`
}

func newUpdateState(path string) (*updateState, error) {
	s := &updateState{path: path, users: make(map[int64]*userState)}

	data, err := os.ReadFile(filepath.Clean(path))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return s, nil
		}

		return nil, fmt.Errorf("read update state: %w", err)
	}

	if err := json.Unmarshal(data, &s.users); err != nil {
		return nil, fmt.Errorf("parse update state: %w", err)
	}

	return s, nil
}

func (s *updateState) user(id int64) *userState {
	u, ok := s.users[id]
	if !ok {
		u = &userState{Channels: make(map[int64]int), Hashes: make(map[int64]int64)}
		s.users[id] = u
	}

	if u.Channels == nil {
		u.Channels = make(map[int64]int)
	}

	if u.Hashes == nil {
		u.Hashes = make(map[int64]int64)
	}

	return u
}

func (s *updateState) save() error {
	data, err := json.Marshal(s.users)
	if err != nil {
		return fmt.Errorf("encode update state: %w", err)
	}

	if err := os.MkdirAll(filepath.Dir(s.path), 0o700); err != nil {
		return fmt.Errorf("create update state dir: %w", err)
	}

	if err := os.WriteFile(s.path, data, 0o600); err != nil {
		return fmt.Errorf("write update state: %w", err)
	}

	return nil
}

func (s *updateState) GetState(_ context.Context, userID int64) (updates.State, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	u, ok := s.users[userID]
	if !ok || u.State == nil {
		return updates.State{}, false, nil
	}

	return *u.State, true, nil
}

func (s *updateState) SetState(_ context.Context, userID int64, state updates.State) error {
	return s.mutate(userID, func(u *userState) { u.State = &state })
}

func (s *updateState) SetPts(_ context.Context, userID int64, pts int) error {
	return s.mutateExisting(userID, func(u *userState) { u.State.Pts = pts })
}

func (s *updateState) SetQts(_ context.Context, userID int64, qts int) error {
	return s.mutateExisting(userID, func(u *userState) { u.State.Qts = qts })
}

func (s *updateState) SetDate(_ context.Context, userID int64, date int) error {
	return s.mutateExisting(userID, func(u *userState) { u.State.Date = date })
}

func (s *updateState) SetSeq(_ context.Context, userID int64, seq int) error {
	return s.mutateExisting(userID, func(u *userState) { u.State.Seq = seq })
}

func (s *updateState) SetDateSeq(_ context.Context, userID int64, date, seq int) error {
	return s.mutateExisting(userID, func(u *userState) { u.State.Date, u.State.Seq = date, seq })
}

func (s *updateState) GetChannelPts(_ context.Context, userID, channelID int64) (int, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	u, ok := s.users[userID]
	if !ok {
		return 0, false, nil
	}

	pts, ok := u.Channels[channelID]

	return pts, ok, nil
}

func (s *updateState) SetChannelPts(_ context.Context, userID, channelID int64, pts int) error {
	return s.mutate(userID, func(u *userState) { u.Channels[channelID] = pts })
}

func (s *updateState) ForEachChannels(ctx context.Context, userID int64, fn func(ctx context.Context, channelID int64, pts int) error) error {
	s.mu.Lock()

	channels := make(map[int64]int)

	if u, ok := s.users[userID]; ok {
		for id, pts := range u.Channels {
			channels[id] = pts
		}
	}

	s.mu.Unlock()

	for id, pts := range channels {
		if err := fn(ctx, id, pts); err != nil {
			return err
		}
	}

	return nil
}

func (s *updateState) SetChannelAccessHash(_ context.Context, userID, channelID, accessHash int64) error {
	return s.mutate(userID, func(u *userState) { u.Hashes[channelID] = accessHash })
}

func (s *updateState) GetChannelAccessHash(_ context.Context, userID, channelID int64) (int64, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	u, ok := s.users[userID]
	if !ok {
		return 0, false, nil
	}

	hash, ok := u.Hashes[channelID]

	return hash, ok, nil
}

func (s *updateState) mutate(userID int64, fn func(u *userState)) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	fn(s.user(userID))

	return s.save()
}

func (s *updateState) mutateExisting(userID int64, fn func(u *userState)) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	u, ok := s.users[userID]
	if !ok || u.State == nil {
		return fmt.Errorf("update state for user %d does not exist", userID)
	}

	fn(u)

	return s.save()
}
