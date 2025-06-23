package main

import (
	"os"
	"testing"

	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/NordSecurity/nordvpn-linux/test/category"
	"github.com/stretchr/testify/assert"
)

const TestDataPath = "testdata/"

func TestCreateMooseDB(t *testing.T) {
	category.Set(t, category.Unit)

	path := TestDataPath + "test_moose_db.db"
	const perm os.FileMode = internal.PermUserRWGroupRW
	defer os.Remove(path)

	checkPermissionsFn := func() {
		info, err := os.Stat(path)
		assert.NoError(t, err)
		assert.Equal(t, perm, info.Mode().Perm())
	}

	// create DB
	assignMooseDBPermissions(path)

	assert.FileExists(t, path)
	checkPermissionsFn()

	// file exists, check that the permissions are updated
	os.Chmod(path, internal.PermUserRW)
	assignMooseDBPermissions(path)

	checkPermissionsFn()
}
