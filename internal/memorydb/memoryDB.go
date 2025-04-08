package memorydb

import (
	"github.com/Jashanpreet2/fragments/internal/logger"
)

type LocalDB struct {
	db_map map[string]map[string]any
}

func (db *LocalDB) Put(pk string, sk string, val any) bool {
	if db.db_map == nil {
		db.db_map = map[string]map[string]any{}
	}
	if db.db_map[pk] == nil {
		db.db_map[pk] = map[string]any{}
	}
	db.db_map[pk][sk] = val
	return true
}

func (db *LocalDB) GetSKs(pk string) []string {
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

func (db *LocalDB) GetValue(pk string, sk string) (any, bool) {
	logger.Sugar.Info("Finding a value in the following map:")
	logger.Sugar.Info(db.db_map)
	if db.db_map[pk] != nil && db.db_map[pk][sk] != nil {
		return db.db_map[pk][sk], true
	}
	return nil, false
}

func (db *LocalDB) DeleteValue(pk string, sk string) bool {
	if db.db_map[pk] != nil && db.db_map[pk][sk] != nil {
		delete(db.db_map[pk], sk)
		return true
	}
	return false
}
