// Copyright (2012) Sandia Corporation.
// Under the terms of Contract DE-AC04-94AL85000 with Sandia Corporation,
// the U.S. Government retains certain rights in this software.
//
// David Fritz <djfritz@sandia.gov>

// command line interface for minimega
//
// The command line interface wraps a number of commands listed in the
// cliCommands map. Each entry to the map defines a function that is called
// when the command is invoked on the command line, as well as short and long
// form help. The record parameter instructs the cli to put the command in the
// command history.
//
// The cli uses the readline library for command history and tab completion.
// A separate command history is kept and used for writing the buffer out to
// disk.

package main

import (
	"fmt"
	"goreadline"
	"minicli"
	log "minilog"
	"sort"
	"strings"
	"sync"
)

const (
	COMMAND_TIMEOUT = 10
)

var (
	commandBuf []string // command history for the write command

	// incoming commands for the cli to parse. these can come from the cli
	// proper (readline), or from a network source, etc. the cli will parse
	// them all as if they were typed locally.
	commandChanLocal   chan cliCommand
	commandChanSocket  chan cliCommand
	commandChanMeshage chan cliCommand

	ackChanLocal   chan cliResponse // acknowledgements from the cli, one per incoming command
	ackChanSocket  chan cliResponse
	ackChanMeshage chan cliResponse

	// Prevents multiple commands from running at the same time
	cmdLock sync.Mutex
)

type cliCommand struct {
	Command string
	Args    []string
	ackChan chan cliResponse
	TID     int32
}

type cliResponse struct {
	Response string
	Error    string // because you can't gob/json encode an error type
	More     bool   // more is set if the called command will be sending multiple responses
	TID      int32
}

func init() {
	commandChanLocal = make(chan cliCommand)
	commandChanSocket = make(chan cliCommand)
	commandChanMeshage = make(chan cliCommand)
	ackChanLocal = make(chan cliResponse)
	ackChanSocket = make(chan cliResponse)
	ackChanMeshage = make(chan cliResponse)
}

// Wrapper for minicli.ProcessString. Ensures that the command execution lock
// is acquired before running the command.
func runCommand(cmd *minicli.Command, record bool) chan minicli.Responses {
	cmdLock.Lock()
	defer cmdLock.Unlock()

	return minicli.ProcessCommand(cmd, record)
}

// local command line interface, wrapping readline
func cliLocal() {
	for {
		prompt := "minimega$ "
		line, err := goreadline.Rlwrap(prompt)
		if err != nil {
			break // EOF
		}
		command := string(line)
		log.Debug("got from stdin:", command)

		cmd, err := minicli.CompileCommand(command)
		if err != nil {
			fmt.Printf("err: %v, closest match: TODO\n", err)
			continue
		}

		// No command was returned, must have been a blank line or a comment
		// line. Either way, don't try to run a nil command.
		if cmd == nil {
			continue
		}

		// HAX: Don't record the read command
		record := !strings.HasPrefix(command, "read")

		for resp := range runCommand(cmd, record) {
			log.Debug("cli resp: %v", resp)
			// print the responses
			fmt.Println(resp)
		}
	}
}

// sort and walk the api, emitting markdown for each entry
func docGen() {
	var keys []string
	// TODO
	//for k, _ := range cliCommands {
	//	keys = append(keys, k)
	//}
	sort.Strings(keys)

	fmt.Println("# minimega API")

	for _, k := range keys {
		fmt.Printf("<h2 id=%v>%v</h2>\n", k, k)
		//fmt.Println(cliCommands[k].Helplong)
	}
}

var poeticDeath = `
Willst du immer weiterschweifen?
Sieh, das Gute liegt so nah.
Lerne nur das Glück ergreifen,
denn das Glück ist immer da.
`
