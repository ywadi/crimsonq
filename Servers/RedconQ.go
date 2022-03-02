package Servers

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
	"ywadi/crimsonq/Defs"
	"ywadi/crimsonq/Structs"
	"ywadi/crimsonq/Utils"

	log "github.com/sirupsen/logrus"

	"github.com/spf13/viper"
	"github.com/tidwall/redcon"
)

type ConnContext struct {
	SelectDB string
	Auth     bool
}

var PS redcon.PubSub
var crimsonQ *Structs.S_GOQ

func StartRedCon(addr string, cq *Structs.S_GOQ) {
	go cq.Init()
	HeartBeat()
	crimsonQ = cq
	log.Info("Started server at %s", addr)
	err := redcon.ListenAndServe(addr,
		execCommand,
		func(conn redcon.Conn) bool {
			ConnContext := ConnContext{Auth: false, SelectDB: ""}
			conn.SetContext(ConnContext)
			remoteIp := strings.Split(conn.RemoteAddr(), ":")[0]
			log.Info("Client connected from ", remoteIp)
			if viper.GetString("RESP.ip_whitelist") == "*" {
				return true
			} else {
				grant := Utils.SliceContains(viper.GetStringSlice("RESP.ip_whitelist"), remoteIp)
				return grant
			}

		},
		func(conn redcon.Conn, err error) {
			// This is called when the connection has been closed
			log.Printf("closed: %s, err: %v", conn.RemoteAddr(), err)

		},
	)
	if err != nil {
		log.Fatal(err)
	}
}

func HeartBeat() {
	log.Info("Heartbeat Started...")
	ticker := time.NewTicker(time.Duration(viper.GetInt64("RESP.heartbeat_seconds")) * time.Second)
	done := make(chan bool)
	go func() {
		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				for _, s := range crimsonQ.QDBPool {
					count, err := crimsonQ.GetKeyCount(s.QdbId)
					if err != nil {
						log.WithFields(log.Fields{"ConsumerId": s.QdbId, "Status": Defs.STATUS_PENDING}).Error("JSON Parse error at heartbeart", err)
					}
					PS.Publish(s.QdbId, "pendingCount:"+fmt.Sprint(count[Defs.STATUS_PENDING]))
				}
			}
		}
	}()
}

func execCommand(conn redcon.Conn, cmd redcon.Command) {
	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		cCmd := strings.ToLower(string(cmd.Args[0]))
		cmdString := ""
		for _, x := range cmd.Args {
			cmdString = cmdString + " " + string(x)
		}

		log.WithFields(log.Fields{"addr": conn.RemoteAddr(), "cmd": cmdString}).Info("command executed")
		if conn.Context().(ConnContext).Auth || cCmd == "auth" {
			if val, ok := Commands[cCmd]; ok {

				if len(val.ArgsCmd) == len(cmd.Args)-1 {
					val.Redcon_Function.(func(con redcon.Conn, values ...[][]byte) error)(conn, cmd.Args[1:])
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

func RC_Ping(con redcon.Conn, args ...[][]byte) error {
	con.WriteString("Pong! " + string(args[0][0]))
	return nil
}

func RC_Consumer_Info(con redcon.Conn, args ...[][]byte) error {
	consumerId := string(args[0][0])
	con.WriteString(crimsonQ.ConsumerInfo(consumerId).JsonStringify())
	return nil
}

func RC_Quit(con redcon.Conn, args ...[][]byte) error {
	con.WriteString("Bye!")
	con.Close()
	return nil
}

func RC_Auth(con redcon.Conn, args ...[][]byte) error {
	respPassVal, _ := os.LookupEnv("CRIMSONQ_RESP_PASS")
	if string(args[0][0]) == respPassVal {
		cntxt := con.Context().(ConnContext)
		cntxt.Auth = true
		con.SetContext(cntxt)
		con.WriteString("ok")

	} else {
		con.WriteError("Inccorect Auth.")
	}
	return nil
}

func RC_Command(con redcon.Conn, args ...[][]byte) error {
	con.WriteArray(len(Commands))
	for k, v := range Commands {
		con.WriteBulk([]byte(k + " [" + strings.Join(v.ArgsCmd, "] [") + "]"))
	}
	return nil
}

func RC_Subscribe(con redcon.Conn, args ...[][]byte) error {
	consumerId := string(args[0][0])
	if crimsonQ.ConsumerExists(consumerId) {
		PS.Subscribe(con, consumerId)
	} else {
		con.WriteError(Defs.ERRincorrectConsumerId)
	}
	return nil
}

func RC_Exists(con redcon.Conn, args ...[][]byte) error {
	con.WriteString(strconv.FormatBool(crimsonQ.ConsumerExists(string(args[0][0]))))
	return nil
}

func RC_Consumer_Create(con redcon.Conn, args ...[][]byte) error {
	consumerId := string(args[0][0])
	if crimsonQ.ConsumerExists(consumerId) {
		con.WriteError(Defs.ERROConsumerAlreadyExists)
		return nil
	}
	consumerTopics := string(args[0][1])
	consumerConcurrent := string(args[0][2])
	crimsonQ.CreateQDB(consumerId, viper.GetString("crimson_settings.data_rootpath"))
	crimsonQ.SetConcurrency(consumerId, consumerConcurrent)
	crimsonQ.SetTopics(consumerId, consumerTopics)
	con.WriteString("Created:" + consumerId)
	return nil
}

func RC_Set_Concurrency(con redcon.Conn, args ...[][]byte) error {
	consumerId := string(args[0][0])
	concurrency := string(args[0][1])
	if crimsonQ.ConsumerExists(consumerId) {
		crimsonQ.SetConcurrency(consumerId, concurrency)
		con.WriteString("ok")
	} else {
		err := errors.New(Defs.ERRincorrectConsumerId)
		con.WriteError(fmt.Sprint(err))
		return err
	}
	return nil
}

func RC_SetConsumerTopics(con redcon.Conn, args ...[][]byte) error {
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

func RC_ConsumerConcurrencyOk(con redcon.Conn, args ...[][]byte) error {
	consumerId := string(args[0][0])
	if crimsonQ.ConsumerExists(consumerId) {
		con.WriteString(strconv.FormatBool(crimsonQ.ConcurrencyOk(consumerId)))
	} else {
		err := errors.New(Defs.ERRincorrectConsumerId)
		con.WriteError(fmt.Sprint(err))
		return err
	}
	return nil
}

func RC_GetConsumerTopics(con redcon.Conn, args ...[][]byte) error {
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
func RC_Destroy(con redcon.Conn, args ...[][]byte) error {
	consumerId := string(args[0][0])
	if crimsonQ.ConsumerExists(consumerId) {
		crimsonQ.DestroyQDB(consumerId)
		con.WriteString("ok")
	} else {
		con.WriteError(Defs.ERRincorrectConsumerId)
	}
	return nil
}

func RC_List(con redcon.Conn, args ...[][]byte) error {
	clist := crimsonQ.ListConsumers()
	con.WriteArray(len(clist))
	for _, s := range clist {
		con.WriteBulkString(s)
	}
	return nil
}

func RC_Msg_Keys(con redcon.Conn, args ...[][]byte) error {
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

func RC_Msg_Counts(con redcon.Conn, args ...[][]byte) error {
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

func RC_Msg_Push_Topic(con redcon.Conn, args ...[][]byte) error {
	topic := string(args[0][0])
	message := string(args[0][1])
	consumers := crimsonQ.PushTopic(topic, message)
	for consumer, msgkey := range consumers {
		msgkeySplit := strings.Split(msgkey, ":")
		PS.Publish(consumer, msgkeySplit[1])
	}
	con.WriteString("Ok")
	return nil
}

func RC_Msg_Push_Consumer(con redcon.Conn, args ...[][]byte) error {
	consumerId := string(args[0][0])
	topic := string(args[0][0])
	message := string(args[0][1])
	if crimsonQ.ConsumerExists(consumerId) {
		msgkey := crimsonQ.PushConsumer(consumerId, "direct:"+topic, message)
		msgkeySplit := strings.Split(msgkey, ":")
		PS.Publish(consumerId, msgkeySplit[1])
		con.WriteString("Ok")
		return nil
	} else {
		err := errors.New(Defs.ERRincorrectConsumerId)
		con.WriteError(fmt.Sprint(err))
		return err
	}

}

func RC_Msg_Pull(con redcon.Conn, args ...[][]byte) error {
	consumerId := string(args[0][0])
	if crimsonQ.ConsumerExists(consumerId) {
		msg, err := crimsonQ.Pull(consumerId)
		if err != nil {
			con.WriteError(err.Error())
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
func RC_Msg_Del(con redcon.Conn, args ...[][]byte) error {
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

func RC_Msg_Get_Status_Json(con redcon.Conn, args ...[][]byte) error {
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

func RC_Msg_Fail(con redcon.Conn, args ...[][]byte) error {
	consumerId := string(args[0][0])
	messageId := string(args[0][1])
	errMsg := string(args[0][2])
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

func RC_Msg_Complete(con redcon.Conn, args ...[][]byte) error {
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

func RC_Msg_Retry(con redcon.Conn, args ...[][]byte) error {
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

func RC_Msg_Retry_All(con redcon.Conn, args ...[][]byte) error {
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

func RC_Flush_Complete(con redcon.Conn, args ...[][]byte) error {
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

func RC_Flush_Failed(con redcon.Conn, args ...[][]byte) error {
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

func RC_Info(con redcon.Conn, args ...[][]byte) error {
	info := []string{"CrimsonQ Server", "A NextGen Message Queue!"}
	con.WriteArray(len(info))
	for _, x := range info {
		con.WriteString(x)
	}
	return nil
}
