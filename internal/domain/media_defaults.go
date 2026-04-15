package domain

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"
)

// MediaDefaults stores the user's preferred settings for adding new media items.
// These are saved and loaded to pre-populate the Add Media dialog.
type MediaDefaults struct {
	Quality      VideoQuality `json:"quality"`
	DownloadPath string       `json:"download_path"`
	OnlyAudio    bool         `json:"only_audio"`
	Cookies      Cookies      `json:"cookies,omitempty"`
	Subtitle     Subtitle     `json:"subtitle,omitempty"`
}

func getMediaDefaultsFilePath() string {
	configDir, err := os.UserConfigDir()
	if err != nil {
		log.Printf("Error getting config dir: %v", err)
		return "byto_media_defaults.json"
	}

	bytoDir := filepath.Join(configDir, "byto")
	if err := os.MkdirAll(bytoDir, 0755); err != nil {
		log.Printf("Error creating config dir: %v", err)
		return "byto_media_defaults.json"
	}

	return filepath.Join(bytoDir, "media_defaults.json")
}

func NewMediaDefaults() *MediaDefaults {
	defaults := LoadMediaDefaults()
	if defaults != nil {
		return defaults
	}

	return &MediaDefaults{
		Quality:      Quality1080p,
		DownloadPath: getDefaultDownloadPath(),
	}
}

func LoadMediaDefaults() *MediaDefaults {
	filePath := getMediaDefaultsFilePath()

	data, err := os.ReadFile(filePath)
	if err != nil {
		if !os.IsNotExist(err) {
			log.Printf("Error reading media defaults file: %v", err)
		}
		return nil
	}

	var defaults MediaDefaults
	if err := json.Unmarshal(data, &defaults); err != nil {
		log.Printf("Error parsing media defaults file: %v", err)
		return nil
	}

	log.Printf("Loaded media defaults from %s", filePath)
	return &defaults
}

func (m *MediaDefaults) Save() error {
	filePath := getMediaDefaultsFilePath()

	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}

	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return err
	}

	log.Printf("Media defaults saved to %s", filePath)
	return nil
}

func (m *MediaDefaults) Update(quality VideoQuality, downloadPath string, onlyAudio bool, cookies Cookies, subtitle Subtitle) {
	m.Quality = quality
	m.DownloadPath = downloadPath
	m.OnlyAudio = onlyAudio
	m.Cookies = cookies
	m.Subtitle = subtitle
}
