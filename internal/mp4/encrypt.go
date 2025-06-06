package mp4

import (
	"encoding/hex"
	"fmt"
	"io"

	"github.com/Eyevinn/mp4ff/mp4"
	"github.com/google/uuid"
)

type EncryptParams struct {
	KeyID  uuid.UUID
	Key    uuid.UUID
	IVHex  string
	Scheme string
}

// src should be already fragmented mp4 file
func Encrypt(src io.Reader, dst io.Writer, params EncryptParams) error {
	if params.Scheme != "cenc" && params.Scheme != "cbcs" {
		return fmt.Errorf("scheme must be cenc or cbcs: %s", params.Scheme)
	}

	if len(params.IVHex) != 32 && len(params.IVHex) != 16 {
		return fmt.Errorf("hex iv must have length 16 or 32 chars: %d", len(params.IVHex))
	}

	iv, err := hex.DecodeString(params.IVHex)
	if err != nil {
		return fmt.Errorf("invalid iv %s", params.IVHex)
	}

	inFile, err := mp4.DecodeFile(src)
	if err != nil {
		return err
	}

	if inFile.Init == nil {
		return fmt.Errorf("input does not contain init segment")
	}

	var ipd *mp4.InitProtectData
	ipd, err = mp4.InitProtect(inFile.Init, params.Key[:], iv, params.Scheme, mp4.UUID(params.KeyID[:]), nil)
	if err != nil {
		return fmt.Errorf("init protect: %w", err)
	}

	for _, s := range inFile.Segments {
		for _, f := range s.Fragments {
			err = mp4.EncryptFragment(f, params.Key[:], iv, ipd)
			if err != nil {
				return fmt.Errorf("encrypt fragment: %w", err)
			}
		}
	}

	return inFile.Encode(dst)
}
