package imdb

import (
	"reflect"
)

type value interface {
	value() interface{}
}

type boundValue struct {
	val interface{}
}

type unboundValue struct {
	valPtr interface{}
}

func (v *boundValue) value() interface{} {
	return v.val
}

func (v *unboundValue) value() interface{} {
	return reflect.ValueOf(v.valPtr).Elem().Interface()
}
