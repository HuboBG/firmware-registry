package firmware

import "time"

// Firmware is the internal domain model.
type Firmware struct {
	Type      string
	Version   string
	Filename  string
	SizeBytes int64
	SHA256    string
	CreatedAt time.Time
}

// FirmwareDTO is what we expose over HTTP.
type FirmwareDTO struct {
	Type        string    `json:"type" example:"esp32-main" doc:"Firmware type identifier"`
	Version     string    `json:"version" example:"1.2.3" doc:"Semantic version"`
	Filename    string    `json:"filename" example:"firmware.bin" doc:"Original filename"`
	SizeBytes   int64     `json:"sizeBytes" example:"524288" doc:"File size in bytes"`
	SHA256      string    `json:"sha256" example:"abc123..." doc:"SHA256 checksum"`
	CreatedAt   time.Time `json:"createdAt" example:"2024-01-15T10:30:00Z" doc:"Upload timestamp"`
	DownloadURL string    `json:"downloadUrl,omitempty" example:"http://localhost:8080/api/firmware/esp32-main/1.2.3" doc:"Direct download URL"`
}

func (f Firmware) ToDTO(downloadURL string) FirmwareDTO {
	return FirmwareDTO{
		Type:        f.Type,
		Version:     f.Version,
		Filename:    f.Filename,
		SizeBytes:   f.SizeBytes,
		SHA256:      f.SHA256,
		CreatedAt:   f.CreatedAt,
		DownloadURL: downloadURL,
	}
}
