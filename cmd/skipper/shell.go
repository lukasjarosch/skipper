package main

import (
	"github.com/c-bata/go-prompt"
	"github.com/ryboe/q"
	"github.com/urfave/cli/v2"
)

func completer(d prompt.Document) []prompt.Suggest {
	s := []prompt.Suggest{
		{Text: "project", Description: "Define the path to the skipper project to load"},
		{Text: "expr", Description: "Execute an expression"},
		{Text: "help", Description: "Show help for the skipper shell"},
		{Text: "exit", Description: "Quit the shell"},
	}
	return prompt.FilterHasPrefix(s, d.GetWordBeforeCursor(), true)
}

func shellAction(ctx *cli.Context) error {
	input := prompt.Input(">>> ", completer)
	q.Q(input)
	return nil
}
