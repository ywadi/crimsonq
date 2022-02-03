package Structs

import (
	"errors"
	"log"
	"strings"
	"sync"
	"time"
	"ywadi/goq/DButils"
	"ywadi/goq/Defs"
	"ywadi/goq/Utils"

	"github.com/dgraph-io/badger/v3"
	"github.com/spf13/viper"
)

type S_GOQ struct {
	QDBPool  map[string]*S_QDB
	SystemDb *badger.DB
}

var wg sync.WaitGroup

func (goq *S_GOQ) Init() {
	wg.Add(1)
	goq.StartWatchDog()
	go func() {
		goq.QDBPool = make(map[string]*S_QDB)
		Utils.PrintANSIlogo()
		//Open System
		systemDB, err := DButils.CreateDb("_systemDB", viper.GetString("crimson_settings.system_db_path"))
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

func (goq *S_GOQ) StartWatchDog() {
	println("Watchdog Started...")
	ticker := time.NewTicker(time.Duration(viper.GetInt64("crimson_settings.watchdog_seconds")) * time.Second)
	done := make(chan bool)
	go func() {
		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				for _, s := range goq.QDBPool {
					s.ExpireQmsgFromStatus()
				}

			}
		}
	}()
}

func (goq *S_GOQ) CreateQDB(consumerId string, QDBpath string, QdbTopicFilters string) {
	var qdb S_QDB
	topicFilters := strings.Split(QdbTopicFilters, ",")
	qdb.Init(consumerId, QDBpath, topicFilters)
	DButils.SET(goq.SystemDb, Defs.QDB_PREFIX+consumerId, qdb.Serialize())
	goq.QDBPool[consumerId] = &qdb
}

func (goq *S_GOQ) ListConsumers() []string {
	var cList []string
	for k := range goq.QDBPool {
		cList = append(cList, k)
	}
	return cList
}

func (goq *S_GOQ) ConsumerExists(consumerId string) bool {
	if _, ok := goq.QDBPool[consumerId]; ok {
		return true
	} else {
		return false
	}
}

//Push to consumer
func (goq *S_GOQ) PushConsumer(consumerId string, topic string, message string) {
	consumerQ := goq.QDBPool[consumerId]
	consumerQ.CreateAndPushQMSG(topic, message)
}

//Push to topic
func (goq *S_GOQ) PushTopic(topic string, message string) []*S_QDB {
	consumersPushed := []*S_QDB{}
	consumers := goq.QDBPool
	for _, s := range consumers {
		for _, x := range s.QdbTopicFilters {
			if Utils.MQTTMatch(x, topic) {
				s.CreateAndPushQMSG(topic, message)
				consumersPushed = append(consumersPushed, s)
			}
		}
	}
	return consumersPushed
}

//Pull from consumer
//Needs mutex lock to avoid overlapping behavior with subscribe
func (goq *S_GOQ) Pull(consumerId string) (*S_QMSG, error) {

	consumerQ := goq.QDBPool[consumerId]
	qmg, err := consumerQ.Pull()
	if err != nil {
		return nil, err
	}
	return qmg, nil
}

//MarkMSGIDFailed
func (goq *S_GOQ) MsgFail(consumerId string, msgKey string, errMsg string) error {
	consumerQ := goq.QDBPool[consumerId]
	err := consumerQ.MarkFailed(msgKey, errMsg)
	if err != nil {
		return err
	}
	return nil
}

//MarkMSGIDComplete
func (goq *S_GOQ) MsgComplete(consumerId string, msgKey string) error {
	consumerQ := goq.QDBPool[consumerId]
	err := consumerQ.MarkCompleted(msgKey)
	if err != nil {
		return err
	}
	return nil
}

//RetryMSGIDFailed
func (goq *S_GOQ) MsgRetry(consumerId string, msgKey string) error {
	consumerQ := goq.QDBPool[consumerId]
	err := consumerQ.RetryFailed(msgKey)
	if err != nil {
		return err
	}
	return nil
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

func (goq *S_GOQ) ListAllKeys(consumerId string) ([]string, error) {
	consumerQ := goq.QDBPool[consumerId]
	if goq.ConsumerExists(consumerId) {
		return consumerQ.GetAllKeys(), nil
	}
	return nil, errors.New("001:incorrect_consumer_id")
}

func (goq *S_GOQ) Del(consumerId string, messageId string) error {
	consumerQ := goq.QDBPool[consumerId]
	err := consumerQ.Del(messageId)
	if err != nil {
		return err
	}
	return err
}
