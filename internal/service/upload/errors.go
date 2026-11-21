package upload

import "errors"

var (
	ErrInvalidParam         = errors.New("INVALID_PARAM")
	ErrInvalidFileType      = errors.New("INVALID_FILE_TYPE")
	ErrFileTooLarge         = errors.New("FILE_TOO_LARGE")
	ErrPluginAlreadyExists  = errors.New("PLUGIN_ALREADY_EXISTS")
	ErrVersionAlreadyExists = errors.New("VERSION_ALREADY_EXISTS")
	ErrInvalidFilename      = errors.New("INVALID_FILENAME")
	ErrInvalidPurpose       = errors.New("INVALID_PARAM")
)
