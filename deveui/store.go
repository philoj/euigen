package deveui

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"strconv"
)

type IdStore struct {
	usedOrRejected map[uint64]struct{}
	shortLookup    map[uint64]struct{}
}

func (s *IdStore) generate() uint64 {
	max := new(big.Int).SetUint64(uint64(1<<64 - 1)) // 16 digit hex: [0,2^64)
	for {
		randId, err := rand.Int(rand.Reader, max)
		if err != nil {
			panic(err)
		}
		id := randId.Uint64()
		if s.isValid(id) {
			return id
		}
		fmt.Println("invalid", id, s.usedOrRejected, s.shortLookup)
	}
}

func (s *IdStore) updateStore(r result) {
	if r.success {
		s.shortLookup[r.id] = struct{}{}
	}
	s.usedOrRejected[r.id] = struct{}{}
}
func (s *IdStore) resetStore(ids []string) {
	s.shortLookup = make(map[uint64]struct{}, len(ids))
	s.usedOrRejected = make(map[uint64]struct{}, len(ids))
	for _, sid := range ids {
		id, err := strconv.ParseUint(sid, 16, 64)
		if err != nil {
			panic(err)
		}
		s.shortLookup[last5HexDigits(id)] = struct{}{}
		s.usedOrRejected[id] = struct{}{}
	}
}

func (s *IdStore) isValid(id uint64) bool {
	_, lookupUsed := s.shortLookup[last5HexDigits(id)]
	if lookupUsed {
		return false
	}
	_, trash := s.usedOrRejected[id]
	if trash {
		return false
	}
	return true
}

func last5HexDigits(id uint64) uint64 {
	return id & uint64(1<<20-1) // last five hex digits, 2^20
}

func NewIdStore() *IdStore {
	return &IdStore{
		usedOrRejected: make(map[uint64]struct{}),
		shortLookup:    make(map[uint64]struct{}),
	}
}
