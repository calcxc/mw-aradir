package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/rod/lib/proto"
)

const NEXUS_MODS_URL = "https://www.nexusmods.com/morrowind/mods/"

// navigate to settings/Downloads in Chrome
// disable ask "Ask where to save each file before downloading"
// Close settings and reopen to make sure it saved

func createRodHandler() *rod.Browser {
	u := launcher.NewUserMode().MustLaunch()

	return rod.New().ControlURL(u).MustConnect().NoDefaultDevice()
}

func createPageHandler() *rod.Page {
	rod := createRodHandler()
	return rod.MustPage(NEXUS_MODS_URL)
}

func createJSSelector(elementType string, replaceString string) string {
	selector := `() => {
		const text = "TEXTCONTENT";
		const elements = Array.from(document.querySelectorAll("ELEMENTTYPE"));
	
	return elements.find(el => {
		return el.textContent.includes(text);
	});
	}`
	// ELEMENTTYPE = dt, div, span, etc
	// TEXTCONTENT = what to look for inside

	selector = strings.Replace(selector, "TEXTCONTENT", replaceString, 1)
	selector = strings.Replace(selector, "ELEMENTTYPE", elementType, 1)

	return selector
}

func TryNexusDownload(page *rod.Page, siteFileName string) string {
	tabs := page.MustElement(".modtabs")
	tabs.MustElementR("span", `FILES`).MustClick()
	// page.MustElementR("span", "/Manual/i").MustParent() //.MustClick()

	selector := createJSSelector("dt", siteFileName)
	section := page.MustElement("#mod_files").MustElementByJS(selector).MustNext()
	link := section.MustElementR("span", "/Manual download/i").MustParent()
	link.MustClick()

	time.Sleep(1 * time.Second)
	if page.MustHas(".popup-mod-requirements") {
		popup, _ := page.Element(".popup-mod-requirements")
		popup.MustElementR("a.btn", "/download/i").MustClick()
	}

	time.Sleep(1 * time.Second)
	fileName := ""
	if page.MustHas("div.page-layout") && page.MustHas("div.donation-wrapper") {
		tabSection, _ := page.Element("div.tabcontent")
		fileNameContainer, _ := tabSection.Element("div.header")
		readName, _ := fileNameContainer.Text()

		// Might be fragile, since it catches other content
		splitName := strings.Split(readName, "\n")
		fileName = splitName[0]
	}

	wait := page.WaitEvent(proto.PageDownloadWillBegin{})
	// proto.BrowserDownloadProgress{}
	// watch for proto.BrowserDownloadProgress by event.GUID and event.State
	wait()

	return fileName
}

func nextPage(page *rod.Page, url string) {
	wait := page.MustWaitNavigation()
	page.MustNavigate(url)
	wait()
}

func sliceContains(slice []string, val string) bool {
	for _, value := range slice {
		if value == val {
			return true
		}
	}
	return false
}

func filterByModId(records []ManifestRecord, modId int32) []ManifestRecord {
	var matchedRecords = []ManifestRecord{}
	for _, val := range records {
		if modId == val.ModId {
			matchedRecords = append(matchedRecords, val)
		}
	}
	return matchedRecords
}

func getDownloadCount(steps []DownloadStep, modId int32) int {
	var matchedRecords = []DownloadStep{}
	for _, val := range steps {
		if modId == val.ModId {
			matchedRecords = append(matchedRecords, val)
		}
	}
	return len(matchedRecords)
}

func DownloadMods(listName string, downloadFolder string) (ModListConfig, ManifestListConfig) {
	fullFileName := fmt.Sprint(listName, ".yaml")
	preset := ReadPreset(fullFileName)
	var page = &rod.Page{}

	// Check for existing manifest
	// if manifest exists, return a list of siteFileNames that match the config
	// enables download skipping to save time and storage
	downloadedMods, manifest := GetDownloadedMods(listName, downloadFolder)
	if len(downloadedMods) == len(preset.DownloadSteps) {
		return preset, manifest
	}
	page = createPageHandler()
	for _, step := range preset.DownloadSteps {
		if sliceContains(downloadedMods, step.SiteFileName) {
			continue
		}
		fileName := ""
		/* 		stepCount := getDownloadCount(preset.DownloadSteps, step.ModId)
		   		matchedRecords := filterByModId(manifest.Records, step.ModId)
		   		if record := matchedRecords[stepCount - 1]; record != nil {

		   		} */
		// check manifest for modId
		// check
		// set filename
		nextPage(page, fmt.Sprint(NEXUS_MODS_URL, step.ModId, "/files"))
		if step.Type == "nexus" {
			fileName = TryNexusDownload(page, step.SiteFileName)
		}

		fmt.Println(fileName)
		manifest.Records = append(manifest.Records, ManifestRecord{
			FileName:        fileName,
			ModId:           step.ModId,
			FileDisplayName: step.SiteFileName,
		})

		time.Sleep(500 * time.Millisecond) // wait half second before transitioning
	}

	WriteManifest(&manifest, listName)
	return preset, manifest
}
