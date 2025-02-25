package arptables

import (
	"errors"
	"fmt"
	"log"
	"os/exec"
	"strings"

	"github.com/NordSecurity/nordvpn-linux/internal"
)

type ARPTablesMode int

const (
	Insert ARPTablesMode = iota
	Delete
)

type ARPTablesChain int

const (
	INPUT ARPTablesChain = iota
	OUTPUT
)

type CommandFunc func(cmd string, args ...string) (string, error)

func RunCommand(cmd string, args ...string) (string, error) {
	output, err := exec.Command(cmd, args...).CombinedOutput()
	return string(output), err
}

func isRulePresent(commandFunc CommandFunc, targetRule string, chain string) (bool, error) {
	currentRules, err := commandFunc("arptables", "-L", chain)
	if err != nil {
		return false, fmt.Errorf("getting current rules: %w", err)
	}

	for _, rule := range strings.Split(currentRules, "\n") {
		if rule == targetRule {
			return true, nil
		}
	}

	return false, nil
}

var ErrARPRuleAlreadyInChain = errors.New("rule already present in chain")
var ErrARPRuleNotPresentInChain = errors.New("rule not presnet in chain")

func cmdARPTables(commandFunc CommandFunc, mode ARPTablesMode, chain ARPTablesChain, nicName string) error {
	nicChainArg := "INPUT"
	nicDirectionArg := "-i"
	if chain == OUTPUT {
		nicChainArg = "OUTPUT"
		nicDirectionArg = "-o"
	}

	rule := fmt.Sprintf("-j DROP %s %s", nicDirectionArg, nicName)
	rulePresent, err := isRulePresent(commandFunc, rule, nicChainArg)
	if err != nil {
		return fmt.Errorf("checking if rule is present: %w", err)
	}

	modeArg := ""
	switch mode {
	case Insert:
		if rulePresent {
			return ErrARPRuleAlreadyInChain
		}
		modeArg = "-I"
	case Delete:
		if !rulePresent {
			return ErrARPRuleNotPresentInChain
		}
		modeArg = "-D"
	}

	args := append([]string{modeArg, nicChainArg}, strings.Split(rule, " ")...)
	if _, err := commandFunc("arptables", args...); err != nil {
		return fmt.Errorf("executing arptables command: %w", err)
	}

	return nil
}

func UnblockARP(commandFunc CommandFunc, nicName string) error {
	if err := cmdARPTables(commandFunc, Delete, INPUT, nicName); err != nil {
		if !errors.Is(err, ErrARPRuleNotPresentInChain) {
			return fmt.Errorf("blocking arptables INPUT chain: %w", err)
		}
		log.Println(internal.ErrorPrefix,
			"Attempted to delete rule blocking incoming ARPs for:", nicName, "but it was not present in the chain")
	}
	if err := cmdARPTables(commandFunc, Delete, OUTPUT, nicName); err != nil {
		if !errors.Is(err, ErrARPRuleNotPresentInChain) {
			return fmt.Errorf("blocking arptables OUTPUT chain: %w", err)
		}
		log.Println(internal.ErrorPrefix,
			"Attempted to delete rule blocking outgoing ARPs for:", nicName, "but it was not present in the chain")
	}

	return nil
}

func BlockARP(commandFunc CommandFunc, nicName string) error {
	if err := cmdARPTables(commandFunc, Insert, INPUT, nicName); err != nil {
		if !errors.Is(err, ErrARPRuleAlreadyInChain) {
			return fmt.Errorf("blocking arptables INPUT chain: %w", err)
		}
		log.Println(internal.ErrorPrefix,
			"Attempted to insert rule blocking incoming ARPs for:", nicName, "but it was already present in chain")
	}
	if err := cmdARPTables(commandFunc, Insert, OUTPUT, nicName); err != nil {
		if !errors.Is(err, ErrARPRuleAlreadyInChain) {
			return fmt.Errorf("blocking arptables OUTPUT chain: %w", err)
		}
		log.Println(internal.ErrorPrefix,
			"Attempted to insert rule blocking outgoing ARPs for:", nicName, "but it was already present in chain")
	}

	return nil
}
