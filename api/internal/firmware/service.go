package firmware

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"os"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
)

// Repository persists firmware metadata.
type Repository interface {
	Upsert(Firmware) error
	Get(typeName, version string) (Firmware, error)
	List(typeName string) ([]Firmware, error)
	Delete(typeName, version string) error
}

// Service holds business logic only.
type Service struct {
	Repo       Repository
	Storage    Storage
	PublicBase string
}

// SaveFirmware reads the uploaded binary, computes SHA256,
// writes to disk atomically, and upserts metadata.
func (s *Service) SaveFirmware(typeName, version, filename string, r io.Reader) (Firmware, error) {
	log.Info().
		Str("type", typeName).
		Str("version", version).
		Str("filename", filename).
		Msg("Starting firmware upload")

	data, err := io.ReadAll(r)
	if err != nil {
		log.Error().
			Err(err).
			Str("type", typeName).
			Str("version", version).
			Msg("Failed to read firmware data")
		return Firmware{}, err
	}

	sum := sha256.Sum256(data)
	shaHex := hex.EncodeToString(sum[:])

	log.Debug().
		Str("type", typeName).
		Str("version", version).
		Int64("size_bytes", int64(len(data))).
		Str("sha256", shaHex).
		Msg("Firmware SHA256 computed")

	dir := s.Storage.Dir(typeName, version)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		log.Error().
			Err(err).
			Str("type", typeName).
			Str("version", version).
			Str("dir", dir).
			Msg("Failed to create storage directory")
		return Firmware{}, err
	}

	dest := s.Storage.FilePath(typeName, version)
	tmp := dest + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		log.Error().
			Err(err).
			Str("type", typeName).
			Str("version", version).
			Str("tmp_file", tmp).
			Msg("Failed to write temporary firmware file")
		return Firmware{}, err
	}
	if err := os.Rename(tmp, dest); err != nil {
		log.Error().
			Err(err).
			Str("type", typeName).
			Str("version", version).
			Str("tmp_file", tmp).
			Str("dest_file", dest).
			Msg("Failed to rename firmware file (atomic write)")
		return Firmware{}, err
	}

	log.Debug().
		Str("type", typeName).
		Str("version", version).
		Str("file", dest).
		Msg("Firmware file written to storage")

	rec := Firmware{
		Type:      typeName,
		Version:   version,
		Filename:  filename,
		SizeBytes: int64(len(data)),
		SHA256:    shaHex,
		CreatedAt: time.Now().UTC(),
	}

	if err := s.Repo.Upsert(rec); err != nil {
		log.Error().
			Err(err).
			Str("type", typeName).
			Str("version", version).
			Msg("Failed to upsert firmware metadata to database")
		return Firmware{}, err
	}

	log.Info().
		Str("type", typeName).
		Str("version", version).
		Str("filename", filename).
		Int64("size_bytes", rec.SizeBytes).
		Str("sha256", rec.SHA256).
		Msg("Firmware uploaded successfully")

	return rec, nil
}

func (s *Service) DownloadPath(typeName, version string) string {
	return s.Storage.FilePath(typeName, version)
}

func (s *Service) DownloadURL(typeName, version string) string {
	if s.PublicBase == "" {
		return ""
	}
	base := strings.TrimRight(s.PublicBase, "/")
	return base + "/api/firmware/" + typeName + "/" + version
}
