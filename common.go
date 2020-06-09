package imdb

import (
	"fmt"
	"reflect"
	"sort"
	"strings"
)

const (
	noScore  = 0
	maxScore = int(^uint(0) >> 1)
)

func match(rec interface{}, m map[string]value) bool {
	v := reflect.ValueOf(rec)
	if v.Kind() != reflect.Struct {
		return false
	}
	for key, val := range m {
		f := v.FieldByName(key)
		if !f.IsValid() || !f.CanInterface() || f.Interface() != val.value() {
			return false
		}
	}
	return true
}

func buildKey(m map[string]value) string {
	arr := make([]string, 0, len(m))
	for key, val := range m {
		arr = append(arr, fmt.Sprintf("%s:%s", key, val.value()))
	}
	sort.Strings(arr)
	return strings.Join(arr, "|")
}
