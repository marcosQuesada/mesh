package persistence

import (
	"github.com/boltdb/bolt"
	"github.com/marcosQuesada/mesh/pkg/cluster/command"
	"log"
)

type boltDB struct {
	database   *bolt.DB
	bucketName []byte
}

func NewBoltDb(db, b string) *boltDB {
	return &boltDB{
		database:   initializeDb(db),
		bucketName: []byte(b),
	}
}

func (s *boltDB) Set(args ...command.Args) error {
	key := []byte(args[0][0].(string))
	value := []byte(args[0][1].(string))

	err := s.database.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists(s.bucketName)
		if err != nil {
			return err
		}

		err = bucket.Put(key, value)
		if err != nil {
			return err
		}
		return nil
	})

	return err
}

func (s *boltDB) Get(args ...command.Args) (command.Response, error) {
	var value string

	key := []byte(args[0][0].(string))

	err := s.database.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(s.bucketName)
		value = string(bucket.Get(key))

		return nil
	})

	return value, err
}

func initializeDb(databaseName string) *bolt.DB {
	db, err := bolt.Open(databaseName, 0600, nil)

	if err != nil {
		log.Fatal(err)
	}

	return db
}
