package storage

import (
	"errors"
	"io"
	"os"
)

type Storage struct {
	files map[string]map[string]string
}

func New() (*Storage, error) {
	return &Storage{
		files: map[string]map[string]string{
			"video-1": {
				"360": "",
				"480": "",
				"720": "",
			},
		},
	}, nil
}

func (s *Storage) GetFileAsset(fileName, quality string) (io.Reader, error) {
	assets, ok := s.files[fileName]
	if !ok {
		return nil, errors.New("file not found")
	}

	qualityAsset, ok := assets[quality]
	if !ok {
		return nil, errors.New("file quality not found")
	}

	return os.Open(qualityAsset)
}
