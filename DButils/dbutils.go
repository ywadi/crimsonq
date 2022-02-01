package DButils

import (
	"log"
	"ywadi/goq/Utils"

	badger "github.com/dgraph-io/badger/v3"
)

func CreateDb(dbname string, dbpath string) (database *badger.DB, err error) {
	db, err := badger.Open(badger.DefaultOptions(dbpath + "/" + dbname))
	if err != nil {
		log.Fatal(err)
		return nil, err
	}
	return db, nil
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
		handle(err)
		return err
	})
	handle(err)
	return err
}

func GET(db *badger.DB, key string) []byte {
	var result []byte
	err := db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(key))
		handle(err)
		result, err = item.ValueCopy(nil)
		handle(err)
		return nil
	})
	handle(err)
	return result
}

func DEL(db *badger.DB, key string) error {
	err := db.Update(func(txn *badger.Txn) error {
		err := txn.Delete([]byte(key))
		handle(err)
		return err
	})
	handle(err)
	return err
}

//Move to Pull ?
func DEQ(db *badger.DB) (key []byte, val []byte) {
	var result []byte
	var resultKey []byte
	db.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()
		prefix := []byte("pending")
		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			item := it.Item()
			resultKey = item.Key()
			err := item.Value(func(v []byte) error {
				valCopy, err := item.ValueCopy(nil)
				if err != nil {
					return err
				}
				result = valCopy
				return nil
			})
			if err != nil {
				return err
			}
			break //Just get 1 from itterator, Defer will kill it
		}
		return nil
	})
	return resultKey, result
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

func handle(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
