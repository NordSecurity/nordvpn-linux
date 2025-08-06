//go:build mage

package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
	"strings"

	"github.com/NordSecurity/nordvpn-linux/nstrings"
)

// Define a new type to hide it's content from prints
type secret struct {
	value string
}

func newSecret(val string) secret {
	return secret{value: val}
}
func (s secret) GoString() string {
	return s.String()
}

func (s secret) String() string {
	const display = 10
	str := string(s.value)

	// If the string is too short to mask
	if len(str) <= display*2 {
		return "secret:****"
	}

	return "secret:" + str[:display] + "..." + str[len(str)-display:]
}

// VagrantEnv stores all the env variables needed to setup the VM using vagrant and to execute the tests inside VM
type VagrantEnv struct {
	Cwd             string // project folder
	VagrantFileDir  string // folder containing Vagrantfile
	BoxName         string // box name used, e.g. generic/ubuntu2204
	ShouldDestroyVM bool   // when true, the VM created will be destroyed after tests are executed
	RemoteCwd       string // where was the project synced in VM, by default /vagrant
	TestCredentials secret // testing credentials, env[NA_TESTS_CREDENTIALS]
}

// runVagrantCmd - run vagrant commands
func runVagrantCmd(vagrantEnv VagrantEnv, args ...string) error {
	cmd := exec.Command("vagrant", args...)
	cmd.Dir = vagrantEnv.VagrantFileDir
	cmd.Env = append(os.Environ(),
		fmt.Sprintf("WORKDIR=%s", vagrantEnv.Cwd),
	)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return err
	}

	go streamLines(stdout, os.Stdout)
	go streamLines(stderr, os.Stderr)

	// Start the command
	if err := cmd.Start(); err != nil {
		return err
	}

	return cmd.Wait()
}

func streamLines(r io.Reader, out io.Writer) {
	sc := bufio.NewScanner(r)
	for sc.Scan() {
		fmt.Fprintln(out, sc.Text())
	}
}

// runCommandInVM - runs a command using vagrant ssh inside the VM
func runCommandInVM(vagrantEnv VagrantEnv, args []string) error {
	env := []string{
		fmt.Sprintf("WORKDIR=%s", vagrantEnv.RemoteCwd),
		fmt.Sprintf("NA_TESTS_CREDENTIALS='%s'", vagrantEnv.TestCredentials),
	}

	// compose the remote command to be on single string:
	// cd /vagrant + env + <cmd + params>
	fullRemoteCmd := fmt.Sprintf("cd %s && %s %s",
		vagrantEnv.RemoteCwd,
		strings.Join(env, " "),
		strings.Join(args, " "),
	)
	fmt.Println("executing remote command:", fullRemoteCmd)
	return runVagrantCmd(vagrantEnv, "ssh", vagrantEnv.BoxName, "-c", fullRemoteCmd)
}

func vagrantUp(vagrantEnv VagrantEnv) error {
	return runVagrantCmd(vagrantEnv, "up", vagrantEnv.BoxName)
}

func vagrantStop(vagrantEnv VagrantEnv) {
	runVagrantCmd(vagrantEnv, "halt", "-f", vagrantEnv.BoxName)
	if vagrantEnv.ShouldDestroyVM {
		runVagrantCmd(vagrantEnv, "destroy", "-f", vagrantEnv.BoxName)
	}
}

// buildVagrantEnvEnv will construct a VagrantEnv storing all the env variables needed
// Used env variables:
// * SNAP_TEST_BOX_NAME specifies what box to use, for example generic/ubuntu2204
// * SNAP_TEST_DESTROY_VM_ON_EXIT - optional, when set to 1|true it will delete the VM after running the tests
// * NA_TESTS_CREDENTIALS - test credentials
func buildVagrantEnvEnv() (VagrantEnv, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return VagrantEnv{}, err
	}

	env, err := getEnv()
	if err != nil {
		return VagrantEnv{}, err
	}

	boxName := env["SNAP_TEST_BOX_NAME"]
	if boxName == "" {
		fmt.Println("Missing env var SNAP_TEST_BOX_NAME")
		return VagrantEnv{}, fmt.Errorf("Missing SNAP_TEST_BOX_NAME")
	}
	// store folder that contains Vagrantfile used to create and run the box
	vagrantFileDir := path.Join(cwd, "ci", "snap", "vagrant")

	fmt.Println("WORKDIR=", cwd)
	fmt.Println("Vagrantfile folder", vagrantFileDir)

	shouldDestroyVM := false
	if val, ok := env["SNAP_TEST_DESTROY_VM_ON_EXIT"]; ok {
		shouldDestroyVM, err = nstrings.BoolFromString(val)
		if err != nil {
			shouldDestroyVM = false
		}
	}

	vagrantEnv := VagrantEnv{
		Cwd:             cwd,
		VagrantFileDir:  vagrantFileDir,
		BoxName:         boxName,
		ShouldDestroyVM: shouldDestroyVM,
		RemoteCwd:       "/vagrant",
		TestCredentials: newSecret(env["NA_TESTS_CREDENTIALS"]),
	}
	return vagrantEnv, nil
}

// RunInVM - executes the commands using vagrant ssh in a VM.
// It will also setup, configure and start(+stop) the VM using vagrant
func RunInVM(args ...string) error {
	// select all the needed env variables to create and run the tests
	vagrantEnv, err := buildVagrantEnvEnv()
	if err != nil {
		return err
	}
	fmt.Println("Vagrant env", vagrantEnv)

	// setup and start the VM
	if err := vagrantUp(vagrantEnv); err != nil {
		return err
	}
	defer vagrantStop(vagrantEnv)

	// run the tests inside the VM
	arguments := make([]string, len(args))
	for i := range args {
		arguments[i] = fmt.Sprintf("%q", args[i]) // adds double quotes to the arguments
	}
	return runCommandInVM(vagrantEnv, arguments)
}
