package main

import (
	"io"
	"mime/multipart"
)

type localdb struct {
	db_map map[string]map[string]any
}

func (db *localdb) Put(pk string, sk string, val any) bool {
	if db.db_map == nil {
		db.db_map = map[string]map[string]any{}
	}
	if db.db_map[pk] == nil {
		db.db_map[pk] = map[string]any{}
	}
	db.db_map[pk][sk] = val
	return true
}

func (db *localdb) GetSKs(pk string) []string {
	if db.db_map[pk] == nil {
		return []string{}
	}
	sks := make([]string, len(db.db_map[pk]))
	i := 0
	for k := range db.db_map[pk] {
		sks[i] = k
		i += 1
	}
	return sks
}

func (db *localdb) GetValue(pk string, sk string) (any, bool) {
	sugar.Info("Finding a value in the following map:")
	sugar.Info(db.db_map)
	if db.db_map[pk] != nil && db.db_map[pk][sk] != nil {
		// Delete
		sugar.Info("Fragment data stored in db: ", db.db_map[pk][sk])
		if file, ok := db.db_map[pk][sk].(multipart.File); ok {
			file.Seek(0, 0)
			data, _ := io.ReadAll(file)
			sugar.Info("Value in memoryDB: ", string(data))
		}
		//
		return db.db_map[pk][sk], true
	}
	return nil, false
}

func (db *localdb) DeleteValue(pk string, sk string) bool {
	if db.db_map[pk] != nil && db.db_map[pk][sk] != nil {
		delete(db.db_map[pk], sk)
		return true
	}
	return false
}
