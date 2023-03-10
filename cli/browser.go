package cli

import (
	"fmt"
	"os/exec"
)

func browse(url string) error {
	fmt.Println("Opening the web browser...")
	fmt.Printf("If nothing happens, please visit %s\n", url)
	return exec.Command("xdg-open", url).Run()
}
