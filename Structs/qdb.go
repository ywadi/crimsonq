package Structs

import (
	"bytes"
	"encoding/json"
	"errors"
	"strings"
	"time"
	"ywadi/crimsonq/DButils"
	"ywadi/crimsonq/Defs"
	"ywadi/crimsonq/Utils"

	log "github.com/sirupsen/logrus"

	"github.com/dgraph-io/badger/v3"
	"github.com/spf13/viper"
)

type S_QDB struct {
	QdbId            string    `json:"consumerId"`
	QdbPath          string    `json:"-"`
	QdbTopicFilters  []string  `json:"topics"`
	Last_Active_Pull time.Time `json:"lastActive"`
	Concurrency      int       `json:"concurrency"`
}

var DBpool map[string]*badger.DB

func initDbPool() {
	if DBpool == nil {
		DBpool = make(map[string]*badger.DB)
	}
}

func (qdb *S_QDB) Init(QdbId string, QdbPath string) {
	qdb.QdbId = QdbId
	qdb.QdbPath = QdbPath
	initDbPool()
	qdb.CreateDB()
	//RedconQ.PS.Publish("_system", fmt.Sprint("Initiated ", QdbId, time.Now()))
}

func (qdb *S_QDB) JsonStringify() string {
	return Utils.ToJson(qdb)
}

func (qdb *S_QDB) GetTopics() []string {
	return qdb.QdbTopicFilters
}

func (qdb *S_QDB) Destroy() {
	delete(DBpool, qdb.QdbId)
	DButils.DestroyDb(qdb.QdbId, qdb.QdbPath)
}

func (qdb *S_QDB) Deserialize(data []byte) {
	initDbPool()
	Utils.Deserialize(data, qdb)
	qdb.CreateDB()
}

func (qdb *S_QDB) Serialize() []byte {
	return Utils.Serialize(qdb)
}

func (qdb *S_QDB) DB() *badger.DB {
	return DBpool[qdb.QdbId]
}

func (qdb *S_QDB) SetDB(db *badger.DB) {
	DBpool[qdb.QdbId] = db
}

func (qdb *S_QDB) CreateDB() {
	db, err := DButils.CreateDb(qdb.QdbId, qdb.QdbPath)
	if err != nil {
		log.Fatal(err)
	}
	qdb.SetDB(db)
}

func (qdb *S_QDB) MoveMsg(key string, from string, to string, errMsg string) (*S_QMSG, error) {
	key = from + ":" + key
	value, er := DButils.GET(qdb.DB(), key)
	if er != nil {
		return nil, er
	}
	var qmsg S_QMSG
	if !bytes.Equal(value, []byte{}) {
		qmsg.Deserialize(value)
		qmsg.StatusHistory[to+"_at"] = time.Now()
		qmsg.Status = to
		qmsg.Key = to + ":" + qmsg.RawKey
		if errMsg != "" {
			qmsg.Error = errMsg
		}
		newKey := strings.Replace(key, from, to, 1)
		DButils.DEL(qdb.DB(), key)
		ser := Utils.Serialize(qmsg)
		DButils.SET(qdb.DB(), newKey, ser)
		return &qmsg, nil
	} else {
		return nil, errors.New(Defs.ERRnoDataReturn)
	}
}

func (qdb *S_QDB) ExpireQmsgFromStatus() {
	qdb.MoveBatchOlderThan(Defs.STATUS_ACTIVE, Defs.STATUS_DELAYED, time.Duration(viper.GetInt64("crimson_settings.active_before_delayed"))*time.Second)
	qdb.MoveBatchOlderThan(Defs.STATUS_DELAYED, Defs.STATUS_FAILED, time.Duration(viper.GetInt64("crimson_settings.delayed_before_failed"))*time.Second)
}

func (qdb *S_QDB) MoveBatchOlderThan(from string, to string, duration time.Duration) {
	qdb.DB().View(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()
		prefix := []byte(from)
		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			item := it.Item()
			err := item.Value(func(v []byte) error {
				valCopy, err := item.ValueCopy(nil)
				if err != nil {
					return err
				}
				var cQmsg S_QMSG
				cQmsg.Deserialize(valCopy)
				if time.Since(cQmsg.StatusHistory[from+"_at"]) > duration {
					rawKey := strings.Split(string(item.Key()), ":")
					qdb.MoveMsg(rawKey[1], from, to, "Job took too long to execute")
					log.Info("Moved", string(item.Key()), "from:", from, "to:", to, "duration:", time.Since(cQmsg.StatusHistory[from+"_at"]))
				}
				return nil
			})
			if err != nil {
				return err
			}
		}
		return nil
	})
}

func (qdb *S_QDB) Push(qmsg S_QMSG) {
	//Push message to pending
	DButils.SET(qdb.DB(), qmsg.Key, qmsg.Serialize())
}

func (qdb *S_QDB) ConcurrencyBOEActive() bool {
	db := qdb.DB()
	b := DButils.GetAllPrefix(db, Defs.STATUS_ACTIVE)
	if qdb.Concurrency <= 0 {
		return false
	}
	return qdb.Concurrency <= len(b)
}

func (qdb *S_QDB) Pull() (*S_QMSG, error) {
	//Get message from Pending and add to Active
	//Return message and then turn to JSON
	if qdb.ConcurrencyBOEActive() {
		return nil, errors.New(Defs.ERRExceededConcurrency)
	}
	k, _, err := DButils.DEQ(qdb.DB())
	if err != nil {
		return nil, err
	}
	keySplit := strings.Split(string(k), ":")
	if len(keySplit) < 2 {
		return nil, errors.New("empty queue")
	}
	msgRes, err := qdb.MoveMsg(keySplit[1], Defs.STATUS_PENDING, Defs.STATUS_ACTIVE, "")
	if err != nil {
		return nil, err
	}
	return msgRes, nil
}
func (qdb *S_QDB) MarkDelayed(key string) {
	//Get Message from Pending and add to Delayed
	qdb.MoveMsg(key, Defs.STATUS_ACTIVE, Defs.STATUS_DELAYED, "")

}

func (qdb *S_QDB) RetryFailed(key string) error {
	splitKey := strings.Split(key, ":")
	var lkey string
	if len(splitKey) < 2 {
		lkey = splitKey[0]
	} else {
		lkey = splitKey[1]
	}
	_, err := qdb.MoveMsg(lkey, Defs.STATUS_FAILED, Defs.STATUS_PENDING, "")
	if err != nil {
		return err
	}
	return nil
}

func (qdb *S_QDB) MarkCompleted(key string) error {
	//Get Message from Delayed or Pending and add to Complete
	_, err1 := qdb.MoveMsg(key, Defs.STATUS_ACTIVE, Defs.STATUS_COMPLETED, "")
	if err1 == nil {
		return nil
	}
	_, err2 := qdb.MoveMsg(key, Defs.STATUS_DELAYED, Defs.STATUS_COMPLETED, "")
	if err2 == nil {
		return nil
	}
	return errors.New(err1.Error() + "::" + err2.Error())
}
func (qdb *S_QDB) MarkFailed(key string, errMsg string) error {
	_, err1 := qdb.MoveMsg(key, Defs.STATUS_ACTIVE, Defs.STATUS_FAILED, errMsg)
	if err1 == nil {
		return nil
	}

	_, err2 := qdb.MoveMsg(key, Defs.STATUS_PENDING, Defs.STATUS_FAILED, errMsg)
	if err2 == nil {
		return nil
	}

	return errors.New(err1.Error() + "::" + err2.Error())
}

func (qdb *S_QDB) CreateAndPushQMSG(topic string, message string) *S_QMSG {
	qmsg := new(S_QMSG)
	qmsg.RawKey = Utils.GenerateULID()
	qmsg.Key = (Defs.STATUS_PENDING + ":" + qmsg.RawKey)
	qmsg.Value = message
	qmsg.Status = Defs.STATUS_PENDING
	qmsg.Topic = topic
	qmsg.StatusHistory = make(map[string]time.Time)
	qmsg.StatusHistory[Defs.CREATED_AT] = time.Now()
	qdb.Push(*qmsg)
	return qmsg
}

func (qdb *S_QDB) ClearComplete() {
	DButils.ClearPrefix(qdb.DB(), Defs.STATUS_COMPLETED)
}

func (qdb *S_QDB) ClearFailed() {
	DButils.ClearPrefix(qdb.DB(), Defs.STATUS_FAILED)
}

func (qdb *S_QDB) GetAllFailed() ([]S_QMSG, error) {
	msgList := []S_QMSG{}

	b := DButils.GetAllPrefix(qdb.DB(), Defs.STATUS_FAILED)
	for _, s := range b {
		newMSG := S_QMSG{}
		newMSG.Deserialize(s)
		msgList = append(msgList, newMSG)
	}
	return msgList, nil
}

func (qdb *S_QDB) GetMsgByStatusJson(status string) (string, error) {
	msgList := []S_QMSG{}
	if !(status == Defs.STATUS_ACTIVE || status == Defs.STATUS_COMPLETED || status == Defs.STATUS_DELAYED || status == Defs.STATUS_FAILED || status == Defs.STATUS_PENDING) {
		return "", errors.New(Defs.ERRIncorrectStatus)
	}
	b := DButils.GetAllPrefix(qdb.DB(), status)
	for _, s := range b {
		newMSG := S_QMSG{}
		newMSG.Deserialize(s)
		msgList = append(msgList, newMSG)
	}
	json, err := json.Marshal(msgList)
	if err != nil {
		return "", err
	}
	return string(json), nil
}

func (qdb *S_QDB) GetAllKeys() (keys []string, count map[string]uint16) {
	msgKeyList := []string{}
	msgKeyCounts := make(map[string]uint16)
	db := qdb.DB()
	b := DButils.GetAllPrefix(db, Defs.STATUS_ACTIVE)
	msgKeyCounts[Defs.STATUS_ACTIVE] = 0
	for _, s := range b {
		newMSG := S_QMSG{}
		newMSG.Deserialize(s)
		msgKeyList = append(msgKeyList, newMSG.Key)
		msgKeyCounts[Defs.STATUS_ACTIVE]++
	}

	b = DButils.GetAllPrefix(db, Defs.STATUS_COMPLETED)
	msgKeyCounts[Defs.STATUS_COMPLETED] = 0
	for _, s := range b {
		newMSG := S_QMSG{}
		newMSG.Deserialize(s)
		msgKeyList = append(msgKeyList, newMSG.Key)
		msgKeyCounts[Defs.STATUS_COMPLETED]++
	}

	b = DButils.GetAllPrefix(db, Defs.STATUS_DELAYED)
	msgKeyCounts[Defs.STATUS_DELAYED] = 0
	for _, s := range b {
		newMSG := S_QMSG{}
		newMSG.Deserialize(s)
		msgKeyList = append(msgKeyList, newMSG.Key)
		msgKeyCounts[Defs.STATUS_DELAYED]++
	}

	b = DButils.GetAllPrefix(db, Defs.STATUS_FAILED)
	msgKeyCounts[Defs.STATUS_FAILED] = 0
	for _, s := range b {
		newMSG := S_QMSG{}
		newMSG.Deserialize(s)
		msgKeyList = append(msgKeyList, newMSG.Key)
		msgKeyCounts[Defs.STATUS_FAILED]++
	}

	b = DButils.GetAllPrefix(db, Defs.STATUS_PENDING)
	msgKeyCounts[Defs.STATUS_PENDING] = 0
	for _, s := range b {
		newMSG := S_QMSG{}
		newMSG.Deserialize(s)
		msgKeyList = append(msgKeyList, newMSG.Key)
		msgKeyCounts[Defs.STATUS_PENDING]++
	}
	return msgKeyList, msgKeyCounts
}

func (qdb *S_QDB) RetryAllFailed() error {
	msgs, err := qdb.GetAllFailed()
	if err != nil {
		return err
	}
	for _, m := range msgs {
		err = qdb.RetryFailed(m.Key)
		if err != nil {
			return err
		}
	}
	return nil
}

func (qdb *S_QDB) Del(status string, messageId string) error {
	messageId = status + ":" + messageId
	err := DButils.DEL(qdb.DB(), messageId)
	if err != nil {
		return err
	}
	return nil
}
