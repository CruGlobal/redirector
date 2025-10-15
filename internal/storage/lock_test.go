package storage_test

import (
	"path"
	"sync"

	"github.com/CruGlobal/redirector/internal/storage"
	"github.com/CruGlobal/redirector/redirtest"
)

func (ts *StorageTestSuite) TestStorage_LockUnlock() {
	ctx := ts.T().Context()
	key := path.Join("acme", "acme-v02.api.example.com", "sites", "example.com", "example.com.crt")

	// Test successful lock acquisition
	err := ts.dbs.Lock(ctx, key)
	ts.Require().NoError(err)

	// Test successful unlock
	err = ts.dbs.Unlock(ctx, key)
	ts.Require().NoError(err)
}

func (ts *StorageTestSuite) TestStorage_TwoLocks() {
	ctx := ts.T().Context()
	key := path.Join("acme", "acme-v02.api.example.com", "sites", "example.com", "example.com.crt")

	// Create a second storage instance using the same DynamoDB Table
	dbs2 := storage.NewDynamoDBStorage()
	dbs2.Table = ts.dbs.Table
	err := dbs2.Provision(redirtest.NewRedirectorCaddyContext(ts.T()))
	ts.Require().NoError(err)

	// Test successful lock acquisition
	err = ts.dbs.Lock(ctx, key)
	ts.Require().NoError(err)

	// Test concurrent lock attempts
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		err = dbs2.Lock(ctx, key)
		ts.Error(err)
	}()
	wg.Wait()

	// Test successful unlock
	err = ts.dbs.Unlock(ctx, key)
	ts.Require().NoError(err)

	// Verify lock can be acquired after unlock
	err = dbs2.Lock(ctx, key)
	ts.Require().NoError(err)
	err = dbs2.Unlock(ctx, key)
	ts.Require().NoError(err)
}
