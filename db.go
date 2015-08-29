package hostlist

import (
	"appengine"
	"appengine/datastore"
	"time"
)

type HostRecord struct {
	Author  string
	Content string
	Date    time.Time
}

type entRecord struct {
	kind       string
	idStr      string
	noIdInt    int64
	noAncestor *datastore.Key
}

type Db struct {
	ctx       appengine.Context
	sortField string
	table     string
	maxRows   int
	entity    entRecord
}

func NewDbHandle(ctx appengine.Context) *Db {
	return &Db{
		ctx:   ctx,
		table: "Hosts",
		entity: entRecord{
			kind:       "hostRecord",
			idStr:      "idOfTheSingleRecord",
			noIdInt:    0,
			noAncestor: nil,
		},
		sortField: "-Date",
		maxRows:   10,
	}
}

// Set the same parent key on every HostRecord entity to ensure each
// HostRecord is in the same entity group. Queries across the single
// entity group will be consistent. However, the write rate to a
// single entity group should be limited to ~1/second.
func (db *Db) Write(g *HostRecord) (err error) {
	key := datastore.NewIncompleteKey(db.ctx, db.table, db.key())
	_, err = datastore.Put(db.ctx, key, g)
	return
}

// Key used to recover all data.
func (db *Db) key() *datastore.Key {
	return datastore.NewKey(db.ctx, db.entity.kind,
		db.entity.idStr, db.entity.noIdInt, db.entity.noAncestor)
}

// Ancestor queries like that below are strongly consistent with the
// High Replication Datastore. Queries that span entity groups are
// eventually consistent. If we omitted the .Ancestor from this
// query there would be a slight chance that freshly written data
// would not show up in a query.
func (db *Db) Read() (result []HostRecord, err error) {
	q := datastore.NewQuery(db.table).Ancestor(
		db.key()).Order(db.sortField).Limit(db.maxRows)
	result = make([]HostRecord, 0, db.maxRows)
	_, err = q.GetAll(db.ctx, &result)
	return
}
