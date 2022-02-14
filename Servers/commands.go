package Servers

import (
	"ywadi/crimsonq/Defs"

	"github.com/gofiber/fiber/v2"
)

var Commands map[string]CommandRecord

type CommandRecord struct {
	Redcon_Function    interface{}
	HTTP_Function      fiber.Handler
	ArgsCmd            []string
	RequiresConsumerId bool
	HTTP_Method        string
}

func InitCommands() {
	Commands = map[string]CommandRecord{
		"ping":                     {HTTP_Function: HTTP_Ping, HTTP_Method: Defs.HTTP_GET, Redcon_Function: RC_Ping, ArgsCmd: []string{"messageString"}},
		"quit":                     {HTTP_Function: HTTP_Quit, HTTP_Method: Defs.HTTP_GET, Redcon_Function: RC_Quit, ArgsCmd: []string{}},
		"auth":                     {HTTP_Function: HTTP_Auth, HTTP_Method: Defs.HTTP_GET, Redcon_Function: RC_Auth, ArgsCmd: []string{"password"}},
		"command":                  {HTTP_Function: HTTP_Command, HTTP_Method: Defs.HTTP_GET, Redcon_Function: RC_Command, ArgsCmd: []string{}},
		"info":                     {HTTP_Function: HTTP_Info, HTTP_Method: Defs.HTTP_GET, Redcon_Function: RC_Info, ArgsCmd: []string{}},
		"subscribe":                {HTTP_Function: HTTP_Subscribe, HTTP_Method: Defs.HTTP_GET, Redcon_Function: RC_Subscribe, ArgsCmd: []string{"consumerId"}},
		"consumer.exists":          {HTTP_Function: HTTP_Exists, HTTP_Method: Defs.HTTP_GET, Redcon_Function: RC_Exists, ArgsCmd: []string{"consumerId"}},
		"consumer.create":          {HTTP_Function: HTTP_Consumer_Create, HTTP_Method: Defs.HTTP_POST, Redcon_Function: RC_Consumer_Create, ArgsCmd: []string{"consumerId", "topics", "concurrency"}},
		"consumer.destroy":         {HTTP_Function: HTTP_Destroy, HTTP_Method: Defs.HTTP_POST, Redcon_Function: RC_Destroy, ArgsCmd: []string{"consumerId"}},
		"consumer.list":            {HTTP_Function: HTTP_List, HTTP_Method: Defs.HTTP_GET, Redcon_Function: RC_List, ArgsCmd: []string{}},
		"msg.keys":                 {HTTP_Function: HTTP_Msg_Keys, HTTP_Method: Defs.HTTP_GET, Redcon_Function: RC_Msg_Keys, ArgsCmd: []string{"consumerId"}},
		"msg.counts":               {HTTP_Function: HTTP_Msg_Counts, HTTP_Method: Defs.HTTP_GET, Redcon_Function: RC_Msg_Counts, ArgsCmd: []string{"consumerId"}},
		"msg.push.topic":           {HTTP_Function: HTTP_Msg_Push_Topic, HTTP_Method: Defs.HTTP_POST, Redcon_Function: RC_Msg_Push_Topic, ArgsCmd: []string{"topicString", "messageString"}},
		"msg.push.consumer":        {HTTP_Function: HTTP_Msg_Push_Consumer, HTTP_Method: Defs.HTTP_POST, Redcon_Function: RC_Msg_Push_Consumer, ArgsCmd: []string{"consumerId", "messageString"}},
		"msg.pull":                 {HTTP_Function: HTTP_Msg_Pull, HTTP_Method: Defs.HTTP_GET, Redcon_Function: RC_Msg_Pull, ArgsCmd: []string{"consumerId"}},
		"msg.del":                  {HTTP_Function: HTTP_Msg_Del, HTTP_Method: Defs.HTTP_POST, Redcon_Function: RC_Msg_Del, ArgsCmd: []string{"consumerId", "status", "messageId"}},
		"msg.fail":                 {HTTP_Function: HTTP_Msg_Fail, HTTP_Method: Defs.HTTP_POST, Redcon_Function: RC_Msg_Fail, ArgsCmd: []string{"consumerId", "messageId", "errMsg"}},
		"msg.complete":             {HTTP_Function: HTTP_Msg_Complete, HTTP_Method: Defs.HTTP_POST, Redcon_Function: RC_Msg_Complete, ArgsCmd: []string{"consumerId", "messageId"}},
		"msg.retry":                {HTTP_Function: HTTP_Msg_Retry, HTTP_Method: Defs.HTTP_POST, Redcon_Function: RC_Msg_Retry, ArgsCmd: []string{"consumerId", "messageId"}},
		"msg.retryall":             {HTTP_Function: HTTP_Msg_Retry_All, HTTP_Method: Defs.HTTP_POST, Redcon_Function: RC_Msg_Retry_All, ArgsCmd: []string{"consumerId"}},
		"consumer.flush.complete":  {HTTP_Function: HTTP_Flush_Complete, HTTP_Method: Defs.HTTP_POST, Redcon_Function: RC_Flush_Complete, ArgsCmd: []string{"consumerId"}},
		"consumer.flush.failed":    {HTTP_Function: HTTP_Flush_Failed, HTTP_Method: Defs.HTTP_POST, Redcon_Function: RC_Flush_Failed, ArgsCmd: []string{"consumerId"}},
		"consumer.topics.set":      {HTTP_Function: HTTP_SetConsumerTopics, HTTP_Method: Defs.HTTP_POST, Redcon_Function: RC_SetConsumerTopics, ArgsCmd: []string{"consumerId", "topics"}},
		"consumer.topics.get":      {HTTP_Function: HTTP_GetConsumerTopics, HTTP_Method: Defs.HTTP_GET, Redcon_Function: RC_GetConsumerTopics, ArgsCmd: []string{"consumerId"}},
		"consumer.concurrency.set": {HTTP_Function: HTTP_Set_Concurrency, HTTP_Method: Defs.HTTP_POST, Redcon_Function: RC_Set_Concurrency, ArgsCmd: []string{"consumerId", "concurrency"}},
		"msg.list.json":            {HTTP_Function: HTTP_Msg_Get_Status_Json, HTTP_Method: Defs.HTTP_GET, Redcon_Function: RC_Msg_Get_Status_Json, ArgsCmd: []string{"consumerId", "status"}},
		"consumer.concurrency.ok":  {HTTP_Function: HTTP_ConsumerConcurrencyOk, HTTP_Method: Defs.HTTP_GET, Redcon_Function: RC_ConsumerConcurrencyOk, ArgsCmd: []string{"consumerId"}},
		"consumer.info":            {HTTP_Function: HTTP_ConsumerInfo, HTTP_Method: Defs.HTTP_GET, Redcon_Function: RC_Consumer_Info, ArgsCmd: []string{"consumerId"}},
	}
}
