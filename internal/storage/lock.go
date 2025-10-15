package storage

import (
	"context"
	"fmt"

	"cirello.io/dynamolock/v2"
	"go.uber.org/zap"
)

func (dbs DynamoDBStorage) prefixLock(key string) string {
	return fmt.Sprintf("LOCK-%s", key)
}

func (dbs DynamoDBStorage) GetLock(key string) (*dynamolock.Lock, bool) {
	dbs.mutex.RLock()
	defer dbs.mutex.RUnlock()

	if lock, exists := dbs.locks[key]; exists {
		return lock, true
	}
	return nil, false
}

// Lock acquires a distributed lock for the given key or blocks until it gets one.
func (dbs DynamoDBStorage) Lock(ctx context.Context, key string) error {
	dbs.logger.Debug("acquiring lock", zap.String("key", key))

	if _, isLocked := dbs.GetLock(key); isLocked {
		return nil
	}

	lock, err := dbs.Locker.AcquireLockWithContext(ctx, dbs.prefixLock(key))
	if err != nil {
		return fmt.Errorf("unable to acquire lock: %w", err)
	}

	dbs.mutex.Lock()
	dbs.locks[key] = lock
	dbs.mutex.Unlock()

	return nil
}

// Unlock releases a specific lock.
func (dbs DynamoDBStorage) Unlock(ctx context.Context, name string) error {
	// check if we own it and unlock
	lock, exists := dbs.GetLock(name)
	if !exists {
		return fmt.Errorf("lock %s not found", name)
	}

	_, err := dbs.Locker.ReleaseLockWithContext(ctx, lock)
	if err != nil {
		return fmt.Errorf("unable to unlock %s: %w", name, err)
	}

	dbs.mutex.RLock()
	delete(dbs.locks, name)
	dbs.mutex.RUnlock()

	return nil
}
