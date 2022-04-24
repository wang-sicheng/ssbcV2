package levelDB

import (
	"github.com/cloudflare/cfssl/log"
	"github.com/syndtr/goleveldb/leveldb"
)

var db *leveldb.DB
var err error

func InitDB(path string) {
	db, err = leveldb.OpenFile("levelDB/db/path/"+path, nil)
	if err != nil {
		log.Errorf("db init %s err: %v", path, err)
	}
}

func DBGet(key string) []byte {
	data, err := db.Get([]byte(key), nil)
	if err != nil {
		log.Errorf("db get %s err: %v", key, err)
		return nil
	}
	return data
}

func DBPut(key string, value []byte) {
	err = db.Put([]byte(key), value, nil)
	if err != nil {
		log.Errorf("db put %s err: %v", key, err)
	}
}

func DBDelete(key string) {
	err = db.Delete([]byte(key), nil)
	if err != nil {
		log.Errorf("db delete %s err: %v", key, err)
	}
}
