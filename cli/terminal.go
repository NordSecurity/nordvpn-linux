package cli

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/NordSecurity/nordvpn-linux/internal"
	"golang.org/x/crypto/ssh/terminal"
	"golang.org/x/term"
)

func checkUsernamePasswordIsEmpty(username, password string) error {
	if username == "" {
		return fmt.Errorf("Email / Username must not be empty")
	}

	if password == "" {
		return fmt.Errorf("Password must not be empty")
	}

	return nil
}

// ReadCredentialsFromTerminal reads username and password from terminal
// this overrides current terminal and restores it upon completion
func ReadCredentialsFromTerminal() (string, string, error) {
	var (
		username string
		password string
	)

	if !term.IsTerminal(0) || !term.IsTerminal(1) {
		return username, password, fmt.Errorf("Stdin/Stdout should be terminal")
	}
	oldState, err := terminal.MakeRaw(0)
	if err != nil {
		return username, password, err
	}
	defer func() {
		if err := terminal.Restore(0, oldState); err != nil {
			log.Println(internal.DeferPrefix, err)
		}
	}()

	screen := struct {
		io.Reader
		io.Writer
	}{os.Stdin, os.Stdout}
	term := terminal.NewTerminal(screen, "")
	term.SetPrompt("Email: ")
	username, err = term.ReadLine()
	if err != nil {
		return username, password, err
	}

	term.AutoCompleteCallback = func(line string, pos int, key rune) (string, int, bool) {
		// Mask the password output to the console with '*'
		line += "*"
		// Handle backspaces.
		password = string([]rune(password)[:pos])
		// Advance the cursor.
		pos++
		// Add the actual key presses to the password.
		password += string(key)
		return line, pos, true
	}
	term.SetPrompt("Password: ")
	_, err = term.ReadLine()
	if err != nil {
		return username, password, err
	}

	err = checkUsernamePasswordIsEmpty(username, password)
	return username, password, err
}

func ReadPlanFromTerminal() (int, error) {
	var planID int
	if !terminal.IsTerminal(0) || !terminal.IsTerminal(1) {
		return planID, fmt.Errorf("Stdin/Stdout should be terminal")
	}

	oldState, err := terminal.MakeRaw(0)
	if err != nil {
		return planID, err
	}
	defer func() {
		if err := terminal.Restore(0, oldState); err != nil {
			log.Println(internal.DeferPrefix, err)
		}
	}()

	screen := struct {
		io.Reader
		io.Writer
	}{os.Stdin, os.Stdout}
	term := terminal.NewTerminal(screen, "")

	for {
		term.SetPrompt("Plan number: ")
		plan, err := term.ReadLine()
		if err != nil {
			return planID, err
		}

		planID, err = strconv.Atoi(plan)
		if err != nil {
			switch err.(type) {
			case *strconv.NumError:
				continue
			}
			return planID, err
		}
		break
	}
	return planID, nil
}

// formats a list of strings to a tidy column representation
func columns(data []string) (string, error) {
	width, _, err := cliDimensions()
	if err != nil {
		// workaround for tests because terminal size cannot be calculated
		if flag.Lookup("test.v") != nil {
			return strings.Join(data, " "), err
		}
		return "", err
	}

	return formatTable(data, width)
}

func formatTable(data []string, width int) (string, error) {
	if width <= 0 {
		return "", fmt.Errorf("invalid width size")
	}

	if len(data) == 0 {
		return "", nil
	}

	// Calculate the maximum width of an item
	maxItemWidth := 0
	for _, item := range data {
		if len(item) > maxItemWidth {
			maxItemWidth = len(item)
		}
	}

	// Calculate the number of columns that can fit in the terminal width
	columnWidth := min(width, maxItemWidth+4)
	columns := width / columnWidth

	if columns < 1 {
		// if the width is very small then display everything in one column
		columns = 1
	}

	var builder strings.Builder
	for i, value := range data {
		itemWidth := columnWidth
		isLastElementOnLine := ((i+1)%columns == 0)
		isLast := (i == len(data)-1)
		if isLastElementOnLine || isLast {
			// don't add extra space for the last item on the row
			itemWidth = 0
		}

		builder.WriteString(fmt.Sprintf("%-*s", itemWidth, value))

		if isLastElementOnLine && !isLast {
			// add new line for rows, except last one
			builder.WriteString("\n")
		}
	}

	return builder.String(), nil
}

// Gets the size of CLI window
func cliDimensions() (int, int, error) {
	width, height, err := term.GetSize(int(os.Stdout.Fd()))
	if err == nil {
		return width, height, nil
	}

	log.Println(internal.InfoPrefix, "failed to get windows size, try stty")

	// if getting the window size from stdout fails, usually when piped, try to execute stty
	cmd := exec.Command(internal.SttyExec, "size")
	cmd.Stdin = os.Stdin

	out, errStty := cmd.CombinedOutput()
	if errStty != nil {
		return 0, 0, errors.Join(err, errStty)
	}

	n, errScanf := fmt.Sscanf(string(out), "%d %d", &height, &width)
	if n != 2 || errScanf != nil {
		return 0, 0, errors.Join(err, errScanf)
	}

	return width, height, nil
}
