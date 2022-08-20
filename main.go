package main

import (
	"os"
)

func main() {
	// decide to run UI or CMD process
	if len(os.Args) == 0 {
		// run visual UI
	} else {
		// run command line
		RunTerminal()
	}

}

/*
* ./aradir.exe --preset=<presetname>
    * runs solely from preferences.yaml and does both download and unpack
    * runs checks on the paths in there to make sure they at least exist and fails early if not
		* mutually ex
* "--download"
* "--unpack"

* You can skip download, but if you download and not unpack, Aradir wont run because it would have a stale mod folder state
* this means you can go through the download process by itself and not go througn unpack until later if you dont want to
* this is good for testing, as well, but wouldn't be used by users often
* "./aradir.exe --download"
* "./aradir.exe --unpack"


* paths
* "./aradir.exe
	--downloads=D:/downloads/openmw
	--game-data=D:/games/morrowind
	--settings=D:/docs/openmw
	--openmw=D:/games/openMW
	--delta=C:/download/DeltaPlugin

if your preset requires delta, it will also include a delta path
--delta=C:/download/DeltaPlugin

*/
