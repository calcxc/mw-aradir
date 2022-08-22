package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"
)

func RunDelta(path string, configPath string, modinstallFolder string) {
	command := fmt.Sprint(path)

	config := fmt.Sprint(configPath, "openmw.cfg")
	merge := fmt.Sprint(modinstallFolder, "/DeltaPluginMerged.omwaddon")
	cmd := exec.Command(command, "--openmw-cfg", config, "merge", merge)
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println(string(output))
		fmt.Println(err.Error())
	}
}

func CreateDeltaPlugin(path string, configPath string, modinstallFolder string) {
	file, err := os.Stat(path)
	checkError(err)
	err = os.MkdirAll(modinstallFolder, 0755)
	checkError(err)
	command := fmt.Sprint(path, "/delta_plugin.exe")
	if file.IsDir() {
		files, err := ioutil.ReadDir(path)
		checkError(err)
		for _, dirEntry := range files {
			if strings.Contains(dirEntry.Name(), "delta_plugin") {
				RunDelta(command, configPath, modinstallFolder)
			}
		}

	} else if PathIncludesArchive(path) {
		// extract into a folder beside archive and run it
	} else if strings.Contains(path, ".exe") {
		RunDelta(command, configPath, modinstallFolder)
	} else {
		log.Fatal("Delta Plugin not found.")
	}
}
