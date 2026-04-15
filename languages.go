package main

import (
	"encoding/json"

	_ "embed"
)

type languageEntry struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

//go:embed languages.json
var availableLanguagesJSON []byte

var availableLanguages = mustLoadAvailableLanguages(availableLanguagesJSON)

func mustLoadAvailableLanguages(raw []byte) map[string]string {
	var entries []languageEntry
	if err := json.Unmarshal(raw, &entries); err != nil {
		panic("failed to parse embedded languages.json: " + err.Error())
	}

	languages := make(map[string]string, len(entries))
	for _, entry := range entries {
		if entry.Code == "" || entry.Name == "" {
			continue
		}
		languages[entry.Code] = entry.Name
	}

	return languages
}
