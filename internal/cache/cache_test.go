package cache

import (
	"testing"
	"time"
)

func TestMap_SetAndGet(t *testing.T) {
	c := NewMap[string, int](time.Minute, 10)

	c.Set("a", 1)
	c.Set("b", 2)

	val, ok := c.Get("a")
	if !ok || val != 1 {
		t.Errorf("Get(a) = %d, %v; want 1, true", val, ok)
	}

	val, ok = c.Get("b")
	if !ok || val != 2 {
		t.Errorf("Get(b) = %d, %v; want 2, true", val, ok)
	}
}

func TestMap_GetMissing(t *testing.T) {
	c := NewMap[string, int](time.Minute, 10)

	val, ok := c.Get("missing")
	if ok {
		t.Errorf("Get(missing) = %d, %v; want 0, false", val, ok)
	}
}

func TestMap_Expiration(t *testing.T) {
	c := NewMap[string, int](10*time.Millisecond, 10)

	c.Set("key", 42)

	val, ok := c.Get("key")
	if !ok || val != 42 {
		t.Errorf("Get(key) before expiry = %d, %v; want 42, true", val, ok)
	}

	time.Sleep(20 * time.Millisecond)

	_, ok = c.Get("key")
	if ok {
		t.Error("Get(key) after expiry should return false")
	}
}

func TestMap_SetWithTTL(t *testing.T) {
	c := NewMap[string, int](time.Hour, 10) // default TTL is long

	c.SetWithTTL("short", 1, 10*time.Millisecond)
	c.Set("long", 2)

	time.Sleep(20 * time.Millisecond)

	_, ok := c.Get("short")
	if ok {
		t.Error("short-TTL entry should have expired")
	}

	val, ok := c.Get("long")
	if !ok || val != 2 {
		t.Errorf("long-TTL entry should still exist, got %d, %v", val, ok)
	}
}

func TestMap_Eviction(t *testing.T) {
	c := NewMap[string, int](time.Minute, 3)

	c.Set("a", 1)
	c.Set("b", 2)
	c.Set("c", 3)
	// Cache is full (3/3), next set triggers eviction
	c.Set("d", 4)

	// "a" should have been evicted (oldest)
	_, ok := c.Get("a")
	if ok {
		t.Error("entry 'a' should have been evicted")
	}

	// "d" should exist
	val, ok := c.Get("d")
	if !ok || val != 4 {
		t.Errorf("Get(d) = %d, %v; want 4, true", val, ok)
	}
}

func TestMap_Delete(t *testing.T) {
	c := NewMap[string, int](time.Minute, 10)

	c.Set("a", 1)
	c.Delete("a")

	_, ok := c.Get("a")
	if ok {
		t.Error("deleted entry should not be found")
	}
}

func TestMap_Clear(t *testing.T) {
	c := NewMap[string, int](time.Minute, 10)

	c.Set("a", 1)
	c.Set("b", 2)
	c.Clear()

	_, ok := c.Get("a")
	if ok {
		t.Error("cleared cache should have no entries")
	}
}

func TestMap_Keys(t *testing.T) {
	c := NewMap[string, int](time.Minute, 10)

	c.Set("a", 1)
	c.Set("b", 2)
	c.Set("c", 3)

	keys := c.Keys()
	if len(keys) != 3 {
		t.Errorf("Keys() returned %d keys, want 3", len(keys))
	}

	keySet := make(map[string]bool)
	for _, k := range keys {
		keySet[k] = true
	}
	for _, expected := range []string{"a", "b", "c"} {
		if !keySet[expected] {
			t.Errorf("Keys() missing key %q", expected)
		}
	}
}

func TestMap_Keys_excludesExpired(t *testing.T) {
	c := NewMap[string, int](10*time.Millisecond, 10)

	c.Set("expired", 1)
	c.SetWithTTL("fresh", 2, time.Hour)

	time.Sleep(20 * time.Millisecond)

	keys := c.Keys()
	if len(keys) != 1 || keys[0] != "fresh" {
		t.Errorf("Keys() = %v; want [fresh]", keys)
	}
}

func TestMap_IntKeys(t *testing.T) {
	c := NewMap[int, string](time.Minute, 10)

	c.Set(1, "one")
	c.Set(2, "two")

	val, ok := c.Get(1)
	if !ok || val != "one" {
		t.Errorf("Get(1) = %q, %v; want \"one\", true", val, ok)
	}
}
