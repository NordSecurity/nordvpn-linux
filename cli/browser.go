package cli

import (
	"fmt"
	"os/exec"
)

// browse opens url in the browser. altURL is displayed as an alternate link in the console window.
func browse(url string) error {
	fmt.Println("Opening the web browser...")
	fmt.Printf("If nothing happens, please visit %s\n", url)
	return exec.Command("xdg-open", url).Run()
}
