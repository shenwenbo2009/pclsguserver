package store

import (
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
)

type FsStore struct {
	root string
}

func NewFsStore(root string) (*FsStore, error) {
	if root == "" {
		root = "./data"
	}
	if err := os.MkdirAll(root, 0o755); err != nil {
		return nil, err
	}
	return &FsStore{root: root}, nil
}

func (s *FsStore) full(path string) string {
	return filepath.Join(s.root, filepath.Clean("/"+path))
}

func (s *FsStore) Save(ctx context.Context, path string, reader io.Reader) (int64, error) {
	full := s.full(path)
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		return 0, err
	}
	f, err := os.Create(full)
	if err != nil {
		return 0, err
	}
	defer f.Close()
	return io.Copy(f, reader)
}

func (s *FsStore) Read(ctx context.Context, path string) (io.ReadCloser, error) {
	return os.Open(s.full(path))
}

func (s *FsStore) ReadRange(ctx context.Context, path string, offset, length int64) (io.ReadCloser, error) {
	f, err := os.Open(s.full(path))
	if err != nil {
		return nil, err
	}
	if _, err := f.Seek(offset, io.SeekStart); err != nil {
		f.Close()
		return nil, err
	}
	if length < 0 {
		return f, nil
	}
	return &limitedReadCloser{R: io.LimitReader(f, length), C: f}, nil
}

func (s *FsStore) Stat(ctx context.Context, path string) (int64, error) {
	info, err := os.Stat(s.full(path))
	if err != nil {
		return 0, err
	}
	return info.Size(), nil
}

func (s *FsStore) Delete(ctx context.Context, path string) error {
	err := os.Remove(s.full(path))
	if err != nil && errors.Is(err, os.ErrNotExist) {
		return nil
	}
	return err
}

func (s *FsStore) Exists(ctx context.Context, path string) (bool, error) {
	_, err := os.Stat(s.full(path))
	if err == nil {
		return true, nil
	}
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	return false, err
}

type limitedReadCloser struct {
	R io.Reader
	C io.Closer
}

func (l *limitedReadCloser) Read(p []byte) (int, error) {
	return l.R.Read(p)
}

func (l *limitedReadCloser) Close() error {
	return l.C.Close()
}
