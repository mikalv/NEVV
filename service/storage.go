package service

import (
	"errors"
	"sync"

	"github.com/dedis/cothority/skipchain"
	"github.com/qantik/nevv/protocol"
)

// Storage offers the possibilty to store elections permanently on
// the disk. This is especially useful when multiple elections have to
// kept alive after potential shutdowns of the conode.
type Storage struct {
	sync.Mutex

	Elections map[string]*Election
}

// Get retrieves an election for a given name.
func (storage *Storage) get(name string) (*Election, error) {

	storage.Lock()
	defer storage.Unlock()

	election, found := storage.Elections[name]
	if !found {
		return nil, errors.New("Election " + name + " not found")
	}

	return election, nil
}

// CreateElection adds a new election structure to the storage map.
func (storage *Storage) createElection(name string, genesis, latest *skipchain.SkipBlock,
	shared *protocol.SharedSecret) {

	storage.Lock()
	defer storage.Unlock()

	if latest == nil {
		storage.Elections[name] = &Election{genesis, genesis, shared}
	} else {
		storage.Elections[name] = &Election{genesis, latest, shared}
	}
}

// UpdateLatest replaces the latest SkipBlock of an election by a given SkipBlock.
func (storage *Storage) updateLatest(name string, latest *skipchain.SkipBlock) {
	storage.Lock()
	defer storage.Unlock()

	storage.Elections[name].Latest = latest
}