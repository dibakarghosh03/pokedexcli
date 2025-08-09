package pokecache

import (
	"testing"
	"time"
)

func TestCacheAddAndGet(t *testing.T) {
	c := NewCache(50 * time.Millisecond)

	key := "test-key"
	val := []byte("hello")

	c.Add(key, val)

	data, ok := c.Get(key)
	if !ok {
		t.Fatalf("expected to find key %q", key)
	}

	if string(data) != "hello" {
		t.Fatalf("expected value %q, got %q", "hello", string(data))
	}
}

func TestCacheExpiration(t *testing.T) {
	c := NewCache(50 * time.Millisecond)

	key := "test-key"
	val := []byte("expire me")

	c.Add(key, val)

	time.Sleep(80 * time.Millisecond) // wait for expiry

	if _, ok := c.Get(key); ok {
		t.Errorf("expected key %q to be expired", key)
	}
}