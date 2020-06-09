package imdb

type index interface {
	add(rec interface{}) error
	remove(rec interface{}) error
	match(q query) (score int)
	build(q query) (indexed map[string]value, nonIndexed map[string]value)
	query(q query) (recs []interface{})
}