package Servers

var Commands map[string]CommandRecord

type CommandRecord struct {
	Redcon_Function    interface{}
	HTTP_Function      interface{}
	ArgsCmd            []string
	RequiresConsumerId bool
}

func InitCommands() {
	Commands = map[string]CommandRecord{
		"ping":                     {Redcon_Function: RC_Ping, ArgsCmd: []string{"messageString"}},
		"quit":                     {Redcon_Function: RC_Quit, ArgsCmd: []string{}},
		"auth":                     {Redcon_Function: RC_Auth, ArgsCmd: []string{"password"}},
		"command":                  {Redcon_Function: RC_Command, ArgsCmd: []string{}},
		"info":                     {Redcon_Function: RC_Info, ArgsCmd: []string{}},
		"subscribe":                {Redcon_Function: RC_Subscribe, ArgsCmd: []string{"consumerId"}},
		"consumer.exists":          {Redcon_Function: RC_Exists, ArgsCmd: []string{"consumerId"}},
		"consumer.create":          {Redcon_Function: RC_Consumer_Create, ArgsCmd: []string{"consumerId", "topics", "concurrency"}},
		"consumer.destroy":         {Redcon_Function: RC_Destroy, ArgsCmd: []string{"consumerId"}},
		"consumer.list":            {Redcon_Function: RC_List, ArgsCmd: []string{}},
		"msg.keys":                 {Redcon_Function: RC_Msg_Keys, ArgsCmd: []string{"consumerId"}},
		"msg.counts":               {Redcon_Function: RC_Msg_Counts, ArgsCmd: []string{"consumerId"}},
		"msg.push.topic":           {Redcon_Function: RC_Msg_Push_Topic, ArgsCmd: []string{"topicString", "messageString"}},
		"msg.push.consumer":        {Redcon_Function: RC_Msg_Push_Consumer, ArgsCmd: []string{"consumerId", "messageString"}},
		"msg.pull":                 {Redcon_Function: RC_Msg_Pull, ArgsCmd: []string{"consumerId"}},
		"msg.del":                  {Redcon_Function: RC_Msg_Del, ArgsCmd: []string{"consumerId", "status", "messageId"}},
		"msg.fail":                 {Redcon_Function: RC_Msg_Fail, ArgsCmd: []string{"consumerId", "messageId", "errMsg"}},
		"msg.complete":             {Redcon_Function: RC_Msg_Complete, ArgsCmd: []string{"consumerId", "messageId"}},
		"msg.retry":                {Redcon_Function: RC_Msg_Retry, ArgsCmd: []string{"consumerId", "messageId"}},
		"msg.retryall":             {Redcon_Function: RC_Msg_Retry_All, ArgsCmd: []string{"consumerId"}},
		"consumer.flush.complete":  {Redcon_Function: RC_Flush_Complete, ArgsCmd: []string{"consumerId"}},
		"consumer.flush.failed":    {Redcon_Function: RC_Flush_Failed, ArgsCmd: []string{"consumerId"}},
		"consumer.topics.set":      {Redcon_Function: RC_SetConsumerTopics, ArgsCmd: []string{"consumerId", "topics"}},
		"consumer.topics.get":      {Redcon_Function: RC_GetConsumerTopics, ArgsCmd: []string{"consumerId"}},
		"consumer.concurrency.set": {Redcon_Function: RC_Set_Concurrency, ArgsCmd: []string{"consumerId", "concurrency"}},
		"msg.list.json":            {Redcon_Function: RC_Msg_Get_Status_Json, ArgsCmd: []string{"consumerId", "status"}},
		"consumer.concurrency.ok":  {Redcon_Function: RC_ConsumerConcurrencyOk, ArgsCmd: []string{"consumerId"}},
	}
}
