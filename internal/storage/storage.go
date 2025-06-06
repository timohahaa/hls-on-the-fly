package storage

import (
	"errors"
	"os"
)

const basePath = "/home/timohahaa/work/kinescope/sandbox/hls-on-the-fly/"

type Storage struct {
	files map[string]map[string]string
}

func New() (*Storage, error) {
	return &Storage{
		files: map[string]map[string]string{
			"video-1": {
				"360":   basePath + "testdata/test-360.mp4",
				"480":   basePath + "testdata/test-480.mp4",
				"720":   basePath + "testdata/test-720.mp4",
				"audio": basePath + "testdata/test-audio.mp4",
			},
		},
	}, nil
}

func (s *Storage) GetFileAsset(fileName, quality string) (*os.File, error) {
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
