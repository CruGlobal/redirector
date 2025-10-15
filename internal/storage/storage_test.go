package storage_test

import (
	"bytes"
	"path"
	"testing"

	"github.com/CruGlobal/redirector/internal/storage"
	"github.com/CruGlobal/redirector/redirtest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type StorageTestSuite struct {
	suite.Suite

	dbs *storage.DynamoDBStorage
}

func (ts *StorageTestSuite) SetupSuite() {
	ctx := redirtest.NewRedirectorCaddyContext(ts.T())

	dbs := storage.NewDynamoDBStorage()
	dbs.Table = "RedirectorCertsTesting"
	err := dbs.Provision(ctx)
	ts.Require().NoError(err)

	ts.dbs = dbs
}

// SetupTest creates the table before each test.
func (ts *StorageTestSuite) SetupTest() {
	redirtest.CreateDynamoDBTable(ts.T(), ts.dbs.Client, ts.dbs.Table, "Key")
}

// TearDownTest deletes the table after each test.
func (ts *StorageTestSuite) TearDownTest() {
	redirtest.DeleteDynamoDBTable(ts.T(), ts.dbs.Client, ts.dbs.Table)
}

func TestStorageTestSuite(t *testing.T) {
	suite.Run(t, new(StorageTestSuite))
}

type args struct {
	key   string
	value []byte
}

func (ts *StorageTestSuite) TestStorage_Store() {
	testcases := []struct {
		name      string
		args      args
		expectErr bool
	}{
		{
			name:      "missing key",
			expectErr: true,
		},
		{
			name: "missing value",
			args: args{
				key: path.Join("acme", "acme-v02.api.example.com", "sites", "example.com", "example.com.crt"),
			},
		},
		{
			name: "key/value",
			args: args{
				key:   path.Join("acme", "acme-v02.api.example.com", "sites", "example.org", "example.org.crt"),
				value: bytes.Repeat([]byte("b"), 4096),
			},
		},
	}
	for _, tc := range testcases {
		ts.Run(tc.name, func() {
			t := ts.T()
			err := ts.dbs.Store(t.Context(), tc.args.key, tc.args.value)
			if tc.expectErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func (ts *StorageTestSuite) TestStorage_Exists() {
	testcases := []struct {
		name   string
		seed   string
		key    string
		expect bool
	}{
		{
			name:   "missing key",
			expect: false,
		},
		{
			name:   "exists",
			seed:   path.Join("acme", "acme-v02.api.example.com", "sites", "example.org", "example.org.crt"),
			key:    path.Join("acme", "acme-v02.api.example.com", "sites", "example.org", "example.org.crt"),
			expect: true,
		},
		{
			name:   "does not exist",
			seed:   path.Join("acme", "acme-v02.api.example.com", "sites", "example.com", "example.com.crt"),
			key:    path.Join("acme", "acme-v02.api.example.com", "sites", "example.org", "example.org.key"),
			expect: false,
		},
		{
			name:   "path",
			seed:   path.Join("acme", "acme-v02.api.example.com", "sites", "example.net", "example.net.crt"),
			key:    path.Join("acme", "acme-v02.api.example.com", "sites", "example.net"),
			expect: false,
		},
	}
	for _, tc := range testcases {
		ts.Run(tc.name, func() {
			t := ts.T()
			if tc.seed != "" {
				err := ts.dbs.Store(t.Context(), tc.seed, bytes.Repeat([]byte("c"), 4096))
				require.NoError(t, err)
			}
			exists := ts.dbs.Exists(t.Context(), tc.key)
			assert.Equal(t, tc.expect, exists)
		})
	}
}

func (ts *StorageTestSuite) TestStorage_Load() {
	testcases := []struct {
		name      string
		args      args
		key       string
		expectErr bool
	}{
		{
			name:      "missing key",
			expectErr: true,
		},
		{
			name: "exists",
			args: args{
				key:   path.Join("acme", "acme-v02.api.example.com", "sites", "example.com", "example.com.crt"),
				value: bytes.Repeat([]byte("e"), 4096),
			},
			key: path.Join("acme", "acme-v02.api.example.com", "sites", "example.com", "example.com.crt"),
		},
		{
			name: "not found",
			args: args{
				key:   path.Join("acme", "acme-v02.api.example.com", "sites", "example.net", "example.net.crt"),
				value: bytes.Repeat([]byte("e"), 4096),
			},
			key:       path.Join("acme", "acme-v02.api.example.com", "sites", "example.net"),
			expectErr: true,
		},
	}
	for _, tc := range testcases {
		ts.Run(tc.name, func() {
			t := ts.T()
			if tc.args.key != "" {
				err := ts.dbs.Store(t.Context(), tc.args.key, tc.args.value)
				require.NoError(t, err)
			}
			content, err := ts.dbs.Load(t.Context(), tc.key)
			if tc.expectErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.args.value, content)
			}
		})
	}
}

func (ts *StorageTestSuite) TestStorage_Delete() {
	testcases := []struct {
		name      string
		args      args
		key       string
		expectErr bool
	}{
		{
			name:      "missing key",
			expectErr: true,
		},
		{
			name: "exists",
			args: args{
				key:   path.Join("acme", "acme-v02.api.example.com", "sites", "example.com", "example.com.crt"),
				value: bytes.Repeat([]byte("f"), 4096),
			},
			key: path.Join("acme", "acme-v02.api.example.com", "sites", "example.com", "example.com.crt"),
		},
		{
			name: "not found",
			args: args{
				key:   path.Join("acme", "acme-v02.api.example.com", "sites", "example.net", "example.net.crt"),
				value: bytes.Repeat([]byte("e"), 4096),
			},
			key: path.Join("acme", "acme-v02.api.example.com", "sites", "example.net"),
		},
	}
	for _, tc := range testcases {
		ts.Run(tc.name, func() {
			t := ts.T()
			if tc.args.key != "" {
				err := ts.dbs.Store(t.Context(), tc.args.key, tc.args.value)
				require.NoError(t, err)
			}
			err := ts.dbs.Delete(t.Context(), tc.key)
			if tc.expectErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
			exists := ts.dbs.Exists(t.Context(), tc.key)
			assert.False(t, exists)
		})
	}
}

func (ts *StorageTestSuite) TestStorage_Stat() {
	testcases := []struct {
		name      string
		args      args
		key       string
		expectErr bool
	}{
		{
			name:      "missing key",
			expectErr: true,
		},
		{
			name: "exists",
			args: args{
				key:   path.Join("acme", "acme-v02.api.example.com", "sites", "example.com", "example.com.crt"),
				value: bytes.Repeat([]byte("g"), 4096),
			},
			key: path.Join("acme", "acme-v02.api.example.com", "sites", "example.com", "example.com.crt"),
		},
		{
			name: "not found",
			args: args{
				key:   path.Join("acme", "acme-v02.api.example.com", "sites", "example.net", "example.net.crt"),
				value: bytes.Repeat([]byte("h"), 4096),
			},
			key:       path.Join("acme", "acme-v02.api.example.com", "sites", "example.net"),
			expectErr: true,
		},
	}
	for _, tc := range testcases {
		ts.Run(tc.name, func() {
			t := ts.T()
			if tc.args.key != "" {
				err := ts.dbs.Store(t.Context(), tc.args.key, tc.args.value)
				require.NoError(t, err)
			}
			info, err := ts.dbs.Stat(t.Context(), tc.key)
			if tc.expectErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.args.key, info.Key)
				assert.Equal(t, int64(len(tc.args.value)), info.Size)
			}
		})
	}
}

func (ts *StorageTestSuite) TestStorage_List() {
	seed := []args{
		{
			key:   path.Join("acme", "acme-v02.api.example.com", "sites", "example.com", "example.com.crt"),
			value: bytes.Repeat([]byte("a"), 4096),
		},
		{
			key:   path.Join("acme", "acme-v02.api.example.com", "sites", "example.com", "example.com.key"),
			value: bytes.Repeat([]byte("b"), 4096),
		},
		{
			key:   path.Join("acme", "acme-v02.api.example.com", "sites", "example.com", "example.com.json"),
			value: []byte(`{"example: "name"}`),
		},
		{
			key:   path.Join("acme", "acme-v02.api.example.com", "sites", "example.org", "example.org.crt"),
			value: bytes.Repeat([]byte("b"), 4096),
		},
	}
	for _, arg := range seed {
		err := ts.dbs.Store(ts.T().Context(), arg.key, arg.value)
		ts.Require().NoError(err)
	}

	testcases := []struct {
		name      string
		path      string
		recursive bool
		keys      []string
		expectErr bool
	}{
		{
			name:      "missing path",
			expectErr: true,
		},
		{
			name:      "non-recursive example.com",
			path:      path.Join("acme", "acme-v02.api.example.com", "sites", "example.com"),
			recursive: false,
			keys: []string{
				seed[0].key,
				seed[1].key,
				seed[2].key,
			},
		},
		{
			name:      "non-recursive exact file",
			path:      seed[1].key,
			recursive: false,
			expectErr: true,
		},
		{
			name:      "non-recursive sites",
			path:      path.Join("acme", "acme-v02.api.example.com", "sites"),
			recursive: false,
			expectErr: true,
		},
		{
			name:      "recursive example.com",
			path:      path.Join("acme", "acme-v02.api.example.com", "sites", "example.com"),
			recursive: true,
			keys: []string{
				seed[0].key,
				seed[1].key,
				seed[2].key,
			},
		},
		{
			name:      "recursive sites",
			path:      path.Join("acme", "acme-v02.api.example.com", "sites"),
			recursive: true,
			keys: []string{
				seed[0].key,
				seed[1].key,
				seed[2].key,
				seed[3].key,
			},
		},
	}
	for _, tc := range testcases {
		ts.Run(tc.name, func() {
			t := ts.T()
			keys, err := ts.dbs.List(t.Context(), tc.path, tc.recursive)
			if tc.expectErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Len(t, keys, len(tc.keys))
				assert.ElementsMatch(t, tc.keys, keys)
			}
		})
	}
}
