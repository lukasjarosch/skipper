package main

import (
	"errors"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/fatih/color"
	"github.com/manifoldco/promptui"

	"github.com/lukasjarosch/skipper/v1/expression"
)

var (
	errorColor   = color.New(color.Bold, color.FgRed)
	printfError  = errorColor.PrintfFunc()
	printlnError = errorColor.PrintlnFunc()

	previousInput string
)

// run is the function which is being called on every iteration of the app-loop.
func run() error {
	templates := &promptui.PromptTemplates{
		Prompt:  "{{ . }}",
		Valid:   ">>> ",
		Invalid: "xxx ",
		Success: "",
	}

	p := promptui.Prompt{
		Label:       "",
		Templates:   templates,
		HideEntered: true,
	}

	input, err := p.Run()
	if err != nil {
		return err
	}
	if len(input) == 0 {
		return nil
	}

	fmt.Printf(">>> %s\n", input)

	previousInput = input
	input = strings.TrimSpace(input)
	inputWords := strings.Split(input, " ")

	switch inputWords[0] {
	case "expr":
		expr(strings.Join(inputWords[1:], " "))
	}

	return nil
}

func expr(input string) {
	input = strings.TrimSpace(input)

	// expression.Parse may panic, so be prepared
	defer func() {
		e := recover()
		if e != nil {
			printlnError(e)
		}
	}()

	if !strings.HasPrefix(input, "${") {
		input = "${ " + input
	}
	if !strings.HasSuffix(input, "}") {
		input = input + " }"
	}

	expr := expression.Parse(input)
	if len(expr) == 0 {
		printlnError("no expression present ")
		return
	}

	fmt.Println(color.GreenString("✔ expression:"), expr[0].Text())
}

func main() {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	done := make(chan bool, 1)

	go func() {
		// start app-loop
		for {
			err := run()
			if err != nil {
				if !errors.Is(err, promptui.ErrInterrupt) {
					fmt.Println("ERR:", err)
				}
				done <- true
				return
			}

		}
	}()

	// Wait for a signal or error
	select {
	case <-sig:
		done <- true
		os.Exit(0)
	case <-done:
		return
	}
}
