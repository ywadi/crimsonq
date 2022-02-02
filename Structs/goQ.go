package Structs

import (
	"log"
	"sync"
	"ywadi/goq/DButils"
	"ywadi/goq/Defs"
	"ywadi/goq/Utils"

	"github.com/dgraph-io/badger/v3"
)

//Manage the QDBs in _systemdb
//CreateSystemDB has the QDBs and Settings, all as prefix

type S_GOQ struct {
	QDBPool  map[string]*S_QDB
	Settings string
	SystemDb *badger.DB
}

func (goq *S_GOQ) Init(settings string) {
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		Utils.PrintANSIlogo()
		//Open System Db
		goq.QDBPool = make(map[string]*S_QDB)
		goq.Settings = "Replace with settings"
		systemDB, err := DButils.CreateDb("_systemDB", "/home/ywadi/.goQ/_system")
		if err != nil {
			log.Fatal(err)
		}
		goq.SystemDb = systemDB
		qdbList := DButils.GetAllPrefix(systemDB, Defs.QDB_PREFIX)

		//Init QDBs and Load all QDBs into QDBpool
		for _, s := range qdbList {
			qdb := S_QDB{}
			qdb.Deserialize(s)
			goq.QDBPool[qdb.QdbId] = &qdb
		}
		//wg.Done() //TODO: Command off channel
	}()
	wg.Wait()
}

//Push to consumer
func (goq *S_GOQ) PushConsumer(consumerId string, topic string, message string) {
	consumerQ := goq.QDBPool[consumerId]
	consumerQ.CreateAndPushQMSG(topic, message)
}

//Push to topic
func (goq *S_GOQ) PushToipc(topic string, message string) {

}

//Pull from consumer
func (goq *S_GOQ) Pull(consumerId string) S_QMSG {
	consumerQ := goq.QDBPool[consumerId]
	qmg := consumerQ.Pull()
	return qmg
}

//MarkMSGIDFailed
func (goq *S_GOQ) MsgFail(consumerId string, msgKey string) {
	consumerQ := goq.QDBPool[consumerId]
	consumerQ.MarkFailed(msgKey)

}

//MarkMSGIDComplete
func (goq *S_GOQ) MsgComplete(consumerId string, msgKey string) {
	consumerQ := goq.QDBPool[consumerId]
	consumerQ.MarkCompleted(msgKey)
}

//RetryMSGIDFailed
func (goq *S_GOQ) MsgRetry(consumerId string, msgKey string) {
	consumerQ := goq.QDBPool[consumerId]
	consumerQ.RetryFailed(msgKey)
}

//RetryAllFailed
func (goq *S_GOQ) ReqAllFailed(consumerId string) {
	goq.QDBPool[consumerId].RetryAllFailed()
}

//ClearComplete
func (goq *S_GOQ) ClearComplete(consumerId string) {
	consumerQ := goq.QDBPool[consumerId]
	consumerQ.ClearComplete()
}

//ClearFailed
func (goq *S_GOQ) ClearFailed(consumerId string) {
	consumerQ := goq.QDBPool[consumerId]
	consumerQ.ClearFailed()
}
