package firmware

import (
	"database/sql"
	"time"

	"github.com/rs/zerolog/log"
)

// SQLiteRepo implements Repository over SQLite.
type SQLiteRepo struct {
	DB *sql.DB
}

func (r *SQLiteRepo) Upsert(f Firmware) error {
	_, err := r.DB.Exec(`
INSERT INTO firmwares(type, version, filename, size_bytes, sha256, created_at)
VALUES(?,?,?,?,?,?)
ON CONFLICT(type, version) DO UPDATE SET
  filename=excluded.filename,
  size_bytes=excluded.size_bytes,
  sha256=excluded.sha256,
  created_at=excluded.created_at
`, f.Type, f.Version, f.Filename, f.SizeBytes, f.SHA256, f.CreatedAt.Format(time.RFC3339))
	return err
}

func (r *SQLiteRepo) Get(typeName, version string) (Firmware, error) {
	var f Firmware
	var created string
	err := r.DB.QueryRow(`
SELECT type, version, filename, size_bytes, sha256, created_at
FROM firmwares WHERE type=? AND version=?
`, typeName, version).Scan(
		&f.Type, &f.Version, &f.Filename, &f.SizeBytes, &f.SHA256, &created,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Debug().
				Str("type", typeName).
				Str("version", version).
				Msg("Firmware not found in database")
		} else {
			log.Error().
				Err(err).
				Str("type", typeName).
				Str("version", version).
				Msg("Database error querying firmware")
		}
		return f, err
	}
	f.CreatedAt, _ = time.Parse(time.RFC3339, created)
	return f, nil
}

func (r *SQLiteRepo) List(typeName string) ([]Firmware, error) {
	rows, err := r.DB.Query(`
SELECT type, version, filename, size_bytes, sha256, created_at
FROM firmwares WHERE type=?
`, typeName)
	if err != nil {
		return nil, err
	}
	defer func(rows *sql.Rows) {
		_ = rows.Close()
	}(rows)

	var out []Firmware
	for rows.Next() {
		var f Firmware
		var created string
		if err := rows.Scan(&f.Type, &f.Version, &f.Filename, &f.SizeBytes, &f.SHA256, &created); err != nil {
			continue
		}
		f.CreatedAt, _ = time.Parse(time.RFC3339, created)
		out = append(out, f)
	}
	return out, nil
}

func (r *SQLiteRepo) Delete(typeName, version string) error {
	_, err := r.DB.Exec(`DELETE FROM firmwares WHERE type=? AND version=?`, typeName, version)
	return err
}
