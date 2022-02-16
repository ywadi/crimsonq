package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"reflect"
	"strings"
	"ywadi/crimsonq/Servers"

	"github.com/chzyer/readline"
	"github.com/cosiner/argv"
	"github.com/go-redis/redis/v8"
)

var ctx = context.Background()

var completer *readline.PrefixCompleter

func filterInput(r rune) (rune, bool) {
	switch r {
	// block CtrlZ feature
	case readline.CharCtrlZ:
		return r, false
	}
	return r, true
}

func main() {
	hostPtr := flag.String("host", "localhost:9001", "Host to connect to with port, ex: localhost:9001")
	passwordPtr := flag.String("password", "", "Password to connect to CrimsonQ! server.")
	flag.Parse()
	rdb := redis.NewClient(&redis.Options{
		Addr:     *hostPtr,
		Password: *passwordPtr, // no password set
	})
	fmt.Println("\033[31m")
	fmt.Println(`
╔═╗┬─┐┬┌┬┐┌─┐┌─┐┌┐┌╔═╗ 
║  ├┬┘││││└─┐│ ││││║═╬╗
╚═╝┴└─┴┴ ┴└─┘└─┘┘└┘╚═╝╚`)
	fmt.Println("Cli Version 1.0.0 \033[0m")

	Servers.InitCommands()
	cmds := Servers.Commands
	completer = readline.NewPrefixCompleter()
	acItems := []readline.PrefixCompleterInterface{}
	for c, _ := range cmds {
		acItems = append(acItems, readline.PcItem(c))
	}
	completer.SetChildren(acItems)
	l, err := readline.NewEx(&readline.Config{
		Prompt:          "\033[32mcrimsonQ»\033[0m ",
		HistoryFile:     "/tmp/readline.tmp",
		AutoComplete:    completer,
		InterruptPrompt: "^C",
		EOFPrompt:       "exit",

		HistorySearchFold:   true,
		FuncFilterInputRune: filterInput,
	})
	if err != nil {
		panic(err)
	}
	defer l.Close()

	setPasswordCfg := l.GenPasswordConfig()
	setPasswordCfg.SetListener(func(line []rune, pos int, key rune) (newLine []rune, newPos int, ok bool) {
		l.SetPrompt(fmt.Sprintf("Enter password(%v): ", len(line)))
		l.Refresh()
		return nil, 0, false
	})

	log.SetOutput(l.Stderr())
	for {
		line, err := l.Readline()
		if err == readline.ErrInterrupt {
			if len(line) == 0 {
				break
			} else {
				continue
			}
		} else if err == io.EOF {
			break
		} else if len(line) == 0 {
			continue
		}

		line = strings.TrimSpace(line)
		cmdArgs := splitArgs(line)

		if strings.ToLower(cmdArgs[0].(string)) == "subscribe" {
			pubsub := rdb.Subscribe(ctx, cmdArgs[1].(string))
			//TODO: If error on command call, how to manage it
			fmt.Println("Listening on " + cmdArgs[1].(string))
			for {
				msg, err := pubsub.ReceiveMessage(ctx)
				if err != nil {
					fmt.Println(err)
				}

				fmt.Println(msg.Channel, msg.Payload)
			}
		}

		val, err := rdb.Do(ctx, cmdArgs...).Result()
		if err != nil {
			if err == redis.Nil {
				fmt.Println("error", err)
			}
			fmt.Println("error", err)
		} else {
			v := reflect.ValueOf(val)
			if v.Len() == 0 {
				fmt.Println("null")
				continue
			}
			if reflect.TypeOf(val).Kind() == reflect.Slice {
				for i := 0; i < v.Len(); i++ {
					e := v.Index(i)
					fmt.Printf("%v) %v\n", i, e)
				}
			} else {
				fmt.Println(">", val.(string))
			}
		}

		if strings.ToLower(cmdArgs[0].(string)) == "quit" {
			os.Exit(0)
		}

	}
}

func splitArgs(cmdline string) []interface{} {
	args, err := argv.Argv(cmdline, func(backquoted string) (string, error) {
		return backquoted, nil
	}, nil)
	if err != nil {
		log.Fatal(err)
	}
	var argsInterface []interface{}
	for _, v := range args[0] {
		argsInterface = append(argsInterface, v)
	}
	return argsInterface
}
