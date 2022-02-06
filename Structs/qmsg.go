package Structs

import (
	"time"
	"ywadi/goq/Defs"
	"ywadi/goq/Utils"
)

type S_QMSG struct {
	Key           string
	RawKey        string               `json:"key"`
	Topic         string               `json:"topic"`
	Value         string               `json:"value"`
	Status        string               `json:"status"`
	StatusHistory map[string]time.Time `json:"statusHistory"`
	Error         string               `json:"error,omitempty"`
}

func (qmsg *S_QMSG) Init(Key string, RawKey string, Topic string, Value string) {
	qmsg.Key = Key
	qmsg.RawKey = RawKey
	qmsg.Topic = Topic
	qmsg.Value = Value
	qmsg.Status = Defs.STATUS_PENDING
	qmsg.StatusHistory = make(map[string]time.Time)
	qmsg.Error = ""
}

func (qmsg *S_QMSG) Serialize() []byte {
	return Utils.Serialize(qmsg)
}

func (qmsg *S_QMSG) Deserialize(data []byte) {
	Utils.Deserialize(data, qmsg)
}

func (qmsg *S_QMSG) JsonStringify() string {
	return Utils.ToJson(qmsg)
}
