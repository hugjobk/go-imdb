package imdb

import (
	"fmt"
	"reflect"
)

type uniqueIndex struct {
	keys *set
	recs map[string]interface{}
}

func (idx *uniqueIndex) add(rec interface{}) error {
	key, err := idx.buildKey(rec)
	if err != nil {
		return err
	}
	if _, ok := idx.recs[key]; ok {
		return fmt.Errorf("uniqueIndex%s violated", idx.keys)
	}
	idx.recs[key] = rec
	return nil
}

func (idx *uniqueIndex) remove(rec interface{}) error {
	key, err := idx.buildKey(rec)
	if err != nil {
		return err
	}
	delete(idx.recs, key)
	return nil
}

func (idx *uniqueIndex) match(q query) int {
	for key := range idx.keys.elements {
		if _, ok := q.filters[key.(string)]; !ok {
			return noScore
		}
	}
	return maxScore
}

func (idx *uniqueIndex) build(q query) (map[string]value, map[string]value) {
	if len(q.filters) == idx.keys.Len() {
		return q.filters, nil
	}
	indexed := make(map[string]value, idx.keys.Len())
	nonIndexed := make(map[string]value, len(q.filters)-idx.keys.Len())
	for key, val := range q.filters {
		if idx.keys.Has(key) {
			indexed[key] = val
		} else {
			nonIndexed[key] = val
		}
	}
	return indexed, nonIndexed
}

func (idx *uniqueIndex) query(q query) []interface{} {
	key := buildKey(q.indexed)
	rec, ok := idx.recs[key]
	if !ok || (len(q.nonIndexed) > 0 && match(rec, q.nonIndexed)) {
		return nil
	}
	return []interface{}{rec}
}

func (idx *uniqueIndex) buildKey(rec interface{}) (string, error) {
	v := reflect.ValueOf(rec)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return "", fmt.Errorf("invalid type of record, got %s expected struct", v.Kind())
	}
	t := v.Type()
	indexed := make(map[string]value)
	for i := 0; i < t.NumField(); i++ {
		sf := t.Field(i)
		if idx.keys.Has(sf.Name) {
			indexed[sf.Name] = &boundValue{v.Field(i).Interface()}
		}
	}
	key := buildKey(indexed)
	return key, nil
}
