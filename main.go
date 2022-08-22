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
