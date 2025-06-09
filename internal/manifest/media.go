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
		buf       = make([]byte, 0, 4096)
		m3u       = bytes.NewBuffer(buf)
		fragments []fragmentInfo

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

	if fragments, err = getFragmentInfo(mp4f); err != nil {
		return nil, err
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
		m3u.WriteString(strconv.FormatInt(int64(mp4f.Init.Size()), 10))
		// byte range always starts at 0 I guess
		m3u.WriteString("@0\"\n")

		// TODO: add encryption params below
	}

	m3u.WriteString("#EXT-X-MEDIA-SEQUENCE:0\n")

	// fragmtents
	{
		for _, frag := range fragments {
			m3u.WriteString("#EXTINF:")
			m3u.WriteString(strconv.FormatFloat(frag.duration, 'f', -1, 64))
			m3u.WriteString(",\n")

			m3u.WriteString("#EXT-X-BYTERANGE:")
			m3u.WriteString(strconv.FormatInt(frag.size, 10))
			m3u.WriteString("@")
			m3u.WriteString(strconv.FormatInt(frag.startPos, 10))
			m3u.WriteString("\n")

			var edgeLink = fileURL +
				"?from=" + strconv.FormatInt(frag.startPos, 10) +
				"?size=" + strconv.FormatInt(frag.size, 10)

			m3u.WriteString(edgeLink)
			m3u.WriteString("\n")
		}
	}

	m3u.WriteString("#EXT-X-ENDLIST")

	return m3u.Bytes(), nil
}

type fragmentInfo struct {
	duration float64
	startPos int64
	size     int64
}

func getFragmentInfo(file *mp4.File) ([]fragmentInfo, error) {
	var (
		timescale = file.Moov.Trak.Mdia.Mdhd.Timescale
		fragInfos = make([]fragmentInfo, 0, len(file.Segments))
	)

	// инвариант: в 1-м mp4-сегменте 1 mp4-фрагмент
	for _, s := range file.Segments {
		for _, f := range s.Fragments {

			var (
				defaultSampleDur = f.Moof.Traf.Tfhd.DefaultSampleDuration
				fragmentDur      = defaultSampleDur * f.Moof.Traf.Trun.SampleCount()
			)

			fragInfos = append(fragInfos, fragmentInfo{
				duration: float64(fragmentDur) / float64(timescale),
				startPos: int64(f.StartPos),
				size:     int64(f.Size()),
			})
		}
	}

	return fragInfos, nil
}

func getTargetDuration(file *mp4.File) int {
	var (
		dur = float64(file.Moov.Mvhd.Duration)
		tsc = float64(file.Moov.Mvhd.Timescale)
	)
	return int(math.Ceil(dur / tsc))
}
