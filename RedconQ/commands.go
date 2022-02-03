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

func Subscribe(con redcon.Conn) { //TODO
	//consumerId := string(args[0][0])
}

func Exists(con redcon.Conn, args ...[][]byte) {
	con.WriteString(strconv.FormatBool(crimsonQ.ConsumerExists(string(args[0][0]))))
}
func Select(con redcon.Conn, args ...[][]byte) {
	ctx := con.Context().(ConnContext)
	consumerId := string(args[0][0])
	topicFilters := string(args[0][1])
	if crimsonQ.ConsumerExists(consumerId) {
		ctx.SelectDB = consumerId
		con.SetContext(ctx)
		con.WriteString("Selected [" + consumerId + "]")
	} else {
		crimsonQ.CreateQDB(consumerId, "/home/ywadi/_crimson/_dbs", topicFilters)
		con.WriteString("No such consumer id, created and selecting " + consumerId)
	}

}
func Destroy(con redcon.Conn, args ...[][]byte) {
	//TODO
	//consumerId := string(args[0][0])
}

func List(con redcon.Conn, args ...[][]byte) {
	clist := crimsonQ.ListConsumers()
	con.WriteArray(len(clist))
	for _, s := range clist {
		con.WriteBulkString(s)
	}
}

func Msg_Keys(con redcon.Conn, args ...[][]byte) {
	consumerId := string(args[0][0])
	list := crimsonQ.ListAllKeys(consumerId)
	con.WriteArray(len(list))
	for _, s := range list {
		con.WriteBulkString(s)
	}

}

func Msg_Push_Topic(con redcon.Conn, args ...[][]byte) {
	topic := string(args[0][0])
	message := string(args[0][1])
	crimsonQ.PushTopic(topic, message)
	con.WriteString("Ok")
}

func Msg_Push_Consumer(con redcon.Conn, args ...[][]byte) {
	consumerId := string(args[0][0])
	topic := string(args[0][0])
	message := string(args[0][1])
	crimsonQ.PushConsumer(consumerId, topic, message)
	con.WriteString("Ok")
}

func Msg_Pull(con redcon.Conn, args ...[][]byte) {
	consumerId := string(args[0][0])
	msg := crimsonQ.Pull(consumerId)
	con.WriteString(msg.JsonStringify())
}

func Msg_Del(con redcon.Conn, args ...[][]byte) {
	consumerId := string(args[0][0])
	messageId := string(args[0][1])
	crimsonQ.Del(consumerId, messageId)
	con.WriteString("Ok")
}

func Msg_Fail(con redcon.Conn, args ...[][]byte) {
	consumerId := string(args[0][0])
	messageId := string(args[0][1])
	crimsonQ.MsgFail(consumerId, messageId)
	con.WriteString("Ok")
}

func Msg_Complete(con redcon.Conn, args ...[][]byte) {
	consumerId := string(args[0][0])
	messageId := string(args[0][1])
	crimsonQ.MsgComplete(consumerId, messageId)
	con.WriteString("Ok")
}

func Msg_Retry(con redcon.Conn, args ...[][]byte) {
	consumerId := string(args[0][0])
	messageId := string(args[0][1])
	crimsonQ.MsgRetry(consumerId, messageId)
	con.WriteString("Ok")
}

func Msg_Retry_All(con redcon.Conn, args ...[][]byte) {
	consumerId := string(args[0][0])
	crimsonQ.ReqAllFailed(consumerId)
	con.WriteString("Ok")
}

func Flush_Complete(con redcon.Conn, args ...[][]byte) {
	consumerId := string(args[0][0])
	crimsonQ.ClearComplete(consumerId)
	con.WriteString("Ok")
}

func Flush_Failed(con redcon.Conn, args ...[][]byte) {
	consumerId := string(args[0][0])
	crimsonQ.ClearFailed(consumerId)
	con.WriteString("Ok")
}

var Commands map[string]RedConCmds

type RedConCmds struct {
	Function           interface{}
	ArgsCmd            []string
	RequiresConsumerId bool
}

func initCommands() {
	Commands = map[string]RedConCmds{
		"ping":              {Function: Ping, ArgsCmd: []string{"messageString"}, RequiresConsumerId: false},
		"quit":              {Function: Quit, ArgsCmd: []string{}, RequiresConsumerId: false},
		"auth":              {Function: Auth, ArgsCmd: []string{"password"}, RequiresConsumerId: false},
		"command":           {Function: Command, ArgsCmd: []string{}, RequiresConsumerId: false},
		"subscribe":         {Function: Subscribe, ArgsCmd: []string{"consumerId"}, RequiresConsumerId: true},
		"exists":            {Function: Exists, ArgsCmd: []string{"consumerId"}, RequiresConsumerId: false},
		"select":            {Function: Select, ArgsCmd: []string{"consumerId", "topicFilters"}, RequiresConsumerId: false},
		"destroy":           {Function: Destroy, ArgsCmd: []string{"consumerId"}, RequiresConsumerId: true},
		"list":              {Function: List, ArgsCmd: []string{}, RequiresConsumerId: false},
		"msg_keys":          {Function: Msg_Keys, ArgsCmd: []string{"consumerId"}, RequiresConsumerId: true},
		"msg_push_Topic":    {Function: Msg_Push_Topic, ArgsCmd: []string{"topicString", "messageString"}, RequiresConsumerId: false},
		"msg_push_consumer": {Function: Msg_Push_Consumer, ArgsCmd: []string{"consumerId", "messageString"}, RequiresConsumerId: true},
		"msg_pull":          {Function: Msg_Pull, ArgsCmd: []string{"consumerId"}, RequiresConsumerId: true},
		"msg_del":           {Function: Msg_Del, ArgsCmd: []string{"consumerId", "messageId"}, RequiresConsumerId: true},
		"msg_fail":          {Function: Msg_Fail, ArgsCmd: []string{"consumerId", "messageId"}, RequiresConsumerId: true},
		"msg_complete":      {Function: Msg_Complete, ArgsCmd: []string{"consumerId", "messageId"}, RequiresConsumerId: true},
		"msg_retry":         {Function: Msg_Retry, ArgsCmd: []string{"consumerId", "messageId"}, RequiresConsumerId: true},
		"msg_retry_all":     {Function: Msg_Retry_All, ArgsCmd: []string{"consumerId"}, RequiresConsumerId: true},
		"flush_complete":    {Function: Flush_Complete, ArgsCmd: []string{"consumerId"}, RequiresConsumerId: true},
		"flush_failed":      {Function: Flush_Failed, ArgsCmd: []string{"consumerId"}, RequiresConsumerId: true},
	}
}
func execCommand(conn redcon.Conn, cmd redcon.Command) {
	cCmd := strings.ToLower(string(cmd.Args[0]))
	if conn.Context().(ConnContext).Auth || cCmd == "auth" {
		if val, ok := Commands[cCmd]; ok {
			//Check if the select context is there, it is inject into args as a first after command arg
			if val.RequiresConsumerId {
				if conn.Context().(ConnContext).SelectDB != "" {
					//Add consumerId as first argument
					cmd.Args = append([][]byte{[]byte(conn.Context().(ConnContext).SelectDB)}, cmd.Args...)
				}
			}
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
