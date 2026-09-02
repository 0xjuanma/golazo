package reddit

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	"github.com/0xjuanma/golazo/internal/data"
)

const queueStateFileName = "reddit_queue_state.json"

// queueState is the on-disk representation of goalQueue's block state.
type queueState struct {
	CooldownUntil     time.Time `json:"cooldown_until"`
	ConsecutiveBlocks int       `json:"consecutive_blocks"`
}

// queueStateStore persists goalQueue's block state so a cooldown survives
// process restarts. Best-effort like GoalLinkCache: load/save errors are
// swallowed by callers, never block a fetch. Only accessed from the queue's
// single worker goroutine (save) and once at construction (load), so no
// internal locking is needed.
type queueStateStore struct {
	filePath string
}

// newQueueStateStore creates a store backed by the standard config dir.
func newQueueStateStore() (*queueStateStore, error) {
	dir, err := data.ConfigDir()
	if err != nil {
		return nil, err
	}
	return &queueStateStore{filePath: filepath.Join(dir, queueStateFileName)}, nil
}

// load reads the persisted state from disk. Returns a zero-value queueState
// if the file doesn't exist or can't be parsed — mirrors GoalLinkCache.load's
// "start empty" behavior.
func (s *queueStateStore) load() queueState {
	data, err := os.ReadFile(s.filePath)
	if err != nil {
		return queueState{}
	}

	var state queueState
	if err := json.Unmarshal(data, &state); err != nil {
		return queueState{}
	}
	return state
}

// save persists state to disk.
func (s *queueStateStore) save(state queueState) error {
	data, err := json.Marshal(state)
	if err != nil {
		return err
	}
	return os.WriteFile(s.filePath, data, 0644)
}
