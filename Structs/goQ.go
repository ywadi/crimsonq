package Structs

import (
	"errors"
	"strconv"
	"strings"
	"sync"
	"time"
	"ywadi/crimsonq/DButils"
	"ywadi/crimsonq/Defs"
	"ywadi/crimsonq/Utils"

	log "github.com/sirupsen/logrus"

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
	goq.StartDiskSyncTime()
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
	log.Info("Watch Dog Started")
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
					if time.Since(s.Last_Active_Pull) > (time.Hour * time.Duration(viper.GetInt("crimson_settings.consumer_inactive_destroy_hours"))) {
						goq.DestroyQDB(s.QdbId)
					}
				}

			}
		}
	}()
}

func (goq *S_GOQ) StartDiskSyncTime() {
	log.Info("DiskSync Timer Started...")
	if !viper.GetBool("crimson_settings.db_full_persist") {
		ticker := time.NewTicker(time.Duration(viper.GetInt64("crimson_settings.disk_sync_seconds")) * time.Second)
		done := make(chan bool)
		go func() {
			for {
				select {
				case <-done:
					return
				case <-ticker.C:
					for _, s := range goq.QDBPool {
						log.WithFields(log.Fields{"ConsumerId": s.QdbId}).Info("Synced data to disk:" + s.QdbId)
						s.DB().Sync()
					}

				}
			}
		}()
	}

}

func (goq *S_GOQ) CreateQDB(consumerId string, QDBpath string) {
	var qdb S_QDB
	qdb.Last_Active_Pull = time.Now()
	qdb.Init(consumerId, QDBpath)
	DButils.SET(goq.SystemDb, Defs.QDB_PREFIX+consumerId, qdb.Serialize())
	goq.QDBPool[consumerId] = &qdb
}

func (goq *S_GOQ) UpdateQDBinDB(consumerId string) {
	consumerQ := goq.QDBPool[consumerId]
	DButils.SET(goq.SystemDb, Defs.QDB_PREFIX+consumerId, consumerQ.Serialize())
}

func (goq *S_GOQ) SetTopics(consumerId string, topics string) {
	topicsArray := strings.Split(topics, ",")
	consumerQ := goq.QDBPool[consumerId]
	consumerQ.QdbTopicFilters = topicsArray
	goq.UpdateQDBinDB(consumerId)
}

func (goq *S_GOQ) SetConcurrency(consumerId string, concurrency string) {
	concurrencyInt, err := strconv.Atoi(concurrency)
	if err != nil {
		// handle error
		log.Error("ERROR parsing to int")
	}
	consumerQ := goq.QDBPool[consumerId]
	consumerQ.Concurrency = concurrencyInt
	goq.UpdateQDBinDB(consumerId)
}

func (goq *S_GOQ) SetLastPullDate(consumerId string) {
	consumerQ := goq.QDBPool[consumerId]
	consumerQ.Last_Active_Pull = time.Now()
	goq.UpdateQDBinDB(consumerId)
}

func (goq *S_GOQ) GetTopics(consumerId string) []string {
	consumerQ := goq.QDBPool[consumerId]
	return consumerQ.GetTopics()
}

func (goq *S_GOQ) DestroyQDB(consumerId string) {
	goq.QDBPool[consumerId].Destroy()
	delete(goq.QDBPool, consumerId)
	DButils.DEL(goq.SystemDb, Defs.QDB_PREFIX+consumerId)
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
func (goq *S_GOQ) PushConsumer(consumerId string, topic string, message string) string {
	consumerQ := goq.QDBPool[consumerId]
	qmsg := consumerQ.CreateAndPushQMSG(topic, message)
	return qmsg.Key
}

//Push to topic
func (goq *S_GOQ) PushTopic(topic string, message string) map[string]string {
	res := make(map[string]string)

	consumers := goq.QDBPool
	for _, s := range consumers {
		for _, x := range s.QdbTopicFilters {
			if Utils.MQTTMatch(topic, x) {
				qmsg := s.CreateAndPushQMSG(topic, message)
				res[s.QdbId] = qmsg.Key
			}
		}
	}
	return res
}

//Pull from consumer
func (goq *S_GOQ) Pull(consumerId string) (*S_QMSG, error) {
	consumerQ := goq.QDBPool[consumerId]
	goq.SetLastPullDate(consumerId)
	qmg, err := consumerQ.Pull()
	if err != nil {
		return nil, err
	}
	return qmg, nil
}

func (goq *S_GOQ) ConcurrencyOk(consumerId string) bool {
	consumerQ := goq.QDBPool[consumerId]
	return !consumerQ.ConcurrencyBOEActive()
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
		keys, _ := consumerQ.GetAllKeys()
		return keys, nil
	}
	return nil, errors.New(Defs.ERRincorrectConsumerId)
}

func (goq *S_GOQ) GetAllByStatusJson(consumerId string, status string) (string, error) {
	consumerQ := goq.QDBPool[consumerId]
	if goq.ConsumerExists(consumerId) {
		json, err := consumerQ.GetMsgByStatusJson(status)
		if err != nil {
			return "", err
		}
		return json, nil
	} else {
		return "", errors.New(Defs.ERRincorrectConsumerId)
	}
}

func (goq *S_GOQ) GetKeyCount(consumerId string) (map[string]uint16, error) {
	consumerQ := goq.QDBPool[consumerId]
	//Update last active if a get key count was requested.
	//Used by heartbeat hence if beating then should stay alive
	consumerQ.Last_Active_Pull = time.Now()
	goq.UpdateQDBinDB(consumerId)
	if goq.ConsumerExists(consumerId) {
		_, Count := consumerQ.GetAllKeys()
		return Count, nil
	}
	return nil, errors.New("001:incorrect_consumer_id")
}

func (goq *S_GOQ) Del(status string, consumerId string, messageId string) error {
	consumerQ := goq.QDBPool[consumerId]
	err := consumerQ.Del(status, messageId)
	if err != nil {
		return err
	}
	return err
}
