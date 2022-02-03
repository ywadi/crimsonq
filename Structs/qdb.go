//TODO Event manager for all actions and to go to logs
package Structs

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"
	"ywadi/goq/DButils"
	"ywadi/goq/Defs"
	"ywadi/goq/Utils"

	"github.com/dgraph-io/badger/v3"
)

type S_QDB struct {
	QdbId           string
	QdbPath         string
	QdbTopicFilters []string
}

var DBpool map[string]*badger.DB

func (qdb *S_QDB) Init(QdbId string, QdbPath string, QdbTopicFilters []string) {
	qdb.QdbId = QdbId
	qdb.QdbPath = QdbPath
	qdb.QdbTopicFilters = QdbTopicFilters
	DBpool = make(map[string]*badger.DB)
	qdb.CreateDB()
}

func (qdb *S_QDB) Destroy() {
	DBpool[qdb.QdbId].Close()
	delete(DBpool, qdb.QdbId)
}

func (qdb *S_QDB) Deserialize(data []byte) {
	DBpool = make(map[string]*badger.DB)
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
	qdb.SetDB(db) //TODO: Figure out why flagged
}

func (qdb *S_QDB) AddTopicFilter(newTopicFilter string) {
	qdb.QdbTopicFilters = append(qdb.QdbTopicFilters, newTopicFilter)
}

//TODO
func (qdb *S_QDB) RemoveTopicFilter() {
	// for i, s := range qdb.QdbTopicFilters {
	// 	//TODO: implement slice, is it by ref or var the s?!
	// }
}

func (qdb *S_QDB) MoveMsg(key string, from string, to string, err string) (*S_QMSG, error) {
	value, er := DButils.GET(qdb.DB(), key)
	if er != nil {
		return nil, er
	}
	var qmsg S_QMSG
	if !bytes.Equal(value, []byte{}) {
		print("!!!!", "[", value, "]", bytes.Equal(value, []byte{}))
		qmsg.Deserialize(value)
		qmsg.StatusHistory[to+"_at"] = time.Now()
		qmsg.Status = to
		qmsg.Key = to + ":" + qmsg.RawKey
		if err != "" {
			qmsg.Error = err
		}
		newKey := strings.Replace(key, from, to, 1)
		DButils.DEL(qdb.DB(), key)
		ser := Utils.Serialize(qmsg)
		DButils.SET(qdb.DB(), newKey, ser)
		return &qmsg, nil
	} else {
		return nil, errors.New("no data returned")
	}
}

func (qdb *S_QDB) ExpireQmsgFromStatus() {
	//TODO Settings for duration
	qdb.MoveBatchOlderThan(Defs.STATUS_ACTIVE, Defs.STATUS_DELAYED, 10*time.Second)
	qdb.MoveBatchOlderThan(Defs.STATUS_DELAYED, Defs.STATUS_FAILED, 10*time.Second)
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
					qdb.MoveMsg(string(item.Key()), from, to, "Job took too long to execute")
					fmt.Println("Moved", string(item.Key()), "from:", from, "to:", to, "duration:", time.Since(cQmsg.StatusHistory[from+"_at"]))
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
func (qdb *S_QDB) Pull() (*S_QMSG, error) {
	//Get message from Pending and add to Active
	//Return message and then turn to JSON
	k, _, err := DButils.DEQ(qdb.DB())
	if err != nil {
		return nil, err
	}

	msgRes, err := qdb.MoveMsg(string(k), Defs.STATUS_PENDING, Defs.STATUS_ACTIVE, "")
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
	_, err := qdb.MoveMsg(key, Defs.STATUS_FAILED, Defs.STATUS_PENDING, "")
	if err != nil {
		return err
	}
	return nil
}

func (qdb *S_QDB) MarkCompleted(key string) error {
	//Get Message from Delayed or Pending and add to Complete
	//TODO IF key is there or not, needs to be managed
	_, err := qdb.MoveMsg(key, Defs.STATUS_ACTIVE, Defs.STATUS_COMPLETED, "")
	if err != nil {
		return err
	}
	_, err = qdb.MoveMsg(key, Defs.STATUS_PENDING, Defs.STATUS_COMPLETED, "")
	if err != nil {
		return err
	}
	return nil
}
func (qdb *S_QDB) MarkFailed(key string) error {
	//Get Message from Delayed or Pending and add to Failed
	//TODO IFs
	_, err := qdb.MoveMsg(key, Defs.STATUS_ACTIVE, Defs.STATUS_FAILED, "")
	if err != nil {
		return err
	}
	_, err = qdb.MoveMsg(key, Defs.STATUS_PENDING, Defs.STATUS_FAILED, "")
	if err != nil {
		return err
	}
	return nil
}

func (qdb *S_QDB) CreateAndPushQMSG(topic string, message string) {
	qmsg := new(S_QMSG)
	qmsg.RawKey = DButils.GetNextKey(qdb.DB())
	qmsg.Key = (Defs.STATUS_PENDING + ":" + qmsg.RawKey)
	qmsg.Value = message
	qmsg.Status = Defs.STATUS_PENDING
	qmsg.Topic = topic
	qmsg.StatusHistory = make(map[string]time.Time)
	qmsg.StatusHistory[Defs.CREATED_AT] = time.Now()
	println(qmsg.Key)
	qdb.Push(*qmsg)
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

func (qdb *S_QDB) GetAllKeys() []string {
	msgKeyList := []string{}
	db := qdb.DB()
	b := DButils.GetAllPrefix(db, Defs.STATUS_ACTIVE)
	print(b)
	for _, s := range b {
		newMSG := S_QMSG{}
		newMSG.Deserialize(s)
		msgKeyList = append(msgKeyList, newMSG.Key)
	}

	b = DButils.GetAllPrefix(db, Defs.STATUS_COMPLETED)
	for _, s := range b {
		newMSG := S_QMSG{}
		newMSG.Deserialize(s)
		msgKeyList = append(msgKeyList, newMSG.Key)
	}

	b = DButils.GetAllPrefix(db, Defs.STATUS_DELAYED)
	for _, s := range b {
		newMSG := S_QMSG{}
		newMSG.Deserialize(s)
		msgKeyList = append(msgKeyList, newMSG.Key)
	}

	b = DButils.GetAllPrefix(db, Defs.STATUS_FAILED)
	for _, s := range b {
		newMSG := S_QMSG{}
		newMSG.Deserialize(s)
		msgKeyList = append(msgKeyList, newMSG.Key)
	}

	b = DButils.GetAllPrefix(db, Defs.STATUS_PENDING)
	for _, s := range b {
		newMSG := S_QMSG{}
		newMSG.Deserialize(s)
		msgKeyList = append(msgKeyList, newMSG.Key)
	}
	return msgKeyList
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

func (qdb *S_QDB) Del(messageId string) error {
	err := DButils.DEL(qdb.DB(), messageId)
	if err != nil {
		return err
	}
	return nil
}
