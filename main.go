package main

import (
	"caido_plugin_installer/cmd"
	"caido_plugin_installer/utils"
	"database/sql"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
)

const origin = "store"

var homeDir, _ = os.UserHomeDir()
var pluginDirectory = filepath.Join(homeDir, ".local", "share", "caido", "plugins")
var manifestPath = "/dist/plugin_package/manifest.json"
var pluginBackendPath = "/dist/plugin_package/backend"
var pluginFrontendPath = "/dist/plugin_package/frontend"
var pluginsDB = filepath.Join(homeDir, ".local", "share", "caido", "plugins.db")

func main() {
	pluginArchive := flag.String("package", "", "Plugin archive")

	flag.Parse()

	pluginPath, err := cmd.ExtractPKG(pluginArchive)
	if err != nil {
		panic(err)
	}

	err = cmd.BuildPKG(pluginPath)
	if err != nil {
		panic(err)
	}

	manifest, err := utils.ParseManifest(pluginPath + manifestPath)
	manifest.UUID.Backend = utils.GenerateUUID()
	manifest.UUID.Frontend = utils.GenerateUUID()
	manifest.UUID.PluginID = utils.GenerateUUID()

	fmt.Printf("Backend UUID: %s\nFrontend UUID: %s\nPlugin UUID: %s\n", manifest.UUID.Backend, manifest.UUID.Frontend, manifest.UUID.PluginID)
	pluginDirectory = pluginDirectory + "/"
	err = os.MkdirAll(pluginDirectory+manifest.UUID.Backend, 0775)
	if err != nil {
		fmt.Println("Error creating directory:", err)
		return
	}
	fmt.Printf("Directory created successfully: %s\n", pluginDirectory+manifest.UUID.Backend)

	err = os.MkdirAll(pluginDirectory+manifest.UUID.Frontend, 0775)
	if err != nil {
		fmt.Println("Error creating directory:", err)
		return
	}
	fmt.Printf("Directory created successfully: %s\n", pluginDirectory+manifest.UUID.Frontend)

	cmd.CopyDir(pluginPath+pluginBackendPath, pluginDirectory+manifest.UUID.Backend)
	cmd.CopyDir(pluginPath+pluginFrontendPath, pluginDirectory+manifest.UUID.Frontend)

	db, err := sql.Open("sqlite3", pluginsDB)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	if err := utils.InsertPluginPackage(db, manifest.UUID.PluginID, manifest, origin); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	if err := utils.InsertPluginEntries(db, manifest, manifest.Plugins); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	for _, p := range manifest.Plugins {
		switch p.Kind {
		case "frontend":
			if err := utils.InsertFrontendUserPrefs(db, manifest.UUID.Frontend); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
		case "backend":
			if err := utils.InsertBackendEnabled(db, manifest.UUID.Backend); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
		}
	}

	if err := utils.InsertWorkflows(db, manifest.Plugins); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("\n✅ Plugin installed successfully!")
	fmt.Printf("   Package ID: %s\n", manifest.UUID.PluginID)
	fmt.Printf("   Origin: %s\n", origin)
	fmt.Println("   Plugin parts:")
	for _, p := range manifest.Plugins {
		fmt.Printf("     - %-10s %-20s (%s)\n", p.Kind, p.Name, manifest.UUID.PluginID[:8])
		// fmt.Printf("       Directory: %s\n", p.Dir)
		// entries, _ := os.ReadDir(p.dir)
		// for _, e := range entries {
		// 	info, err := e.Info()
		// 	if err == nil {
		// 		fmt.Printf("         %s (%d bytes)\n", e.Name(), info.Size())
		// 	}
	}
}
