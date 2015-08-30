package hostlist

import (
	"appengine"
	"appengine/datastore"
	"log"
	"time"
)

type HostRecord struct {
	Author  string
	Content string
	Date    time.Time
	Id      int64
}

type entRecord struct {
	name       string
	kind       string
	idStr      string
	noIdInt    int64
	noAncestor *datastore.Key
}

type Db struct {
	ctx       appengine.Context
	sortField string
	maxRows   int
	entity    entRecord
}

func NewDbHandle(ctx appengine.Context) *Db {
	return &Db{
		ctx: ctx,
		entity: entRecord{
			// Changing 'name' breaks lookups
			name: "Hosts",

			// Changing 'kind' breaks lookups.  However, data still goes in
			// and out, and can be interpreted by the app engine tool to
			// scan the DB, so reflection must be used.
			kind: "hostRecord",

			// Changing 'idSt' breaks lookups
			idStr: "idOfTheSingleRecord",

			// Intentionally blank, so idStr is used.
			noIdInt: 0,

			// Intentionally nil, so we always change the 'root' entity.
			noAncestor: nil,
		},
		sortField: "-Date",
		maxRows:   10,
	}
}

// Key used to recover all data.
func (db *Db) parentKey() *datastore.Key {
	return datastore.NewKey(db.ctx, db.entity.kind,
		db.entity.idStr, db.entity.noIdInt, db.entity.noAncestor)
}

// Set the same parent key on every HostRecord entity to ensure each
// HostRecord is in the same entity group. Queries across the single
// entity group will be consistent. However, the write rate to a
// single entity group should be limited to ~1/second.
func (db *Db) Write(g *HostRecord) (err error) {
	origKey := db.parentKey()
	log.Printf(" ")
	log.Printf(" ")
	var key *datastore.Key
	log.Printf("Writing with parentKey = %v", origKey)
	key = datastore.NewIncompleteKey(db.ctx, db.entity.name, origKey)
	log.Printf("Writing with key = %v", key)
	_, err = datastore.Put(db.ctx, key, g)
	return
}

func (db *Db) Delete(id int64) (err error) {
	log.Printf("Delete with id = %d", id)
	key := datastore.NewKey(db.ctx, db.entity.name, "", id, db.parentKey())
	log.Printf("Delete with key = %v", key)
	err = datastore.Delete(db.ctx, key)
	return
}

// Ancestor queries like that below are strongly consistent with the
// High Replication Datastore. Queries that span entity groups are
// eventually consistent. If we omitted the .Ancestor from this
// query there would be a slight chance that freshly written data
// would not show up in a query.
func (db *Db) Read() (result []HostRecord, err error) {
	key := db.parentKey()
	log.Printf("Reading with key = %v", key)
	q := datastore.NewQuery(db.entity.name).Ancestor(key).Order(
		db.sortField).Limit(db.maxRows)
	result = make([]HostRecord, 0, db.maxRows)
	keys, err := q.GetAll(db.ctx, &result)
	for i, k := range keys {
		result[i].Id = k.IntID()
	}
	return
}
