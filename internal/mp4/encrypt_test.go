package mp4_test

import (
	"encoding/hex"
	"os"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/timohahaa/hls-on-the-fly/internal/mp4"
)

func TestEncrypt(t *testing.T) {
	path := "../../testdata/test-720.mp4"
	outPath := "test-720-encrypted.mp4"

	params := mp4.EncryptParams{
		KeyID:  uuid.MustParse("613aa5a4-cb22-491f-89b8-583a5432046a"),
		Key:    uuid.MustParse("4047b82f-a25e-4a58-8c2b-116dbcf81660"),
		IVHex:  strings.ReplaceAll(uuid.MustParse("9184bb2f-cf97-4226-b3a3-8ed8f8b3fe2e").String(), "-", ""),
		Scheme: "cbcs",
	}

	t.Logf("KeyID: %v, Key: %v, IV: %v",
		hex.EncodeToString(params.KeyID[:]),
		hex.EncodeToString(params.Key[:]),
		params.IVHex,
	)

	in, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}

	out, err := os.Create(outPath)
	if err != nil {
		t.Fatal(err)
	}

	if err := mp4.Encrypt(in, out, params); err != nil {
		t.Fatal(err)
	}
}
