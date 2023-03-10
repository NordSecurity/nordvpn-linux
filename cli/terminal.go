package cli

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/NordSecurity/nordvpn-linux/internal"

	"golang.org/x/crypto/ssh/terminal"
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

	if !terminal.IsTerminal(0) || !terminal.IsTerminal(1) {
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

// ReadCredentialsFromStdIn reads username and password from standard input
func ReadCredentialsFromStdIn() (string, string, error) {
	var (
		username string
		password string
	)
	reader := bufio.NewReader(os.Stdin)
	username, err := reader.ReadString('\n')
	if err != nil {
		if err == io.EOF {
			return username, password, fmt.Errorf("%s: %s", internal.ErrStdin, "username")
		}
		return username, password, internal.ErrUnhandled
	}
	password, err = reader.ReadString('\n')
	username = strings.Trim(username, " ")
	username = strings.Trim(username, "\n")
	password = strings.Trim(password, " ")
	password = strings.Trim(password, "\n")
	if err != nil {
		if err == io.EOF {
			return username, password, fmt.Errorf("%s: %s", internal.ErrStdin, "password")
		}
		return username, password, internal.ErrUnhandled
	}

	if err = checkUsernamePasswordIsEmpty(username, password); err != nil {
		return username, password, fmt.Errorf("%s: %v", internal.ErrStdin, err)
	}

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

// isStdInAvailable checks whether standard input is not empty
func isStdInAvailable() bool {
	info, _ := os.Stdin.Stat()
	return info.Size() > 0
}
