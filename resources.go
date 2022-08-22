package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

type DownloadStep struct {
	Type         string `yaml:"type"`
	ModId        int32  `yaml:"modId"`
	SiteFileName string `yaml:"siteFileName"`
}

type UnpackStep struct {
	ModId     int32    `yaml:"modId"`
	FileIndex int16    `yaml:"fileIndex"`
	Type      string   `yaml:"type"`
	Data      []string `yaml:"data"`
}

type ModListConfig struct {
	Name          string         `yaml:"name"`
	LastModified  int32          `yaml:"lastModified"`
	ListUrl       string         `yaml:"listUrl"`
	DownloadSteps []DownloadStep `yaml:"downloadSteps"`
	UnpackSteps   []UnpackStep   `yaml:"unpackSteps"`
}

type ManifestRecord struct {
	FileName        string `yaml:"fileName"`
	ModId           int32  `yaml:"modId"`
	FileDisplayName string `yaml:"fileDisplayName"`
}

type ManifestListConfig struct {
	ListName string           `yaml:"listName"`
	Created  int64            `yaml:"created"`
	Records  []ManifestRecord `yaml:"records"`
}

type PreferencesConfig struct {
	Preset              string `yaml:"preset"`              // preset name
	Downloads           string `yaml:"downloads"`           // downloads path
	Modinstall          string `yaml:"modinstall"`          // mod extract/install path
	Gamedata            string `yaml:"gamedata"`            // morrowind installation path
	Settings            string `yaml:"settings"`            // openmw settings path
	Openmw              string `yaml:"openmw"`              // openmw install path
	Delta               string `yaml:"delta"`               // delta plugin executable path
	Nodownload          bool   `yaml:"nodownload"`          // skip download step completely
	SharedInstallFolder bool   `yaml:"sharedInstallFolder"` // use a combined install folder for all presets or one for each preset
}

// exists returns whether the given file or directory exists
func Exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func RemoveDuplicateStr(strSlice []string) []string {
	allKeys := make(map[string]bool)
	list := []string{}
	for _, item := range strSlice {
		if _, value := allKeys[item]; !value {
			allKeys[item] = true
			list = append(list, item)
		}
	}
	return list
}

func GetDownloadedMods(listName string, downloadPath string) ([]string, ManifestListConfig) {
	downloadedMods := []string{}
	now := time.Now()
	manifestTemplate := ManifestListConfig{
		ListName: listName,
		Created:  now.Unix(),
		Records:  []ManifestRecord{},
	}
	manifestName := fmt.Sprint(listName, "-manifest.yaml")
	listManifestExists, err := Exists(fmt.Sprint(GetCurrentDirPath(), "/manifests/", manifestName))
	checkError(err)

	if listManifestExists {
		manifest := ReadManifest(manifestName)
		manifestTemplate = manifest
		for _, val := range manifest.Records {
			fileExists, err := Exists(fmt.Sprint(downloadPath, "/", val.FileName))
			checkError(err)
			if fileExists {
				downloadedMods = append(downloadedMods, val.FileDisplayName)
			}
		}
	}
	return downloadedMods, manifestTemplate
}

func WriteManifest(manifestData *ManifestListConfig, configFileName string) {
	data, err := yaml.Marshal(&manifestData)
	if err != nil {
		log.Fatal(err)
	}

	val, _ := Exists("./manifests")
	if !val {
		err := os.MkdirAll("./manifests/", os.ModeDir|os.ModePerm)
		checkError(err)
	}

	manName := fmt.Sprint("manifests/", configFileName, "-manifest", ".yaml")
	writeErr := ioutil.WriteFile(manName, data, 0644)
	if writeErr != nil {
		log.Fatal(writeErr)
	}
}

func hasExts(path string, exts []string) bool {
	pathExt := strings.ToLower(filepath.Ext(path))
	for _, ext := range exts {
		if pathExt == strings.ToLower(ext) {
			return true
		}
	}
	return false
}

func CreateFileList() []string {
	var presets []string

	dir := "./"
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
	sourceExts := []string{"manifest"}
	for _, fi := range files {
		if hasExts(fi.Name(), sourceExts) && fi.Name() != "preferences" {
			presets = append(presets, fmt.Sprint(fi.Name(), ".yaml"))
		}
	}
	return presets
}

func ReadPreset(fileName string) ModListConfig {
	preset := ModListConfig{}
	presetName := strings.Replace(fileName, ".yaml", "", 1)
	file, err := ioutil.ReadFile(fmt.Sprint("./", "presets/", presetName, "/", fileName))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}

	parseErr := yaml.Unmarshal([]byte(file), &preset)
	if parseErr != nil {
		fmt.Println(parseErr.Error())
		log.Fatalf("error: %v", err)
	}
	return preset
}

func ReadPrefs(fileName string) PreferencesConfig {
	preset := PreferencesConfig{}
	file, err := ioutil.ReadFile(fmt.Sprint("./", fileName))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}

	parseErr := yaml.Unmarshal([]byte(file), &preset)
	if parseErr != nil {
		fmt.Println(parseErr.Error())
		log.Fatalf("error: %v", err)
	}
	return preset
}

func ReadManifest(fileName string) ManifestListConfig {
	var manifest = ManifestListConfig{}
	file, err := ioutil.ReadFile(fmt.Sprint("./", "manifests/", fileName))

	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}

	parseErr := yaml.Unmarshal([]byte(file), &manifest)
	if parseErr != nil {
		log.Fatalf("error: %v", err)
	}
	return manifest
}

func LoadLists() []ModListConfig {
	lists := []ModListConfig{}
	presetFiles := CreateFileList()
	for _, name := range presetFiles {
		content := ReadPreset(name)
		lists = append(lists, content)
	}
	return lists
}
