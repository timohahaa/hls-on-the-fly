package manifest

import (
	"bytes"
	"errors"
	"io"
	"math"
	"strconv"

	"github.com/Eyevinn/mp4ff/mp4"
)

// TODO:
// - поддержка seek-time и duration (обрезка видео средствами манифеста) для автовебинаров (#EXT-X-MEDIA-SEQUENCE)
// - time_offset (#EXT-X-START:TIME-OFFSET=)
func Media(src io.Reader, fileURL string) ([]byte, error) {
	var (
		buf = make([]byte, 0, 4096)
		m3u = bytes.NewBuffer(buf)

		mp4f, err = mp4.DecodeFile(src)
	)
	if err != nil {
		return nil, err
	}

	if !mp4f.IsFragmented() {
		return nil, errors.New("source file is not fragmented")
	}

	if mp4f.Init == nil {
		return nil, errors.New("source file has no init segment")
	}

	m3u.WriteString("#EXTM3U\n")
	m3u.WriteString("#EXT-X-VERSION:6\n")
	m3u.WriteString("#EXT-X-INDEPENDENT-SEGMENTS\n")
	{
		m3u.WriteString("#EXT-X-TARGETDURATION:")
		m3u.WriteString(strconv.Itoa(getTargetDuration(mp4f)))
		m3u.WriteString("\n")
	}
	m3u.WriteString("#EXT-X-PLAYLIST-TYPE:VOD\n")

	// init segment
	{
		m3u.WriteString("#EXT-X-MAP:URI=\"")
		m3u.WriteString(fileURL)
		m3u.WriteString("\",BYTERANGE=\"")
		// byte range always starts at 0 I guess
		m3u.WriteString("0@")
		m3u.WriteString(strconv.FormatInt(int64(mp4f.Init.Size()), 10))

		m3u.WriteString("\"")

		// TODO: add encryption params below
	}

	m3u.WriteString("#EXT-X-MEDIA-SEQUENCE:0\n")

	// fragmtents

	return nil, nil
}

type fragmentInfo struct {
	duration float64
	startPos int64
	size     int64
}

func getFragmentInfo(file *mp4.File) ([]fragmentInfo, error) {
	var timescale = file.Moov.Mvhd.Timescale

	// инвариант: в 1-м mp4-сегменте 1 mp4-фрагмент
	for _, s := range file.Segments {
		for _, f := range s.Fragments {

		}
	}
}

func getTargetDuration(file *mp4.File) int {
	var (
		dur = float64(file.Moov.Mvhd.Duration)
		tsc = float64(file.Moov.Mvhd.Timescale)
	)
	return int(math.Ceil(dur / tsc))
}
