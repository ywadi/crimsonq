//TODO Event manager for all actions and to go to logs
package Structs

import (
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
	DB              *badger.DB
}

func (qdb *S_QDB) Init(QdbId string, QdbPath string, QdbTopicFilters []string) {
	qdb.QdbId = QdbId
	qdb.QdbPath = QdbPath
	qdb.QdbTopicFilters = QdbTopicFilters
	qdb.CreateDB()
}

func (qdb *S_QDB) Deserialize(data []byte) {
	Utils.Deserialize(data, qdb)
}

func (qdb *S_QDB) CreateDB() {
	db, err := DButils.CreateDb(qdb.QdbId, qdb.QdbPath)
	if err != nil {
		log.Fatal(err)
	}
	qdb.DB = db //TODO: Figure out why flagged
}

func (qdb *S_QDB) StartWatchDog() {
	ticker := time.NewTicker(1 * time.Second)
	done := make(chan bool)
	go func() {
		for {
			select {
			case <-done:
				return
			case t := <-ticker.C:
				fmt.Println("Tick at", t)
				qdb.ExpireQmsgFromStatus()
			}
		}
	}()
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

func (qdb *S_QDB) MoveMsg(key string, from string, to string, err string) {
	value := DButils.GET(qdb.DB, key)
	var qmsg S_QMSG
	qmsg.Deserialize(value)
	qmsg.StatusHistory[to+"_at"] = time.Now()
	qmsg.Status = to
	qmsg.Key = to + ":" + qmsg.RawKey
	if err != "" {
		qmsg.Error = err
	}
	newKey := strings.Replace(key, from, to, 1)
	DButils.DEL(qdb.DB, key)
	DButils.SET(qdb.DB, newKey, Utils.Serialize(qmsg))
}

func (qdb *S_QDB) ExpireQmsgFromStatus() {
	qdb.MoveBatchOlderThan(Defs.STATUS_ACTIVE, Defs.STATUS_DELAYED, 1000)
	qdb.MoveBatchOlderThan(Defs.STATUS_DELAYED, Defs.STATUS_FAILED, 1000)
}

func (qdb *S_QDB) MoveBatchOlderThan(from string, to string, duration time.Duration) {
	qdb.DB.View(func(txn *badger.Txn) error {
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
	DButils.SET(qdb.DB, qmsg.Key, qmsg.Serialize())
}
func (qdb *S_QDB) Pull() S_QMSG {
	//Get message from Pending and add to Active
	//Return message and then turn to JSON
	k, v := DButils.DEQ(qdb.DB)
	qdb.MoveMsg(string(k), Defs.STATUS_PENDING, Defs.STATUS_ACTIVE, "")
	var qmsg S_QMSG
	Utils.Deserialize(v, &qmsg)
	return qmsg
}
func (qdb *S_QDB) MarkDelayed(key string) {
	//Get Message from Pending and add to Delayed
	qdb.MoveMsg(key, Defs.STATUS_ACTIVE, Defs.STATUS_DELAYED, "")

}

func (qdb *S_QDB) RetryFailed(key string) {
	qdb.MoveMsg(key, Defs.STATUS_FAILED, Defs.STATUS_PENDING, "")
}

func (qdb *S_QDB) MarkCompleted(key string) {
	//Get Message from Delayed or Pending and add to Complete
	//TODO IFs
	qdb.MoveMsg(key, Defs.STATUS_ACTIVE, Defs.STATUS_COMPLETED, "")
	qdb.MoveMsg(key, Defs.STATUS_PENDING, Defs.STATUS_COMPLETED, "")
}
func (qdb *S_QDB) MarkFailed(key string) {
	//Get Message from Delayed or Pending and add to Failed
	//TODO IFs
	qdb.MoveMsg(key, Defs.STATUS_ACTIVE, Defs.STATUS_FAILED, "")
	qdb.MoveMsg(key, Defs.STATUS_PENDING, Defs.STATUS_FAILED, "")
}

func (qdb *S_QDB) CreateAndPushQMSG(topic string, message string) {
	qmsg := new(S_QMSG)
	qmsg.RawKey = DButils.GetNextKey(qdb.DB)
	qmsg.Key = (Defs.STATUS_PENDING + ":" + qmsg.RawKey)
	qmsg.Value = message
	qmsg.Status = Defs.STATUS_PENDING
	qmsg.Topic = topic
	qmsg.StatusHistory = make(map[string]time.Time)
	qmsg.StatusHistory[Defs.CREATED_AT] = time.Now()
	qdb.Push(*qmsg)
}

func (qdb *S_QDB) ClearComplete() {
	DButils.ClearPrefix(qdb.DB, Defs.STATUS_COMPLETED)
}

func (qdb *S_QDB) ClearFailed() {
	DButils.ClearPrefix(qdb.DB, Defs.STATUS_FAILED)
}

func (qdb *S_QDB) GetAllFailed() []S_QMSG {
	msgList := []S_QMSG{}

	b := DButils.GetAllPrefix(qdb.DB, Defs.STATUS_FAILED)
	for _, s := range b {
		newMSG := S_QMSG{}
		newMSG.Deserialize(s)
		msgList = append(msgList, newMSG)
	}
	return msgList
}

func (qdb *S_QDB) RetryAllFailed() {
	msgs := qdb.GetAllFailed()
	for _, m := range msgs {
		qdb.RetryFailed(m.Key)
	}
}
