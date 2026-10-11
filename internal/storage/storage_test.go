package storage

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLocalSaveAndDelete(t *testing.T) {
	dir := t.TempDir()
	local, err := NewLocal(dir, "/uploads")
	require.NoError(t, err)

	name, err := local.Save(".jpg", strings.NewReader("image data"))
	require.NoError(t, err)
	assert.True(t, strings.HasSuffix(name, ".jpg"))
	assert.Len(t, name, 32+len(".jpg"), "name is 16 random bytes in hex plus the extension")

	data, err := os.ReadFile(filepath.Join(dir, name))
	require.NoError(t, err)
	assert.Equal(t, "image data", string(data))

	require.NoError(t, local.Delete(name))
	_, err = os.Stat(filepath.Join(dir, name))
	assert.True(t, os.IsNotExist(err))
}

func TestLocalSaveProducesUniqueNames(t *testing.T) {
	local, err := NewLocal(t.TempDir(), "/uploads")
	require.NoError(t, err)

	a, err := local.Save(".png", strings.NewReader("a"))
	require.NoError(t, err)
	b, err := local.Save(".png", strings.NewReader("b"))
	require.NoError(t, err)
	assert.NotEqual(t, a, b)
}

func TestLocalDeleteMissingFileIsNotAnError(t *testing.T) {
	local, err := NewLocal(t.TempDir(), "/uploads")
	require.NoError(t, err)
	assert.NoError(t, local.Delete("does-not-exist.jpg"))
}

func TestLocalDeleteIgnoresDirectoryTraversal(t *testing.T) {
	root := t.TempDir()
	outside := filepath.Join(root, "outside.txt")
	require.NoError(t, os.WriteFile(outside, []byte("keep me"), 0o644))

	uploads := filepath.Join(root, "uploads")
	local, err := NewLocal(uploads, "/uploads")
	require.NoError(t, err)

	// Only the base name is used, so this removes uploads/outside.txt, never ../outside.txt.
	_ = local.Delete("../outside.txt")
	_, err = os.Stat(outside)
	assert.NoError(t, err, "a file outside the uploads folder must not be removed")
}

func TestLocalURL(t *testing.T) {
	local, err := NewLocal(t.TempDir(), "/uploads")
	require.NoError(t, err)
	assert.Equal(t, "/uploads/abc.webp", local.URL("abc.webp"))
}
