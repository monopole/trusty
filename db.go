package hostlist

import (
	"appengine"
	"appengine/datastore"
	"log"
	"time"
)

type HostRecord struct {
	Author   string
	Content  string
	Date     time.Time
	TimeNice string
	Id       int64
}

type entRecord struct {
	name  string
	kind  string
	idStr string
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

			// Changing 'idStr' breaks lookups
			idStr: "idOfTheSingleRecord",
		},
		sortField: "-Date",
		maxRows:   10,
	}
}

// Key of parent entity of all entities in data store.
// One parent used to assure consistency.
func (db *Db) parentKey() *datastore.Key {
	return datastore.NewKey(db.ctx, db.entity.kind,
		db.entity.idStr, 0 /* no int Id */, nil /* no ancestor */)
}

// Set the same parent key on every HostRecord entity to ensure each
// HostRecord is in the same entity group. Queries across the single
// entity group will be consistent. However, the write rate to a
// single entity group should be limited to ~1/second.
func (db *Db) Write(g *HostRecord) (err error) {
	parentKey := db.parentKey()
	key := datastore.NewIncompleteKey(db.ctx, db.entity.name, parentKey)
	if chatty {
		log.Printf(" ")
		log.Printf(" ")
		log.Printf("Writing with parentKey = %v", parentKey)
		log.Printf("              item key = %v", key)
	}
	_, err = datastore.Put(db.ctx, key, g)
	return
}

func (db *Db) Delete(id int64) (err error) {
	key := datastore.NewKey(db.ctx, db.entity.name, "", id, db.parentKey())
	if chatty {
		log.Printf("Delete with id = %d", id)
		log.Printf("           key = %v", key)
	}
	err = datastore.Delete(db.ctx, key)
	return
}

// Ancestor queries like the one below are strongly consistent with
// the High Replication Datastore. Queries that span entity groups are
// eventually consistent. If .Ancestor were omitted from this query
// there would be a slight chance that freshly written data would not
// show up in a query.
func (db *Db) Read() (result []HostRecord, err error) {
	key := db.parentKey()
	if chatty {
		log.Printf("Reading with parentKey = %v", key)
	}
	q := datastore.NewQuery(db.entity.name).Ancestor(key).Order(
		db.sortField).Limit(db.maxRows)
	result = make([]HostRecord, 0, db.maxRows)
	keys, err := q.GetAll(db.ctx, &result)
	for i, k := range keys {
		// Grab the ID so deletion can be offered.
		result[i].Id = k.IntID()
	}
	for i, k := range result {
		//		result[i].TimeNice = k.Date.Format(time.RFC850)
		result[i].TimeNice = k.Date.Format(time.RFC850)
	}
	return
}
