package storage

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// Storage saves uploaded files and removes them. Implementations decide where
// the bytes live; callers only keep the returned filename.
type Storage interface {
	Save(ext string, r io.Reader) (filename string, err error)
	Delete(filename string) error
	URL(filename string) string
}

// Local stores files on the local filesystem under Dir and serves them at
// the URL prefix Prefix.
type Local struct {
	Dir    string
	Prefix string
}

func NewLocal(dir, prefix string) (*Local, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, err
	}
	return &Local{Dir: dir, Prefix: prefix}, nil
}

func (l *Local) Save(ext string, r io.Reader) (string, error) {
	name, err := randomName()
	if err != nil {
		return "", err
	}
	filename := name + ext

	f, err := os.Create(filepath.Join(l.Dir, filename))
	if err != nil {
		return "", err
	}
	defer f.Close()

	if _, err := io.Copy(f, r); err != nil {
		os.Remove(filepath.Join(l.Dir, filename))
		return "", err
	}
	return filename, nil
}

func (l *Local) Delete(filename string) error {
	err := os.Remove(filepath.Join(l.Dir, filepath.Base(filename)))
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

func (l *Local) URL(filename string) string {
	return fmt.Sprintf("%s/%s", l.Prefix, filename)
}

func randomName() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
