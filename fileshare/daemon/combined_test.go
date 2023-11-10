package daemon

import (
	"testing"

	"github.com/NordSecurity/nordvpn-linux/test/category"
	"github.com/stretchr/testify/assert"
)

func TestUsesMainIfNoError(t *testing.T) {
	category.Set(t, category.Unit)

	main := &TestFileshare{}
	backup := &TestFileshare{}
	combined := NewCombinedFileshare(main, backup)

	// Disable
	assert.NoError(t, combined.Enable(0, 0))
	assert.True(t, main.enabled)
	assert.False(t, backup.enabled)
	assert.NoError(t, combined.Disable(0, 0))
	assert.False(t, main.enabled)
	assert.False(t, backup.enabled)

	// Stop
	assert.NoError(t, combined.Enable(0, 0))
	assert.True(t, main.enabled)
	assert.False(t, backup.enabled)
	assert.NoError(t, combined.Stop(0, 0))
	assert.False(t, main.enabled)
	assert.False(t, backup.enabled)
}

func TestUsesBackupIfError(t *testing.T) {
	category.Set(t, category.Unit)

	main := &TestFileshare{doError: true}
	backup := &TestFileshare{}
	combined := NewCombinedFileshare(main, backup)

	// Disable
	assert.NoError(t, combined.Enable(0, 0))
	assert.False(t, main.enabled)
	assert.True(t, backup.enabled)
	assert.NoError(t, combined.Disable(0, 0))
	assert.False(t, main.enabled)
	assert.False(t, backup.enabled)

	// Stop
	assert.NoError(t, combined.Enable(0, 0))
	assert.False(t, main.enabled)
	assert.True(t, backup.enabled)
	assert.NoError(t, combined.Stop(0, 0))
	assert.False(t, main.enabled)
	assert.False(t, backup.enabled)
}
