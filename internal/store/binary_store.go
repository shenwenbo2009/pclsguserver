package store

import (
	"context"
	"io"
)

type BinaryStore interface {
	Save(ctx context.Context, path string, reader io.Reader) (int64, error)
	Read(ctx context.Context, path string) (io.ReadCloser, error)
	ReadRange(ctx context.Context, path string, offset, length int64) (io.ReadCloser, error)
	Stat(ctx context.Context, path string) (int64, error)
	Delete(ctx context.Context, path string) error
	Exists(ctx context.Context, path string) (bool, error)
}

func SwPath(swCode, version string) string {
	return "sw/" + swCode + "/" + version + "/package"
}

func PluginPath(id, version string) string {
	return "plugin/" + id + "/" + version + "/package"
}
func FilePath(purpose, uuid, filename string) string {
	return "file/" + purpose + "/" + uuid + "/" + filename
}
