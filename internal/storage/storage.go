package storage

import (
	"errors"
	"maps"
	"os"
	"slices"
)

const basePath = "/home/timohahaa/work/kinescope/sandbox/hls-on-the-fly/"

type Storage struct {
	files map[string]map[string]Asset
}

func New() (*Storage, error) {
	return &Storage{
		files: map[string]map[string]Asset{
			"video-1": {
				//"audio": basePath + "testdata/test-audio.mp4",
				"360": Asset{
					Quality:    "360",
					Resolution: "640x360",
					FPS:        30,
					Codec:      "avc1.4D401E",
					Duration:   82,
					FilePath:   basePath + "testdata/test-360.mp4",
				},
				"480": Asset{
					Quality:    "480",
					Resolution: "854x480",
					FPS:        30,
					Codec:      "avc1.4D401F",
					Duration:   82,
					FilePath:   basePath + "testdata/test-480.mp4",
				},
				"720": Asset{
					Quality:    "720",
					Resolution: "1280x720",
					FPS:        30,
					Codec:      "avc1.4D401F",
					Duration:   82,
					FilePath:   basePath + "testdata/test-720.mp4",
				},
				"audio": Asset{
					Quality:    "audio",
					Resolution: "none",
					FPS:        0,
					Codec:      "mp4a.40.2",
					Duration:   82,
					FilePath:   basePath + "testdata/test-audio.mp4",
				},
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

	return os.Open(qualityAsset.FilePath)
}

func (s *Storage) GetAllAssets(fileName string) ([]Asset, error) {
	assets, ok := s.files[fileName]
	if !ok {
		return nil, errors.New("file not found")
	}
	return slices.Collect(maps.Values(assets)), nil
}
