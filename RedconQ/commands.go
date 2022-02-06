package RedconQ

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"ywadi/crimsonq/Defs"

	"github.com/spf13/viper"
	"github.com/tidwall/redcon"
)

// var mu sync.RWMutex
var PS redcon.PubSub

func Ping(con redcon.Conn, args ...[][]byte) error {
	con.WriteString("Pong! " + string(args[0][0]))
	return nil
}

func Quit(con redcon.Conn, args ...[][]byte) error {
	con.WriteString("Bye!")
	con.Close()
	return nil
}

func Auth(con redcon.Conn, args ...[][]byte) error {
	fmt.Println(">>>", con)
	if string(args[0][0]) == viper.GetString("crimson_settings.password") {
		cntxt := con.Context().(ConnContext)
		cntxt.Auth = true
		con.SetContext(cntxt)
		con.WriteString("ok")

	}
	return nil
}

func Command(con redcon.Conn, args ...[][]byte) error {
	con.WriteArray(len(Commands))
	for k, v := range Commands {
		con.WriteBulk([]byte(k + " [" + strings.Join(v.ArgsCmd, "] [") + "]"))
	}
	return nil
}

func Subscribe(con redcon.Conn, args ...[][]byte) error {
	consumerId := string(args[0][0])
	if crimsonQ.ConsumerExists(consumerId) {
		PS.Subscribe(con, consumerId)
	} else {
		con.WriteError(Defs.ERRincorrectConsumerId)
	}
	return nil
}

func Exists(con redcon.Conn, args ...[][]byte) error {
	con.WriteString(strconv.FormatBool(crimsonQ.ConsumerExists(string(args[0][0]))))
	return nil
}

func Consumer_Create(con redcon.Conn, args ...[][]byte) error {
	consumerId := string(args[0][0])
	consumerTopics := string(args[0][1])
	crimsonQ.CreateQDB(consumerId, viper.GetString("crimson_settings.data_rootpath"))
	crimsonQ.SetTopics(consumerId, consumerTopics)
	con.WriteString("Created:" + consumerId)
	return nil
}

func SetConsumerTopics(con redcon.Conn, args ...[][]byte) error {
	consumerId := string(args[0][0])
	topicFilters := string(args[0][1])
	if crimsonQ.ConsumerExists(consumerId) {
		crimsonQ.SetTopics(consumerId, topicFilters)
		con.WriteString("ok")
	} else {
		err := errors.New(Defs.ERRincorrectConsumerId)
		con.WriteError(fmt.Sprint(err))
		return err
	}
	return nil
}

func GetConsumerTopics(con redcon.Conn, args ...[][]byte) error {
	consumerId := string(args[0][0])
	if crimsonQ.ConsumerExists(consumerId) {
		topics := crimsonQ.GetTopics(consumerId)
		con.WriteArray(len(topics))
		for _, t := range topics {
			con.WriteString(t)
		}
	} else {
		err := errors.New(Defs.ERRincorrectConsumerId)
		con.WriteError(fmt.Sprint(err))
		return err
	}
	return nil
}
func Destroy(con redcon.Conn, args ...[][]byte) error {
	consumerId := string(args[0][0])
	if crimsonQ.ConsumerExists(consumerId) {
		crimsonQ.DestroyQDB(consumerId)
		con.WriteString("ok")
	} else {
		con.WriteError(Defs.ERRincorrectConsumerId)
	}
	return nil
}

func List(con redcon.Conn, args ...[][]byte) error {
	clist := crimsonQ.ListConsumers()
	con.WriteArray(len(clist))
	for _, s := range clist {
		con.WriteBulkString(s)
	}
	return nil
}

func Msg_Keys(con redcon.Conn, args ...[][]byte) error {
	consumerId := string(args[0][0])
	list, err := crimsonQ.ListAllKeys(consumerId)
	if err != nil {
		con.WriteError(fmt.Sprint(err))
		return err
	}
	con.WriteArray(len(list))
	for _, s := range list {
		con.WriteBulkString(s)
	}
	return nil
}

func Msg_Counts(con redcon.Conn, args ...[][]byte) error {
	consumerId := string(args[0][0])
	counts, err := crimsonQ.GetKeyCount(consumerId)
	if err != nil {
		con.WriteError(fmt.Sprint(err))
		return err
	}
	con.WriteArray(len(counts))
	for k, v := range counts {
		con.WriteBulkString(k + ":" + strconv.Itoa(int(v)))
	}
	return nil
}

func Msg_Push_Topic(con redcon.Conn, args ...[][]byte) error {
	topic := string(args[0][0])
	message := string(args[0][1])
	consumers := crimsonQ.PushTopic(topic, message)
	for consumer, msgkey := range consumers {
		PS.Publish(consumer, msgkey)
	}
	con.WriteString("Ok")
	return nil
}

func Msg_Push_Consumer(con redcon.Conn, args ...[][]byte) error {
	consumerId := string(args[0][0])
	topic := string(args[0][0])
	message := string(args[0][1])
	if crimsonQ.ConsumerExists(consumerId) {
		msgkey := crimsonQ.PushConsumer(consumerId, "direct:"+topic, message)
		PS.Publish(consumerId, msgkey)
		con.WriteString("Ok")
		return nil
	} else {
		err := errors.New(Defs.ERRincorrectConsumerId)
		con.WriteError(fmt.Sprint(err))
		return err
	}

}

func Msg_Pull(con redcon.Conn, args ...[][]byte) error {
	consumerId := string(args[0][0])
	if crimsonQ.ConsumerExists(consumerId) {
		msg, err := crimsonQ.Pull(consumerId)
		if err != nil {
			con.WriteNull()
			return err
		}
		con.WriteString(msg.JsonStringify())
		return nil
	} else {
		err := errors.New(Defs.ERRincorrectConsumerId)
		con.WriteError(fmt.Sprint(err))
		return err
	}
}

//TODO
func Msg_Del(con redcon.Conn, args ...[][]byte) error {
	consumerId := string(args[0][0])
	status := string(args[0][1])
	messageId := string(args[0][2])

	if crimsonQ.ConsumerExists(consumerId) {
		err := crimsonQ.Del(status, consumerId, messageId)
		if err != nil {
			con.WriteError(fmt.Sprint(err))
			return err
		}
		con.WriteString("Ok")
		return nil
	} else {
		err := errors.New(Defs.ERRincorrectConsumerId)
		con.WriteError(fmt.Sprint(err))
		return err
	}
}

func Msg_Get_Status_Json(con redcon.Conn, args ...[][]byte) error {
	consumerId := string(args[0][0])
	status := strings.ToLower(string(args[0][1]))
	if !(status == Defs.STATUS_ACTIVE || status == Defs.STATUS_COMPLETED || status == Defs.STATUS_DELAYED || status == Defs.STATUS_FAILED || status == Defs.STATUS_PENDING) {
		sError := errors.New(Defs.ERRIncorrectStatus)
		con.WriteError(sError.Error())
		return sError
	}
	if crimsonQ.ConsumerExists(consumerId) {
		json, err := crimsonQ.GetAllByStatusJson(consumerId, status)
		if err != nil {
			con.WriteError(fmt.Sprint(err))
			return err
		}
		con.WriteString(json)
		return nil
	} else {
		err := errors.New(Defs.ERRincorrectConsumerId)
		con.WriteError(fmt.Sprint(err))
		return err
	}
}

func Msg_Fail(con redcon.Conn, args ...[][]byte) error {
	consumerId := string(args[0][0])
	messageId := string(args[0][1])
	fmt.Println(messageId)
	errMsg := string(args[0][1])
	if crimsonQ.ConsumerExists(consumerId) {
		err := crimsonQ.MsgFail(consumerId, messageId, errMsg)
		if err != nil {
			con.WriteError("incorrect message id")
			return err
		}
		con.WriteString("Ok")
		return nil
	} else {
		err := errors.New(Defs.ERRincorrectConsumerId)
		con.WriteError(fmt.Sprint(err))
		return err
	}
}

func Msg_Complete(con redcon.Conn, args ...[][]byte) error {
	consumerId := string(args[0][0])
	messageId := string(args[0][1])
	if crimsonQ.ConsumerExists(consumerId) {
		err := crimsonQ.MsgComplete(consumerId, messageId)
		if err != nil {
			con.WriteError(fmt.Sprint(err))
			return err
		}
		con.WriteString("Ok")
		return nil
	} else {
		err := errors.New(Defs.ERRincorrectConsumerId)
		con.WriteError(fmt.Sprint(err))
		return err
	}
}

func Msg_Retry(con redcon.Conn, args ...[][]byte) error {
	consumerId := string(args[0][0])
	messageId := string(args[0][1])
	if crimsonQ.ConsumerExists(consumerId) {
		err := crimsonQ.MsgRetry(consumerId, messageId)
		if err != nil {
			err := errors.New(Defs.ERRincorrectConsumerId)
			return err
		}
		con.WriteString("Ok")
		return nil
	} else {
		err := errors.New(Defs.ERRincorrectConsumerId)
		con.WriteError(fmt.Sprint(err))
		return err
	}
}

func Msg_Retry_All(con redcon.Conn, args ...[][]byte) error {
	consumerId := string(args[0][0])
	if crimsonQ.ConsumerExists(consumerId) {
		crimsonQ.ReqAllFailed(consumerId)
		con.WriteString("Ok")
		return nil
	} else {
		err := errors.New(Defs.ERRincorrectConsumerId)
		con.WriteError(fmt.Sprint(err))
		return err
	}
}

func Flush_Complete(con redcon.Conn, args ...[][]byte) error {
	consumerId := string(args[0][0])
	if crimsonQ.ConsumerExists(consumerId) {
		crimsonQ.ClearComplete(consumerId)
		con.WriteString("Ok")
		return nil
	} else {
		err := errors.New(Defs.ERRincorrectConsumerId)
		con.WriteError(fmt.Sprint(err))
		return err
	}
}

func Flush_Failed(con redcon.Conn, args ...[][]byte) error {
	consumerId := string(args[0][0])
	if crimsonQ.ConsumerExists(consumerId) {
		crimsonQ.ClearFailed(consumerId)
		con.WriteString("Ok")
		return nil
	} else {
		err := errors.New(Defs.ERRincorrectConsumerId)
		con.WriteError(fmt.Sprint(err))
		return err
	}
}

var Commands map[string]RedConCmds

type RedConCmds struct {
	Function           interface{}
	ArgsCmd            []string
	RequiresConsumerId bool
}

func initCommands() {
	Commands = map[string]RedConCmds{
		"ping":                    {Function: Ping, ArgsCmd: []string{"messageString"}, RequiresConsumerId: false},
		"quit":                    {Function: Quit, ArgsCmd: []string{}, RequiresConsumerId: false},
		"auth":                    {Function: Auth, ArgsCmd: []string{"password"}, RequiresConsumerId: false},
		"command":                 {Function: Command, ArgsCmd: []string{}, RequiresConsumerId: false},
		"subscribe":               {Function: Subscribe, ArgsCmd: []string{"consumerId"}, RequiresConsumerId: true},
		"consumer.exists":         {Function: Exists, ArgsCmd: []string{"consumerId"}, RequiresConsumerId: false},
		"consumer.create":         {Function: Consumer_Create, ArgsCmd: []string{"consumerId", "topics"}, RequiresConsumerId: false},
		"consumer.destroy":        {Function: Destroy, ArgsCmd: []string{"consumerId"}, RequiresConsumerId: true},
		"consumer.list":           {Function: List, ArgsCmd: []string{}, RequiresConsumerId: false},
		"msg.keys":                {Function: Msg_Keys, ArgsCmd: []string{"consumerId"}, RequiresConsumerId: true},
		"msg.counts":              {Function: Msg_Counts, ArgsCmd: []string{"consumerId"}, RequiresConsumerId: true},
		"msg.push.topic":          {Function: Msg_Push_Topic, ArgsCmd: []string{"topicString", "messageString"}, RequiresConsumerId: false},
		"msg.push.consumer":       {Function: Msg_Push_Consumer, ArgsCmd: []string{"consumerId", "messageString"}, RequiresConsumerId: true},
		"msg.pull":                {Function: Msg_Pull, ArgsCmd: []string{"consumerId"}, RequiresConsumerId: true},
		"msg.del":                 {Function: Msg_Del, ArgsCmd: []string{"consumerId", "status", "messageId"}, RequiresConsumerId: true},
		"msg.fail":                {Function: Msg_Fail, ArgsCmd: []string{"consumerId", "messageId", "errMsg"}, RequiresConsumerId: true},
		"msg.complete":            {Function: Msg_Complete, ArgsCmd: []string{"consumerId", "messageId"}, RequiresConsumerId: true},
		"msg.retry":               {Function: Msg_Retry, ArgsCmd: []string{"consumerId", "messageId"}, RequiresConsumerId: true},
		"msg.retryall":            {Function: Msg_Retry_All, ArgsCmd: []string{"consumerId"}, RequiresConsumerId: true},
		"consumer.flush.complete": {Function: Flush_Complete, ArgsCmd: []string{"consumerId"}, RequiresConsumerId: true},
		"consumer.flush.failed":   {Function: Flush_Failed, ArgsCmd: []string{"consumerId"}, RequiresConsumerId: true},
		"consumer.topics.set":     {Function: SetConsumerTopics, ArgsCmd: []string{"consumerId", "topics"}, RequiresConsumerId: true},
		"consumer.topics.get":     {Function: GetConsumerTopics, ArgsCmd: []string{"consumerId"}, RequiresConsumerId: true},
		"msg.list.json":           {Function: Msg_Get_Status_Json, ArgsCmd: []string{"consumerId", "status"}, RequiresConsumerId: true},
	}
}

func execCommand(conn redcon.Conn, cmd redcon.Command) {
	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		cCmd := strings.ToLower(string(cmd.Args[0]))
		for _, x := range cmd.Args {
			fmt.Print(string(x), " ")
		}
		fmt.Println()
		if conn.Context().(ConnContext).Auth || cCmd == "auth" {
			if val, ok := Commands[cCmd]; ok {

				if len(val.ArgsCmd) == len(cmd.Args)-1 {
					val.Function.(func(con redcon.Conn, values ...[][]byte) error)(conn, cmd.Args[1:])
					wg.Done()
				} else {
					conn.WriteError("Incorrect number of arguments for " + cCmd + ", expected " + fmt.Sprint(len(cmd.Args)-1) + "(" + strings.Join(val.ArgsCmd, ",") + ") but got " + fmt.Sprint(len(cmd.Args)) + " Args")
					wg.Done()
				}
				return

			}
			conn.WriteError("incorrect command")
			wg.Done()
		} else {
			conn.WriteError("Auth Error: You Shall not pass!")
			wg.Done()
			conn.Close()
		}
	}()
	wg.Wait()
}
