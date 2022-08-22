package main

import (
	"archive/zip"
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/bodgit/sevenzip"
	marchive "github.com/mholt/archiver/v3"
	cp "github.com/otiai10/copy"
)

const SKIP_EXTRACT = false
const USE_PRESET_CONFIGS = true
const SKIP_COPY = false

var SUPPORTED_ARCHIVE_FORMATS = [3]string{".zip", ".7z", ".rar"}

func checkError(e error) {
	if e != nil {
		fmt.Println(e.Error())
	}
}

func PathIncludesArchive(path string) bool {
	for _, val := range SUPPORTED_ARCHIVE_FORMATS {
		if strings.Contains(path, val) {
			return true
		}
	}
	return false
}

func cloneZipItem(f *zip.File, dest string) {
	// Create full directory path
	path := filepath.Join(dest, f.Name)
	err := os.MkdirAll(filepath.Dir(path), os.ModeDir|os.ModePerm)
	checkError(err)

	// Clone if item is a file
	rc, err := f.Open()
	checkError(err)
	if !f.FileInfo().IsDir() {
		// Use os.Create() since Zip don't store file permissions.
		fileCopy, err := os.Create(path)
		checkError(err)
		_, err = io.Copy(fileCopy, rc)
		fileCopy.Close()
		checkError(err)
	}
	rc.Close()
}

func clone7ZipItem(f *sevenzip.File, dest string) {
	path := filepath.Join(dest, f.Name)
	err := os.MkdirAll(filepath.Dir(path), os.ModeDir|os.ModePerm)
	checkError(err)

	// Clone if item is a file
	rc, err := f.Open()
	checkError(err)
	if !f.FileInfo().IsDir() {
		// Use os.Create() since Zip don't store file permissions.
		fileCopy, err := os.Create(path)
		checkError(err)
		_, err = io.Copy(fileCopy, rc)
		fileCopy.Close()
		checkError(err)
	}
	rc.Close()
}

func ExtractArchiveRarArchive(zip_path, dest string) {
	err := marchive.Unarchive(zip_path, dest)
	if err != nil && !strings.Contains(err.Error(), "already exists") {
		log.Fatal(err)
	}
}

func Extract(zip_path, dest string) {
	r, err := zip.OpenReader(zip_path)
	checkError(err)
	defer r.Close()
	for _, f := range r.File {
		cloneZipItem(f, dest)
	}
}

func Extract7Zip(zip_path, dest string) {
	r, err := sevenzip.OpenReader(zip_path)
	checkError(err)
	defer r.Close()
	for _, f := range r.File {
		clone7ZipItem(f, dest)
	}
}

const DATA = "DATA"                               // data to openmw.cfg
const DATA_DIRECT = "DATA_DIRECT"                 // not prefixing
const CONTENT = "CONTENT"                         // adds content bundles, ie esps
const SETTINGS = "SETTINGS"                       // add lines to settings.cfg
const RESOURCES = "RESOURCES"                     // add a resources line to openmw.cfg
const DEELETE_LIST = "DELETE_LIST"                // list of file paths in data
const DELETE_LIST_BY_FILE = "DELETE_LIST_BY_FILE" // list of directories
const INSTALL_TO_OMW = "INSTALL_TO_OMW_FOLDER"    // add content to openMW folder where openmw.cfg is
const DELTA = "DELTA_PLUGIN"                      // run delta plugin tmerge

func getFileName(filename string) string {
	for _, val := range SUPPORTED_ARCHIVE_FORMATS {
		if strings.Contains(filename, val) {
			return strings.Replace(filename, val, "", 1)
		}
	}
	return filename
}

func readLines(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, scanner.Err()
}

// writeLines writes the lines to the given file.
func writeLines(lines []string, path string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	w := bufio.NewWriter(file)
	for _, line := range lines {
		fmt.Fprintln(w, line)
	}
	return w.Flush()
}

func remove(slice []string, s int) []string {
	return append(slice[:s], slice[s+1:]...)
}

func filterArrayByString(lines []string, filters []string) []string {
	var finalLines []string
	for _, line := range lines {
		var matched = false
		for _, filter := range filters {
			if line == filter {
				matched = true
			}
		}
		if !matched {
			finalLines = append(finalLines, line)
		}
	}
	return finalLines
}

var CONFIG_FILTERS = []string{"data=", "resources="}
var CONTENT_FILTERS = []string{"content=Morrowind.esm",
	"content=Tribunal.esm",
	"content=Bloodmoon.esm"}

func LoadBlankConfigArrays(filepath string) ([]string, []string) {
	configLines, err := readLines(filepath)
	var contentLines []string
	var newConfigLines []string
	checkError(err)

	for _, line := range configLines {
		if (!strings.Contains(line, "content=")) || strings.Contains(line, "Morrowind/Data Files") {
			newConfigLines = append(newConfigLines, line)
		}
		if strings.Contains(line, "content=") {
			contentLines = append(contentLines, line)
		}
	}

	newConfigLines = RemoveDuplicateStr(newConfigLines)
	contentLines = RemoveDuplicateStr(contentLines)

	return newConfigLines, contentLines
}

func LoadSettingsArray(filepath string) []string {
	settingsLines, err := readLines(fmt.Sprint(filepath))
	var newLines []string
	checkError(err)

	for _, line := range settingsLines {
		newLines = append(newLines, line)
	}

	return newLines
}

func UnpackSettingsStep(step UnpackStep, record ManifestRecord, filepath string, configLoc string) {
	var newLines []string

	// write lines into config
	settingsPath := fmt.Sprint(configLoc, "/", "settings.cfg")
	settingsLines := LoadSettingsArray(settingsPath)
	for _, setting := range settingsLines {
		newLines = append(newLines, setting)
	}

	for _, setting := range step.Data {
		newLines = append(newLines, setting)
	}

	newLines = RemoveDuplicateStr(newLines)

	writeLines(newLines, settingsPath)
}

func UnpackResourcesStep(step UnpackStep, record ManifestRecord, filepath string, configLines []string, contentLines []string, configLoc string) {
	var newLines []string

	// write lines into config
	configPath := fmt.Sprint(configLoc, "/", "openmw.cfg")
	for _, line := range configLines {
		newLines = append(newLines, line)
	}

	for _, path := range step.Data {
		newLine := fmt.Sprint("resources=\"", filepath, "/", getFileName(record.FileName), "/", path, "\"")
		newLines = append(newLines, newLine)
	}

	// reapply content lines
	for _, content := range contentLines {
		newLines = append(newLines, content)
	}

	writeLines(newLines, configPath)
}

// this method is using the wrong record for the file name, should not using siteFileName
func UnpackInstallToOMWFolder(step UnpackStep, record ManifestRecord, filepath string, configLoc string) {
	folderName := getFileName(record.FileName)

	for i := 0; i <= len(step.Data)-1; i++ {
		arg1 := fmt.Sprint(step.Data[i])
		arg2 := fmt.Sprint(step.Data[i+1])

		copyPath := fmt.Sprint(filepath, "/", folderName, "/", arg1)
		err := cp.Copy(copyPath, fmt.Sprint(configLoc, "/", arg2))
		checkError(err)

		// args are in pairs
		i++
	}
}

func AddBaseContent(filepath string) {
	configLines, err := readLines(filepath)
	checkError(err)

	var missingPlugins []string
	for _, configLine := range configLines {
		for _, content := range CONTENT_FILTERS {
			if configLine == content {
				missingPlugins = append(missingPlugins, content)
			}
		}
	}

	// insert at the top of the content list, since this effects load order
	if len(missingPlugins) > 0 {
		index := len(configLines) - 1
		for configIndex, val := range configLines {
			if strings.Contains(val, "content=") {
				index = configIndex
				break
			}
		}
		configLines = insertIntoSlice(configLines, index, missingPlugins)
	}

	writeLines(configLines, filepath)
}

func AddDeltaContent(filepath string, deltaFolder string) {
	configLines, err := readLines(filepath)
	checkError(err)

	// openmw.exe --config merges base cfg with the input cfg, meaning -
	// we need to remove morrowind, tribunal, and bloodmoon plugins or it wont run
	configLines = filterArrayByString(configLines, CONTENT_FILTERS)
	deltaData := fmt.Sprint("data=\"", deltaFolder, "\"")
	configLines = append(configLines, deltaData, "content=DeltaPluginMerged.omwaddon")
	writeLines(configLines, filepath)
}

func UnpackContentStep(filepath string, data []string) {
	configLines, err := readLines(filepath)
	checkError(err)

	for _, val := range data {
		content := fmt.Sprint("content=", val)
		configLines = append(configLines, content)
	}

	writeLines(configLines, filepath)
}

func UnpackDataStep(step UnpackStep, record ManifestRecord, filepath string, configLines []string, contentLines []string, configLoc string) {
	var newLines []string

	// write lines into config
	configPath := fmt.Sprint(configLoc, "/", "openmw.cfg")
	for _, line := range configLines {
		newLines = append(newLines, line)
	}

	for _, path := range step.Data {
		newLine := fmt.Sprint("data=\"", filepath, "/", getFileName(record.FileName), "/", path, "\"")
		newLines = append(newLines, newLine)
	}

	// reapply content lines
	for _, content := range contentLines {
		newLines = append(newLines, content)
	}

	writeLines(newLines, configPath)
}

func UnpackDataDirectStep(step UnpackStep, record ManifestRecord, filepath string, configLines []string, contentLines []string, configLoc string) {
	var newLines []string

	// write lines into config
	configPath := fmt.Sprint(configLoc, "/", "openmw.cfg")
	for _, line := range configLines {
		newLines = append(newLines, line)
	}

	for _, fallback := range step.Data {
		newLines = append(newLines, fallback)
	}

	// reapply content lines
	for _, content := range contentLines {
		newLines = append(newLines, content)
	}

	writeLines(newLines, configPath)
}

func UnpackDeleteStep(step UnpackStep, record ManifestRecord, filepath string) {
	folderName := getFileName(record.FileName)
	for _, path := range step.Data {
		e := os.Remove(fmt.Sprint(filepath, "/", folderName, "/", path))
		checkError(e)
	}
}

func UnpackDeleteByFileStep(step UnpackStep, record ManifestRecord, filepath string) {
	folderName := getFileName(record.FileName)
	for _, dataPath := range step.Data {
		pathList, err := readLines(fmt.Sprint("./", dataPath))
		checkError(err)
		for _, path := range pathList {
			e := os.Remove(fmt.Sprint(filepath, "/", folderName, "/", path))
			checkError(e)
		}

	}
}

func findFileManifestRecord(records []ManifestRecord, modId int32, index int16) ManifestRecord {
	var relevantRecords []ManifestRecord
	for _, step := range records {
		if step.ModId == modId {
			relevantRecords = append(relevantRecords, step)
		}
	}

	return relevantRecords[index]
}

func GetCurrentDirPath() string {
	currentDirectory, err := os.Getwd()
	if err != nil {
		log.Println(err)
	}
	return strings.Replace(currentDirectory, "\\", "/", -1)
}

func UnpackMods(config ModListConfig, manifest ManifestListConfig, downloadFolder string, prefs PreferencesConfig) {
	modInstallFolder := fmt.Sprint(prefs.Modinstall, "/", manifest.ListName)
	if prefs.SharedInstallFolder {
		modInstallFolder = fmt.Sprint(prefs.Modinstall, "/shared")
	}

	if !SKIP_EXTRACT {
		for _, val := range manifest.Records {
			zipPath := fmt.Sprint(downloadFolder, "/", val.FileName)
			downloadName := getFileName(val.FileName)
			location := fmt.Sprint(modInstallFolder, "/", downloadName)
			extracted, err := Exists(location)
			checkError(err)

			if strings.Contains(val.FileName, ".zip") && !extracted {
				Extract(zipPath, location)
			} else if strings.Contains(val.FileName, ".7z") && !extracted {
				Extract7Zip(zipPath, location)
			} else if strings.Contains(val.FileName, ".rar") && !extracted {
				ExtractArchiveRarArchive(zipPath, location)
			}
		}
	}

	var configPath = fmt.Sprint(prefs.Settings, "/", "openmw.cfg")
	currentDirectory, err := os.Getwd()
	if err != nil {
		log.Println(err)
	}
	currentDirectory = strings.Replace(currentDirectory, "\\", "/", -1)
	currentPresetPath := fmt.Sprint(currentDirectory, "/presets/", config.Name, "/")
	if USE_PRESET_CONFIGS {
		configPath = fmt.Sprint(prefs.Settings, "/", "openmw.cfg")
		newConfigPath := fmt.Sprint(currentPresetPath, "openmw.cfg")
		if exists, _ := Exists(newConfigPath); exists {
			e := os.Remove(newConfigPath)
			if e != nil {
				log.Fatal(e)
			}
		}
		Copy(prefs.Settings, currentPresetPath, "openmw.cfg")
		configPath = fmt.Sprint(currentPresetPath, "/openmw.cfg")
	}

	configLines, contentLines := LoadBlankConfigArrays(configPath)

	for _, step := range config.UnpackSteps {
		var record = ManifestRecord{}
		if step.ModId > 0 {
			record = findFileManifestRecord(manifest.Records, step.ModId, step.FileIndex)
		}
		switch step.Type {
		case DATA:
			UnpackDataStep(step, record, modInstallFolder, configLines, contentLines, currentPresetPath)
			configLines, contentLines = LoadBlankConfigArrays(configPath)
		case DATA_DIRECT:
			UnpackDataDirectStep(step, record, modInstallFolder, configLines, contentLines, currentPresetPath)
			configLines, contentLines = LoadBlankConfigArrays(configPath)
		case DEELETE_LIST:
			UnpackDeleteStep(step, record, modInstallFolder)
		case DELETE_LIST_BY_FILE:
			UnpackDeleteByFileStep(step, record, modInstallFolder)
		case SETTINGS:
			UnpackSettingsStep(step, record, modInstallFolder, currentPresetPath)
		case RESOURCES:
			// UnpackResourcesStep(step, record, modInstallFolder, configLines, contentLines)
			// configLines, contentLines = LoadBlankConfigArrays(configPath)
		case INSTALL_TO_OMW:
			UnpackInstallToOMWFolder(step, record, modInstallFolder, currentPresetPath)
		case DELTA:
			deltaFolderPath := fmt.Sprint(modInstallFolder, "/DeltaPlugin")
			AddBaseContent(configPath)
			CreateDeltaPlugin(prefs.Delta, currentPresetPath, deltaFolderPath)
			AddDeltaContent(configPath, deltaFolderPath)
		case CONTENT:
			UnpackContentStep(configPath, step.Data)
			configLines, contentLines = LoadBlankConfigArrays(configPath)
		default:
			log.Fatal("Unknown instruction type.")
		}
	}
}
