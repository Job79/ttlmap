package ttlmap

import (
	"sync"
	"time"
)

// TTLMap is an efficient concurrent map with TTL support.
//
// It uses a sync.Map internally as storage, and keeps track
// of expiration times using a [][]MapKey slice. The outer
// slice represents a generation, while the inner slice
// contains bucket keys. All keys in a generation expire at
// the same time.
//
// There are (ttl / interval) generations, every time the internal timer
// ticks the map advances a generation. The TTLMap does the following when advancing:
//   - Remove all items that are in the next generation.
//   - Increase the currentGeneration by 1. By doing this
//     new items will be added to the next generation.
//
// Because of this, the map is never scanned and cleaning up
// expired items is very cheap. The downside of this approach
// is that it uses a little more memory, and is not perfectly accurate.
type TTLMap[K comparable, V any] struct {
	items sync.Map

	generations [][]K
	ticker      *time.Ticker
	currentGen  int
}

// New creates a new TTLMap.
//
// The ttl is the time-to-live for each item in the map. The
// interval determines how often the TTLMap checks for
// expired items. A small interval value uses a tiny bit
// more memory and CPU, but is more accurate.
func New[K comparable, V any](ttl, interval time.Duration) *TTLMap[K, V] {
	ttlMap := &TTLMap[K, V]{
		ticker:      time.NewTicker(interval),
		generations: make([][]K, ttl/interval),
	}

	go func() {
		for range ttlMap.ticker.C {
			ttlMap.nextGeneration()
		}
	}()

	return ttlMap
}

// Load returns the value stored in the map for a key, or nil
// if no value is present. The ok result indicates whether
// value was found in the map.
func (m *TTLMap[K, V]) Load(key any) (V, bool) {
	val, ok := m.items.Load(key)
	if !ok {
		return *new(V), false
	}
	return val.(V), ok
}

// Store sets the value for a key.
func (m *TTLMap[K, V]) Store(key K, value V) {
	m.items.Store(key, value)
	m.addToGeneration(key)
}

// Delete deletes the value for a key.
//
// It does not reset the TTL value for the key, because this
// would be a very expensive operation. The ticker will
// automatically remove the TTL value. Do not reuses keys
// after deleting them, else they might expire faster then
// expected.
func (m *TTLMap[K, V]) Delete(key K) {
	m.items.Delete(key)
}

// LoadOrStore returns the existing value for the key if
// present. Otherwise, it stores and returns the given
// value. The loaded result is true if the value was loaded,
// false if stored.
func (m *TTLMap[K, V]) LoadOrStore(key K, value V) (actual V, loaded bool) {
	if actual, loaded := m.items.LoadOrStore(key, value); loaded {
		return actual.(V), true
	} else {
		m.addToGeneration(key)
		return value, false
	}
}

// LoadAndDelete deletes the value for a key, returning the
// previous value if any. The loaded result reports whether
// the key was present.
//
// It does not reset the TTL value for the key, because this
// would be a very expensive operation. The ticker will
// automatically remove the TTL value. Do not reuses keys
// after deleting them, else they might expire faster then
// expected.
func (m *TTLMap[K, V]) LoadAndDelete(key K) (V, bool) {
	if val, ok := m.items.LoadAndDelete(key); ok {
		return val.(V), ok
	}
	return *new(V), false
}

// Range calls f sequentially for each key and value present
// in the map. If f returns false, range stops the
// iteration.
func (m *TTLMap[K, V]) Range(f func(key K, value V) bool) {
	m.items.Range(func(key any, value any) bool {
		return f(key.(K), value.(V))
	})
}

// Close stops the ticker.
func (m *TTLMap[K, V]) Close() {
	m.ticker.Stop()
}

// nextGeneration advances the TTLMap to the next generation.
func (m *TTLMap[K, V]) nextGeneration() {
	nextGen := (m.currentGen + 1) % len(m.generations)

	// Remove all items that are stored in the next
	// generation. These are expired.
	for _, key := range m.generations[nextGen] {
		m.items.Delete(key)
	}

	// addToGeneration grows the backing array of the inner
	// slice when many items are added to a single
	// generation. When the capacity isn't used in the next
	// generation, shrink the slice.
	if len(m.generations) < cap(m.generations[nextGen])/8 {
		m.generations[nextGen] = make([]K, cap(m.generations[nextGen])/8)
	}

	// Reset the next generation.
	m.generations[nextGen] = m.generations[nextGen][:0]
	m.currentGen = nextGen
}

// addToGeneration adds a key to the current generation.
func (m *TTLMap[K, V]) addToGeneration(key K) {
	m.generations[m.currentGen] = append(m.generations[m.currentGen], key)
}
