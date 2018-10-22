package db

import (
	"github.com/dgraph-io/badger"
	"log"
)

//Store
type Store interface {
	Get(key string) (string, error)
	Set(Key, v string) error

	//DB shuold be closed when exit
	Close()
}

//StoreDb
type StoreDb struct {
	db *badger.DB
}

//NewDb
func NewDb() Store {
	opts := badger.DefaultOptions
	opts.Dir = "/tmp/badger"
	opts.ValueDir = "/tmp/badger"
	db, err := badger.Open(opts)
	if err != nil {
		log.Fatal(err)
	}
	return &StoreDb{db: db}
}

//Get only string
func (s *StoreDb) Get(key string) (string, error) {
	var res []byte
	err := s.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(key))
		if err != nil {
			return err
		}
		item.Value(func(val []byte) error {
			res = append([]byte{}, val...)
			return nil
		})
		return nil
	})
	return string(res), err
}

//Set only string data type
func (s *StoreDb) Set(key, v string) error {
	return s.db.Update(func(txn *badger.Txn) error {
		return txn.Set([]byte(key), []byte(v))
	})
}

//Close
func (s *StoreDb) Close() {
	s.db.Close()
}
