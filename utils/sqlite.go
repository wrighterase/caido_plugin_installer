package utils

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

var homeDir, _ = os.UserHomeDir()

var pluginsDB = filepath.Join(homeDir, ".local", "share", "caido", "plugins.db")

func InsertPluginPackage(db *sql.DB, pkgID string, manifest *Manifest, origin string) error {
	author, _ := json.Marshal(manifest.Author)
	if author == nil {
		author = []byte("{}")
	}
	links, _ := json.Marshal(manifest.Links)
	if links == nil {
		links = []byte("{}")
	}

	name := manifest.Name
	if name == "" {
		name = "Unknown Plugin"
	}
	description := manifest.Description
	version := manifest.Version
	if version == "" {
		version = "1.0.0"
	}
	manifestID := manifest.ID
	if manifestID == "" {
		manifestID = "manual-install"
	}

	_, err := db.Exec(`
		INSERT INTO plugin_packages (id, manifest_id, name, description, author,
		                             version, installed_at, links, origin)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, pkgID, manifestID, name, description, string(author), version, time.Now().UTC().Format(time.RFC3339), string(links), origin)

	return err
}

func InsertPluginEntries(db *sql.DB, manifest *Manifest, plugins []Plugin) error {
	for _, p := range plugins {
		var metadata map[string]interface{}
		if p.Kind == "backend" {
			metadata = map[string]interface{}{
				"kind":       "backend",
				"runtime":    p.Runtime,
				"entrypoint": filepath.Base(p.Entrypoint),
				"assets":     nil,
			}
		} else {
			metadata = map[string]interface{}{
				"kind":       "frontend",
				"entrypoint": filepath.Base(p.Entrypoint),
				"style":      filepath.Base(p.Style),
				"assets":     nil,
			}
			if p.Backend != nil {
				metadata["backend"] = map[string]string{"manifest_id": p.Backend.ID}
			}
		}

		metaJSON, err := json.Marshal(metadata)
		if err != nil {
			return fmt.Errorf("failed to marshal metadata for %s: %w", p.ID, err)
		}

		pluginID := manifest.UUID.Backend
		if p.Kind == "frontend" {
			pluginID = manifest.UUID.Frontend
		}

		name := p.Name
		if name == "" {
			name = p.Kind
		}

		_, err = db.Exec(`
			INSERT INTO plugins (id, package_id, manifest_id, name, metadata)
			VALUES (?, ?, ?, ?, ?)
		`, pluginID, manifest.UUID.PluginID, p.Kind, name, string(metaJSON))
		if err != nil {
			return fmt.Errorf("failed to insert plugin entry %s: %w", p.ID, err)
		}
	}
	return nil
}

func InsertFrontendUserPrefs(db *sql.DB, pluginID string) error {
	rows, err := db.Query("SELECT DISTINCT user_id FROM plugins_frontend")
	if err != nil {
		return fmt.Errorf("failed to query frontend user prefs: %w", err)
	}
	defer rows.Close()

	var userIDs []string
	for rows.Next() {
		var userID string
		if err := rows.Scan(&userID); err != nil {
			return fmt.Errorf("failed to scan user_id: %w", err)
		}
		userIDs = append(userIDs, userID)
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("failed to iterate user_ids: %w", err)
	}

	for _, userID := range userIDs {
		_, err = db.Exec(`
			INSERT INTO plugins_frontend (user_id, plugin_id, enabled, data)
			VALUES (?, ?, 1, NULL)
		`, userID, pluginID)
		if err != nil {
			return fmt.Errorf("failed to insert frontend user prefs for %s: %w", userID, err)
		}
	}

	return nil
}

func InsertBackendEnabled(db *sql.DB, pluginID string) error {
	_, err := db.Exec(`
		INSERT INTO plugins_backend (plugin_id, enabled)
		VALUES (?, 1)
	`, pluginID)
	return err
}

func InsertWorkflows(db *sql.DB, plugins []Plugin) error {
	var maxID int
	err := db.QueryRow("SELECT COALESCE(MAX(workflow_id), 0) + 1 FROM plugins_workflow").Scan(&maxID)
	if err != nil {
		return fmt.Errorf("failed to query max workflow_id: %w", err)
	}

	for _, p := range plugins {
		if p.Kind == "workflow" {
			_, err := db.Exec(`
				INSERT INTO plugins_workflow (plugin_id, enabled, workflow_id)
				VALUES (?, 1, ?)
			`, p.ID, maxID)
			if err != nil {
				return fmt.Errorf("failed to insert workflow %s: %w", p.ID, err)
			}
			maxID++
		}
	}
	return nil
}
