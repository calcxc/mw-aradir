package main

import (
	"flag"
	"fmt"
	"log"
	"os/exec"
	"strings"

	reflections "github.com/oleiade/reflections"
)

type FlagDef[T string | bool] struct {
	name        string
	defaultVal  T
	description string
}

type FlagValue[T string | bool] struct {
	name  string
	value T
}

var stringDefs = []FlagDef[string]{
	{name: "preset", defaultVal: "", description: "preset ID"},
	{name: "downloads", defaultVal: "", description: "downloads folder path"},
	{name: "modinstall", defaultVal: "", description: "mods install folder path"},
	{name: "gamedata", defaultVal: "", description: "game data folder path"},
	{name: "settings", defaultVal: "", description: "settings folder path"},
	{name: "openmw", defaultVal: "", description: "openmw install folder path"},
	{name: "delta", defaultVal: "", description: "delta plugin path"},
}

var boolDefs = []FlagDef[bool]{
	{name: "nodownload", defaultVal: false, description: "skip downloading mods"},
}

func RunTerminal() {
	prefs := ReadPrefs("preferences.yaml")

	// fill default values stringDefs, boolDefs
	for i, val := range stringDefs {
		value, err := reflections.GetField(prefs, strings.Title(val.name))
		checkError(err)
		stringDefs[i] = FlagDef[string]{defaultVal: value.(string), name: val.name, description: val.description}
	}

	stringMap := make(map[string]*string)
	boolMap := make(map[string]*bool)
	// Register flags for processing
	for _, strDef := range stringDefs {
		stringMap[strDef.name] = flag.String(strDef.name, strDef.defaultVal, strDef.description)
	}

	for _, boolDef := range boolDefs {
		boolMap[boolDef.name] = flag.Bool(boolDef.name, boolDef.defaultVal, boolDef.description)
	}

	// Execute flags
	flag.Parse()

	// Use values taken from command line args
	for _, strDef := range stringDefs {
		err := reflections.SetField(&prefs, strings.Title(strDef.name), *stringMap[strDef.name])
		if err != nil {
			fmt.Println(err.Error())
		}
	}

	for _, boolDef := range boolDefs {
		err := reflections.SetField(&prefs, strings.Title(boolDef.name), *boolMap[boolDef.name])
		if err != nil {
			fmt.Println(err.Error())
		}
	}

	if prefs.Preset == "" {
		log.Fatal("Preset field is unset")
	}

	configFileName := prefs.Preset
	downloadFolder := prefs.Downloads
	var config = ModListConfig{}
	var manifest = ManifestListConfig{}
	if !prefs.Nodownload {
		config, manifest = DownloadMods(configFileName, downloadFolder)
	} else {
		configName := fmt.Sprint(prefs.Preset, ".yaml")
		manifestName := fmt.Sprint(prefs.Preset, "-manifest.yaml")
		config = ReadPreset(configName)
		manifest = ReadManifest(manifestName)
	}

	UnpackMods(config, manifest, downloadFolder, prefs)
	openMWExe := fmt.Sprint(prefs.Openmw, "/openmw.exe")
	configPath := fmt.Sprint(GetCurrentDirPath(), "/presets/", configFileName, "/")
	RunOpenMW(openMWExe, configPath)
}

func RunOpenMW(path string, configPath string) {
	config := fmt.Sprint("--config=", configPath)
	replace := fmt.Sprint("--replace=config")
	cmd := exec.Command(path, config, replace)
	err := cmd.Start()
	if err != nil {
		fmt.Println(err.Error())
	}
}
