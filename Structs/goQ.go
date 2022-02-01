package Structs

import (
	"log"
	"ywadi/goq/DButils"
	"ywadi/goq/Defs"

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

}

//Push to consumer
//Push to topic
//Pull from consumer
//MarkMSGIDFailed
//MarkMSGIDComplete
//RetryMSGIDFailed
//RetryAllFailed
//ClearComplete
//ClearFailed
