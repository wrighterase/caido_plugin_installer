package utils

import (
	"encoding/json"
	"os"
)

type Author struct {
	Name  string `json:"name"`
	Email string `json:"email"`
	URL   string `json:"url"`
}

type Backend struct {
	ID string `json:"id"`
}
type Frontend struct {
	ID string `json:"id"`
}

type UUID struct {
	Frontend string `json:"fe-uuid"`
	Backend  string `json:"be-uuid"`
	PluginID string `json:"pi-uuid"`
}
type Plugin struct {
	ID         string    `json:"id"`
	Kind       string    `json:"kind"`
	Name       string    `json:"name"`
	Entrypoint string    `json:"entrypoint"`
	Runtime    string    `json:"runtime,omitempty"`
	Style      string    `json:"style,omitempty"`
	Backend    *Backend  `json:"backend,omitempty"`
	Frontend   *Frontend `json:"frontend,omitempty"`
}

type Manifest struct {
	ID          string         `json:"id"`
	Name        string         `json:"name"`
	Version     string         `json:"version"`
	Description string         `json:"description"`
	Author      Author         `json:"author"`
	Links       map[string]any `json:"links"`
	Plugins     []Plugin       `json:"plugins"`
	UUID        UUID           `json:"uuid"`
}

func ParseManifest(path string) (*Manifest, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var m Manifest
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, err
	}
	return &m, nil
}
