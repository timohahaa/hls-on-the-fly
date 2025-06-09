package manifest

import (
	"math"
	"os"

	"github.com/grafov/m3u8"
	"github.com/timohahaa/hls-on-the-fly/internal/storage"
)

func Master(assets []storage.Asset, baseURL string) ([]byte, error) {
	var (
		master = m3u8.NewMasterPlaylist()
		audio  *storage.Asset
	)

	for _, asset := range assets {
		if asset.Quality == "audio" {
			audio = &asset
			break
		}
	}

	master.SetVersion(6)
	master.SetIndependentSegments(true)
	master.Variants = make([]*m3u8.Variant, 0, len(assets))

	for _, asset := range assets {

		if asset.Quality == "audio" {
			continue
		}

		var fileSize int64
		{
			fd, err := os.Open(asset.FilePath)
			if err != nil {
				return nil, err
			}

			info, err := fd.Stat()
			if err != nil {
				return nil, err
			}

			fileSize = info.Size()
		}

		var url string = baseURL + "/" + asset.Quality + "/media.m3u8"

		variant := &m3u8.Variant{
			URI: url,
			VariantParams: m3u8.VariantParams{
				// The PROGRAM-ID attribute of the EXT-X-STREAM-INF
				// and the EXT-X-I-FRAME-STREAM-INF tags was removed in protocol version 6.
				// source: https://datatracker.ietf.org/doc/html/rfc8216
				//ProgramId:        s.ProgramId,
				Audio:      "audio/mp4a",
				Bandwidth:  estimateBandwidth(fileSize, asset.Duration),
				Codecs:     asset.Codec,
				Resolution: asset.Resolution,
				FrameRate:  asset.FPS,
			},
		}

		if audio != nil {
			var audioURL string = baseURL + "/" + audio.Quality + "/media.m3u8"
			variant.Alternatives = append(variant.Alternatives, &m3u8.Alternative{
				Type:     "AUDIO",
				GroupId:  "audio/mp4a",
				Language: "und",
				Name:     "und",
				Default:  true,
				URI:      audioURL,
			})
		}

		master.Variants = append(master.Variants, variant)
	}

	return master.Encode().Bytes(), nil
}

func estimateBandwidth(fileSize int64, duration float64) uint32 {
	// нужно на 8 умножить, потому что размер файла в байтах, а bandwidth в битах :)
	return uint32(math.Round(float64(fileSize)/duration)) * 8
}
