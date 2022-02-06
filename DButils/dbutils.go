package DButils

import (
	"os"
	"ywadi/crimsonq/Defs"
	"ywadi/crimsonq/Utils"

	log "github.com/sirupsen/logrus"

	badger "github.com/dgraph-io/badger/v3"
	"github.com/spf13/viper"
)

func CreateDb(dbname string, dbpath string) (database *badger.DB, err error) {
	db, err := badger.Open(badger.DefaultOptions(dbpath + "/" + dbname).WithLoggingLevel(badger.ERROR).WithMetricsEnabled(false).WithSyncWrites(viper.GetBool("crimson_settings.db_full_persist")).WithDetectConflicts(viper.GetBool("crimson_settings.db_detect_conflicts")))
	if err != nil {
		log.Fatal(err)
		return nil, err
	}
	return db, nil
}

func DestroyDb(dbname string, dbpath string) {
	os.RemoveAll(dbpath + "/" + dbname)
}

func GetNextKey(qdb *badger.DB) string {
	seq, err := qdb.GetSequence([]byte("0"), 10000)
	if err != nil {
		log.Fatal(err)
	}
	num, err := seq.Next()
	if err != nil {
		log.Fatal(err)

	}
	defer seq.Release()
	key := Utils.LexiPack(num)
	return key
}

func SET(db *badger.DB, key string, value []byte) error {
	keyString := []byte(key)
	err := db.Update(func(txn *badger.Txn) error {
		err := txn.Set(keyString, value)
		return err
	})
	return err
}

func GET(db *badger.DB, key string) ([]byte, error) {
	var result []byte
	err := db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(key))
		if err != nil {
			return nil
		}
		result, err = item.ValueCopy(nil)
		if err != nil {
			return nil
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

func DEL(db *badger.DB, key string) error {
	err := db.Update(func(txn *badger.Txn) error {
		err := txn.Delete([]byte(key))
		return err
	})
	return err
}

//Move to Pull ?
func DEQ(db *badger.DB) (key []byte, val []byte, xe error) {
	var result []byte
	var resultKey []byte
	err := db.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()
		prefix := []byte(Defs.STATUS_PENDING)
		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			item := it.Item()
			resultKey = item.Key()
			err := item.Value(func(v []byte) error {
				valCopy, err := item.ValueCopy(nil)
				if err != nil {
					return err
				}
				result = valCopy
				return err
			})
			if err != nil {
				return err
			}
			break
		}
		return nil
	})
	if err != nil {
		return nil, nil, err
	}
	return resultKey, result, nil
}

func ClearDb(db *badger.DB) {
	db.DropAll()
}

func ClearPrefix(db *badger.DB, prefix string) {
	db.DropPrefix([]byte(prefix))
}

func GetAllPrefix(db *badger.DB, filterPrefix string) [][]byte {
	var results [][]byte
	db.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()
		prefix := []byte(filterPrefix)
		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			item := it.Item()
			err := item.Value(func(v []byte) error {
				valCopy, err := item.ValueCopy(nil)
				if err != nil {
					return err
				}
				results = append(results, valCopy)
				return nil
			})
			if err != nil {
				return err
			}
		}
		return nil
	})
	return results
}
