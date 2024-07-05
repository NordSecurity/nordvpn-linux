package cli

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/fatih/color"
	"golang.org/x/term"
)

func isStdoutATerminal() bool {
	return term.IsTerminal(int(os.Stdout.Fd()))
}

func serverNameLen(server *pb.ServerGroup) int {
	return len(server.Name)
}

func formatServerName(server *pb.ServerGroup) string {
	if server.VirtualLocation && isStdoutATerminal() {
		return color.HiBlueString(server.Name)
	}
	return server.Name
}

func footerForServerGroupsList(servers []*pb.ServerGroup) string {
	if !isStdoutATerminal() {
		return ""
	}

	for _, server := range servers {
		if server.VirtualLocation {
			return color.HiBlueString(MsgFooterVirtualLocationNote)
		}
	}
	return ""
}

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
	oldState, err := term.MakeRaw(0)
	if err != nil {
		return username, password, err
	}
	defer func() {
		if err := term.Restore(0, oldState); err != nil {
			log.Println(internal.DeferPrefix, err)
		}
	}()

	screen := struct {
		io.Reader
		io.Writer
	}{os.Stdin, os.Stdout}
	term := term.NewTerminal(screen, "")
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
	if !term.IsTerminal(0) || !term.IsTerminal(1) {
		return planID, fmt.Errorf("Stdin/Stdout should be terminal")
	}

	oldState, err := term.MakeRaw(0)
	if err != nil {
		return planID, err
	}
	defer func() {
		if err := term.Restore(0, oldState); err != nil {
			log.Println(internal.DeferPrefix, err)
		}
	}()

	screen := struct {
		io.Reader
		io.Writer
	}{os.Stdin, os.Stdout}
	term := term.NewTerminal(screen, "")

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
func columns[T any](
	data []T,
	length func(T) int,
	display func(T) string,
	footer string,
) (string, error) {
	width, _, err := cliDimensions()
	if err != nil {
		return "", err
	}

	return formatTable(data, length, display, width, footer)
}

func formatTable[T any](
	data []T,
	length func(T) int,
	display func(T) string,
	width int,
	footer string,
) (string, error) {
	if width <= 0 {
		return "", fmt.Errorf("invalid width size")
	}

	if len(data) == 0 {
		return "", nil
	}

	// Calculate the maximum width of an item
	maxItemWidth := 0
	for _, item := range data {
		itemSize := length(item)
		if itemSize > maxItemWidth {
			maxItemWidth = itemSize
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
	for i, item := range data {
		itemWidth := columnWidth
		isLastElementOnLine := ((i+1)%columns == 0)
		isLast := (i == len(data)-1)
		if isLastElementOnLine || isLast {
			// don't add extra space for the last item on the row
			itemWidth = 0
		}

		value := display(item)
		diff := len(value) - length(item)
		builder.WriteString(fmt.Sprintf("%-*s", itemWidth+diff, value))

		if isLastElementOnLine && !isLast {
			// add new line for rows, except last one
			builder.WriteString("\n")
		}
	}

	if len(footer) > 0 {
		builder.WriteString("\n\n" + footer)
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
