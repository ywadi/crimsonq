package RedconQ

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/tidwall/redcon"
)

// var mu sync.RWMutex
// var ps redcon.PubSub

func Ping(con redcon.Conn, args ...[][]byte) {
	con.WriteString("Pong! " + string(args[0][0]))
}
func Quit(con redcon.Conn, args ...[][]byte) {
	con.WriteString("Bye!")
	con.Close()
}
func Auth(con redcon.Conn, args ...[][]byte) {
	if string(args[0][0]) == "pass" {
		cntxt := con.Context().(ConnContext)
		cntxt.Auth = true
		con.SetContext(cntxt)
		con.WriteString("Yo!")
	}
}
func Command(con redcon.Conn, args ...[][]byte) {
	con.WriteArray(len(Commands))
	for k, v := range Commands {
		con.WriteBulk([]byte(k + " [" + strings.Join(v.ArgsCmd, "] [") + "]"))
	}
}

func Subscribe(con redcon.Conn) {}
func Exists(con redcon.Conn, args ...[][]byte) {
	con.WriteString(strconv.FormatBool(crimsonQ.ConsumerExists(string(args[0][0]))))
}
func Select(con redcon.Conn, args ...[][]byte) {
	ctx := con.Context().(ConnContext)
	consumerId := string(args[0][0])
	if crimsonQ.ConsumerExists(consumerId) {
		ctx.SelectDB = consumerId
		con.SetContext(ctx)
		con.WriteString("Selected [" + consumerId + "]")
	} else {
		//TODO: Create QDB
		con.WriteString("No such consumer id, created and selecting " + consumerId)
	}

}
func Destroy(con redcon.Conn, args ...[][]byte) {}
func List(con redcon.Conn, args ...[][]byte) {
	clist := crimsonQ.ListConsumers()
	con.WriteArray(len(clist))
	for _, s := range clist {
		con.WriteBulkString(s)
	}
}

func Msg_Keys(con redcon.Conn)                            {}
func Msg_Push_Topic(con redcon.Conn, args ...[][]byte)    {}
func Msg_Push_Consumer(con redcon.Conn, args ...[][]byte) {}
func Msg_Pull(con redcon.Conn, args ...[][]byte)          {}
func Msg_Del(con redcon.Conn, args ...[][]byte)           {}
func Msg_Fail(con redcon.Conn, args ...[][]byte)          {}
func Msg_Complete(con redcon.Conn, args ...[][]byte)      {}
func Msg_Retry(con redcon.Conn, args ...[][]byte)         {}
func Msg_Retry_All(con redcon.Conn, args ...[][]byte)     {}
func Flush_Complete(con redcon.Conn, args ...[][]byte)    {}
func Flush_Failed(con redcon.Conn, args ...[][]byte)      {}

var Commands map[string]RedConCmds

type RedConCmds struct {
	Function       interface{}
	ArgsCmd        []string
	RequiresSelect bool
}

func initCommands() {
	Commands = map[string]RedConCmds{
		"ping":              {Function: Ping, ArgsCmd: []string{"messageString"}, RequiresSelect: false},
		"quit":              {Function: Quit, ArgsCmd: []string{}, RequiresSelect: false},
		"auth":              {Function: Auth, ArgsCmd: []string{"password"}, RequiresSelect: false},
		"command":           {Function: Command, ArgsCmd: []string{}, RequiresSelect: false},
		"subscribe":         {Function: Subscribe, ArgsCmd: []string{}, RequiresSelect: false},
		"exists":            {Function: Exists, ArgsCmd: []string{"consumerId"}, RequiresSelect: false},
		"select":            {Function: Select, ArgsCmd: []string{"consumerId"}, RequiresSelect: false},
		"destroy":           {Function: Destroy, ArgsCmd: []string{}, RequiresSelect: true},
		"list":              {Function: List, ArgsCmd: []string{}, RequiresSelect: false},
		"msg_keys":          {Function: Msg_Keys, ArgsCmd: []string{}, RequiresSelect: true},
		"msg_push_Topic":    {Function: Msg_Push_Topic, ArgsCmd: []string{"topicString", "messageString"}, RequiresSelect: true},
		"msg_push_Consumer": {Function: Msg_Push_Consumer, ArgsCmd: []string{"messageString"}, RequiresSelect: true},
		"msg_pull":          {Function: Msg_Pull, ArgsCmd: []string{}, RequiresSelect: true},
		"msg_del":           {Function: Msg_Del, ArgsCmd: []string{"messageId"}, RequiresSelect: true},
		"msg_fail":          {Function: Msg_Fail, ArgsCmd: []string{"messageId"}, RequiresSelect: true},
		"msg_complete":      {Function: Msg_Complete, ArgsCmd: []string{"messageId"}, RequiresSelect: true},
		"msg_retry":         {Function: Msg_Retry, ArgsCmd: []string{"messageId"}, RequiresSelect: true},
		"msg_retry_all":     {Function: Msg_Retry_All, ArgsCmd: []string{}, RequiresSelect: true},
		"flush_complete":    {Function: Flush_Complete, ArgsCmd: []string{}, RequiresSelect: true},
		"flush_failed":      {Function: Flush_Failed, ArgsCmd: []string{}, RequiresSelect: true},
	}
}
func execCommand(conn redcon.Conn, cmd redcon.Command) {
	cCmd := strings.ToLower(string(cmd.Args[0]))
	if conn.Context().(ConnContext).Auth || cCmd == "auth" {
		if val, ok := Commands[cCmd]; ok {
			if len(val.ArgsCmd) == len(cmd.Args)-1 {
				val.Function.(func(con redcon.Conn, values ...[][]byte))(conn, cmd.Args[1:])
			} else {
				conn.WriteError("Incorrect number of arguments for " + cCmd + ", expected " + string(len(cmd.Args)-1) + "(" + strings.Join(val.ArgsCmd, ",") + ") but got " + fmt.Sprint(len(cmd.Args)) + " Args")
			}
		}
	} else {
		conn.WriteError("Auth Error: You Shall not pass!")
		conn.Close()
	}

}
