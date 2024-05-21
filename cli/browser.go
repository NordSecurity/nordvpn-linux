package cli

import (
	"fmt"
	"os/exec"
)

func browse(url string, altURL string) error {
	fmt.Println("Opening the web browser...")
	fmt.Printf("If nothing happens, please visit %s\n", altURL)
	return exec.Command("xdg-open", url).Run()
}
