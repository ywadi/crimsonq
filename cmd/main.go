package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"reflect"
	"strings"
	"ywadi/crimsonq/Servers"

	"github.com/chzyer/readline"
	"github.com/cosiner/argv"
	"github.com/go-redis/redis/v8"
)

var ctx = context.Background()

func usage(w io.Writer) {
	io.WriteString(w, "commands:\n")
	io.WriteString(w, completer.Tree("    "))
}

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
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:9001",
		Password: "crimsonQ!", // no password set
	})
	fmt.Println(`
	╔═╗┬─┐┬┌┬┐┌─┐┌─┐┌┐┌╔═╗ 
	║  ├┬┘││││└─┐│ ││││║═╬╗
	╚═╝┴└─┴┴ ┴└─┘└─┘┘└┘╚═╝╚`)
	fmt.Println("CrimsonQ v1.0.0, gitsha: Ab#12455")

	Servers.InitCommands()
	cmds := Servers.Commands
	completer = readline.NewPrefixCompleter()
	acItems := []readline.PrefixCompleterInterface{}
	for c, _ := range cmds {
		acItems = append(acItems, readline.PcItem(c))
	}
	completer.SetChildren(acItems)
	l, err := readline.NewEx(&readline.Config{
		Prompt:          "\033[31mcrimsonQ»\033[0m ",
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

		if len(line) != 0 {

		}

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
