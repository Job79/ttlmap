package ttlmap

import (
	"strconv"
	"testing"
	"time"
)

func TestLoad(t *testing.T) {
	ttlmap := New[string, string](time.Hour, time.Minute)
	ttlmap.items.Store("key", "value")

	if value, ok := ttlmap.Load("key"); !ok {
		t.Errorf("Expected to find key, but did not")
	} else if value != "value" {
		t.Errorf("Expected value to be 'value', but was '%s'", value)
	}
}

func TestStore(t *testing.T) {
	ttlmap := New[string, string](time.Hour, time.Minute)
	ttlmap.Store("key", "value")

	if value, ok := ttlmap.items.Load("key"); !ok {
		t.Errorf("Expected to find key, but did not")
	} else if value != "value" {
		t.Errorf("Expected value to be 'value', but was '%s'", value)
	} else if ttlmap.generations[ttlmap.currentGen][0] != "key" {
		t.Errorf("Expected key to be in generation 0, but was not")
	}
}

func TestDelete(t *testing.T) {
	ttlmap := New[string, string](time.Hour, time.Minute)
	ttlmap.items.Store("key", "value")
	ttlmap.Delete("key")

	if _, ok := ttlmap.items.Load("key"); ok {
		t.Errorf("Expected to not find key, but did")
	}
}

func TestLoadOrStore(t *testing.T) {
	ttlmap := New[string, string](time.Hour, time.Minute)

	if value, loaded := ttlmap.LoadOrStore("key", "value"); loaded {
		t.Errorf("Expected to not find key, but did")
	} else if value != "value" {
		t.Errorf("Expected value to be 'value', but was '%s'", value)
	} else if ttlmap.generations[ttlmap.currentGen][0] != "key" {
		t.Errorf("Expected key to be in generation 0, but was not")
	} else if value, loaded = ttlmap.LoadOrStore("key", "value2"); !loaded {
		t.Errorf("Expected to find key, but did not")
	} else if value != "value" {
		t.Errorf("Expected value to be 'value', but was '%s'", value)
	}
}

func TestLoadAndDelete(t *testing.T) {
	ttlmap := New[string, string](time.Hour, time.Minute)
	ttlmap.items.Store("key", "value")

	if value, loaded := ttlmap.LoadAndDelete("key"); !loaded {
		t.Errorf("Expected to find key, but did not")
	} else if value != "value" {
		t.Errorf("Expected value to be 'value', but was '%s'", value)
	} else if _, loaded = ttlmap.LoadAndDelete("key"); loaded {
		t.Errorf("Expected to not find key, but did")
	}
}

func TestRange(t *testing.T) {
	ttlmap := New[string, string](time.Hour, time.Minute)
	ttlmap.items.Store("key1", "value1")
	ttlmap.items.Store("key2", "value2")

	var (
		keys   []string
		values []string
	)
	ttlmap.Range(func(key string, value string) bool {
		keys = append(keys, key)
		values = append(values, value)
		return true
	})

	if len(keys) != 2 {
		t.Errorf("Expected 2 keys, but got %d", len(keys))
	} else if len(values) != 2 {
		t.Errorf("Expected 2 values, but got %d", len(values))
	} else if keys[0] != "key1" || keys[1] != "key2" {
		t.Errorf("Expected keys to be 'key1' and 'key2', but were '%s' and '%s'", keys[0], keys[1])
	} else if values[0] != "value1" || values[1] != "value2" {
		t.Errorf("Expected values to be 'value1' and 'value2', but were '%s' and '%s'", values[0], values[1])
	}
}

func TestNextGeneration(t *testing.T) {
	ttlmap := New[string, string](2, 1)
	ttlmap.Store("key1", "value1")

	ttlmap.nextGeneration()
	if _, ok := ttlmap.Load("key1"); !ok {
		t.Errorf("Expected to find key1, but did not")
	}

	ttlmap.nextGeneration()
	if _, ok := ttlmap.Load("key1"); ok {
		t.Errorf("Expected to not find key1, but did")
	}
}

func BenchmarkStore(b *testing.B) {
	ttlmap := New[string, string](time.Duration(b.N+1)*time.Minute, time.Minute)
	for i := 0; i < b.N; i++ {
		ttlmap.Store(strconv.Itoa(i), "value")
		ttlmap.nextGeneration()
	}
}

func BenchmarkLoad(b *testing.B) {
	b.StopTimer()
	ttlmap := New[string, string](time.Duration(b.N+1)*time.Minute, time.Minute)
	for i := 0; i < b.N; i++ {
		ttlmap.Store(strconv.Itoa(i), "value")
		ttlmap.nextGeneration()
	}

	b.StartTimer()
	for i := 0; i < b.N; i++ {
		ttlmap.Load(strconv.Itoa(i))
	}
}

func BenchmarkDelete(b *testing.B) {
	b.StopTimer()
	ttlmap := New[string, string](time.Duration(b.N+1)*time.Minute, time.Minute)
	for i := 0; i < b.N; i++ {
		ttlmap.Store(strconv.Itoa(i), "value")
		ttlmap.nextGeneration()
	}

	b.StartTimer()
	for i := 0; i < b.N; i++ {
		ttlmap.Delete(strconv.Itoa(i))
	}
}

func BenchmarkNextGeneration(b *testing.B) {
	b.StopTimer()
	ttlmap := New[string, string](time.Duration(b.N+1)*time.Minute, time.Minute)
	for i := 0; i < b.N; i++ {
		ttlmap.Store(strconv.Itoa(i), "value")
		ttlmap.nextGeneration()
	}

	b.StartTimer()
	for i := 0; i < b.N; i++ {
		ttlmap.nextGeneration()
	}
}
