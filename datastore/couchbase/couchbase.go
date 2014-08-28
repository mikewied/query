//  Copyright (c) 2014 Couchbase, Inc.
//  Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file
//  except in compliance with the License. You may obtain a copy of the License at
//    http://www.apache.org/licenses/LICENSE-2.0
//  Unless required by applicable law or agreed to in writing, software distributed under the
//  License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
//  either express or implied. See the License for the specific language governing permissions
//  and limitations under the License.

/*

Package file provides a couchbase-server implementation of the datasite
package.

*/

package couchbase

import (
	"fmt"
	"time"

	cb "github.com/couchbaselabs/go-couchbase"
	"github.com/couchbaselabs/query/accounting/logger_retriever"
	"github.com/couchbaselabs/query/datastore"
	"github.com/couchbaselabs/query/errors"
	"github.com/couchbaselabs/query/expression"
	"github.com/couchbaselabs/query/querylog"
	"github.com/couchbaselabs/query/value"
)

var lw *logger_retriever.RetrieverLogger

const (
	PRIMARY_INDEX = "#primary"
	ALLDOCS_INDEX = "#alldocs"
)

// datasite is the root for the couchbase datasite
type site struct {
	client         cb.Client             // instance of go-couchbase client
	namespaceCache map[string]*namespace // map of pool-names and IDs

}

func (s *site) Id() string {
	return s.URL()
}

func (s *site) URL() string {
	return s.client.BaseURL.String()
}

func (s *site) NamespaceIds() ([]string, errors.Error) {
	return s.NamespaceNames()
}

func (s *site) NamespaceNames() ([]string, errors.Error) {
	return []string{"default"}, nil
}

func (s *site) NamespaceById(id string) (p datastore.Namespace, e errors.Error) {
	return s.NamespaceByName(id)
}

func (s *site) NamespaceByName(name string) (p datastore.Namespace, e errors.Error) {
	p, ok := s.namespaceCache[name]
	if !ok {
		var err errors.Error
		p, err = loadNamespace(s, name)
		if err != nil {
			return nil, err
		}
		s.namespaceCache[name] = p.(*namespace)
	}
	return p, nil
}

// NewSite creates a new Couchbase site for the given url.
func NewDatastore(url string) (s datastore.Datastore, e errors.Error) {

	lw = logger_retriever.NewRetrieverLogger(nil)
	if lw == nil {
		fmt.Printf("Unable to initialize logger")
		return nil, errors.NewError(nil, "Unable to initialize logger")
	}

	client, err := cb.Connect(url)
	if err != nil {
		return nil, errors.NewError(err, "Cannot connect to url "+url)
	}

	site := &site{
		client:         client,
		namespaceCache: make(map[string]*namespace),
	}

	// initialize the default pool.
	// TODO can couchbase server contain more than one pool ?

	defaultPool, Err := loadNamespace(site, "default")
	if Err != nil {
		lw.Error("", querylog.DATASTORE, "Cannot connect to default pool")
		return nil, Err
	}

	site.namespaceCache["default"] = defaultPool
	lw.Info("", querylog.DATASTORE, "New site created with url %s", url)

	return site, nil

}

func loadNamespace(s *site, name string) (*namespace, errors.Error) {

	cbpool, err := s.client.GetPool(name)
	if err != nil {
		if name == "default" {
			// if default pool is not available, try reconnecting to the server
			url := s.URL()
			client, err := cb.Connect(url)
			if err != nil {
				return nil, errors.NewError(nil, fmt.Sprintf("Pool %v not found.", name))
			}
			// check if the default pool exists
			cbpool, err = client.GetPool(name)
			if err != nil {
				return nil, errors.NewError(nil, fmt.Sprintf("Pool %v not found.", name))
			}
			s.client = client
		}
	}

	rv := namespace{
		site:          s,
		name:          name,
		cbNamespace:   cbpool,
		keyspaceCache: make(map[string]datastore.Keyspace),
	}
	go keepPoolFresh(&rv)
	return &rv, nil
}

// a namespace represents a couchbase pool
type namespace struct {
	site          *site
	name          string
	cbNamespace   cb.Pool
	keyspaceCache map[string]datastore.Keyspace
}

func (p *namespace) DatastoreId() string {
	return p.site.Id()
}

func (p *namespace) Id() string {
	return p.Name()
}

func (p *namespace) Name() string {
	return p.name
}

func (p *namespace) KeyspaceIds() ([]string, errors.Error) {
	return p.KeyspaceNames()
}

func (p *namespace) KeyspaceNames() ([]string, errors.Error) {
	rv := make([]string, 0, len(p.cbNamespace.BucketMap))
	for name, _ := range p.cbNamespace.BucketMap {
		rv = append(rv, name)
	}
	return rv, nil
}

func (p *namespace) KeyspaceByName(name string) (b datastore.Keyspace, e errors.Error) {

	b, ok := p.keyspaceCache[name]
	if !ok {
		var err errors.Error
		b, err = newKeyspace(p, name)
		if err != nil {
			return nil, errors.NewError(err, "Keyspace "+name+" name not found")
		}
		p.keyspaceCache[name] = b
	}
	return b, nil
}

func (p *namespace) KeyspaceById(id string) (datastore.Keyspace, errors.Error) {
	return p.KeyspaceByName(id)
}

func (p *namespace) refresh() {
	// trigger refresh of this pool
	lw.Info("", querylog.DATASTORE, "Refreshing pool %s", p.name)

	newpool, err := p.site.client.GetPool(p.name)
	if err != nil {
		lw.Error("", querylog.DATASTORE, "Error updating pool name %s: Error %v", p.name, err)
		url := p.site.URL()
		client, err := cb.Connect(url)
		if err != nil {
			lw.Error("", querylog.DATASTORE, "Error connecting to URL %s", url)
			return
		}
		// check if the default pool exists
		newpool, err = client.GetPool(p.name)
		if err != nil {
			lw.Error("", querylog.DATASTORE, "Retry Failed Error updating pool name %s: Error %v", p.name, err)
			return
		}
		p.site.client = client

	}
	p.cbNamespace = newpool
}

func keepPoolFresh(p *namespace) {

	tickChan := time.Tick(1 * time.Minute)

	for _ = range tickChan {
		p.refresh()
	}
}

type keyspace struct {
	namespace *namespace
	name      string
	indexes   map[string]datastore.Index
	primary   datastore.PrimaryIndex
	cbbucket  *cb.Bucket
}

func newKeyspace(p *namespace, name string) (datastore.Keyspace, errors.Error) {

	lw.Info("", querylog.DATASTORE, "Created New Bucket %s", name)
	cbbucket, err := p.cbNamespace.GetBucket(name)
	if err != nil {
		// go-couchbase caches the buckets
		// to be sure no such bucket exists right now
		// we trigger a refresh
		p.refresh()
		// and then check one more time
		cbbucket, err = p.cbNamespace.GetBucket(name)
		if err != nil {
			// really no such bucket exists
			return nil, errors.NewError(nil, fmt.Sprintf("Bucket %v not found.", name))
		}
	}

	rv := &keyspace{
		namespace: p,
		name:      name,
		cbbucket:  cbbucket,
		indexes:   make(map[string]datastore.Index),
	}

	//discover existing indexes TODO
	ierr := rv.loadIndexes()
	if ierr != nil {
		lw.Info("", querylog.DATASTORE, "Error loading index %s", ierr.Error())
		return nil, ierr
	}

	return rv, nil
}

func (b *keyspace) NamespaceId() string {
	return b.namespace.Id()
}

func (b *keyspace) Id() string {
	return b.Name()
}

func (b *keyspace) Name() string {
	return b.name
}

func (b *keyspace) Count() (int64, errors.Error) {
	// need equivalent of all_docs. TODO with view engine supporty
	pi, err := b.IndexByPrimary()
	if err != nil || pi == nil {
		return 0, errors.NewError(nil, "No primary index found. Please create a primary index on bucket "+b.Name())
	}

	var vi *viewIndex
	vi = pi.(*viewIndex)
	totalCount, err := ViewTotalRows(vi.keyspace.cbbucket, vi.DDocName(), vi.ViewName(), map[string]interface{}{})
	if err != nil {
		return 0, err
	}

	return totalCount, nil
}

func (b *keyspace) IndexIds() ([]string, errors.Error) {
	rv := make([]string, 0, len(b.indexes))
	for name, _ := range b.indexes {
		rv = append(rv, name)
	}
	return rv, nil
}

func (b *keyspace) IndexNames() ([]string, errors.Error) {
	rv := make([]string, 0, len(b.indexes))
	for name, _ := range b.indexes {
		rv = append(rv, name)
	}
	return rv, nil
}

func (b *keyspace) IndexById(id string) (datastore.Index, errors.Error) {
	return b.IndexByName(id)
}

func (b *keyspace) IndexByName(name string) (datastore.Index, errors.Error) {
	index, ok := b.indexes[name]
	if !ok {
		return nil, errors.NewError(nil, fmt.Sprintf("Index %v not found.", name))
	}
	return index, nil
}

func (b *keyspace) IndexByPrimary() (datastore.PrimaryIndex, errors.Error) {

	if b.primary == nil {
		idx, ok := b.indexes[PRIMARY_INDEX]
		if ok {
			primary := idx.(datastore.PrimaryIndex)
			return primary, nil
		}
		all, ok := b.indexes[ALLDOCS_INDEX]
		if ok {
			primary := all.(datastore.PrimaryIndex)
			return primary, nil
		}
	}
	return b.primary, nil
}

func (b *keyspace) Indexes() ([]datastore.Index, errors.Error) {
	rv := make([]datastore.Index, 0, len(b.indexes))
	for _, index := range b.indexes {
		rv = append(rv, index)
	}
	return rv, nil
}

func (b *keyspace) CreatePrimaryIndex() (datastore.PrimaryIndex, errors.Error) {
	if _, exists := b.indexes[PRIMARY_INDEX]; exists {
		return nil, errors.NewError(nil, "Primary index already exists")
	}
	idx, err := newPrimaryIndex(b)
	if err != nil {
		return nil, errors.NewError(err, "Error creating primary index")
	}
	b.indexes[idx.Name()] = idx
	return idx, nil
}

func (b *keyspace) CreateIndex(name string, equalKey, rangeKey expression.Expressions, using datastore.IndexType) (datastore.Index, errors.Error) {

	if using == "" {
		// current default is VIEW
		using = datastore.VIEW
	}

	switch using {
	case datastore.VIEW:
		if _, exists := b.indexes[name]; exists {
			return nil, errors.NewError(nil, fmt.Sprintf("Index already exists: %s", name))
		}
		idx, err := newViewIndex(name, datastore.IndexKey(equalKey), b)
		if err != nil {
			return nil, errors.NewError(err, fmt.Sprintf("Error creating index: %s", name))
		}
		b.indexes[idx.Name()] = idx
		return idx, nil

	default:
		return nil, errors.NewError(nil, "Not yet implemented.")
	}
}

func (b *keyspace) Fetch(keys []string) ([]datastore.Pair, errors.Error) {

	if len(keys) == 0 {
		return nil, errors.NewError(nil, "No keys to fetch")
	}

	rv := make([]datastore.Pair, len(keys))
	bulkResponse, err := b.cbbucket.GetBulk(keys)
	if err != nil {
		return nil, errors.NewError(err, "Error doing bulk get")
	}

	i := 0
	for k, v := range bulkResponse {

		var doc datastore.Pair
		doc.Key = k

		Value := value.NewAnnotatedValue(value.NewValueFromBytes(v.Body))

		meta_flags := (v.Extras[0]&0xff)<<24 | (v.Extras[1]&0xff)<<16 | (v.Extras[2]&0xff)<<8 | (v.Extras[3] & 0xff)
		meta_type := "json"
		if Value.Type() == value.NOT_JSON {
			meta_type = "base64"
		}
		Value.SetAttachment("meta", map[string]interface{}{
			"id":    k,
			"cas":   float64(v.Cas),
			"type":  meta_type,
			"flags": float64(meta_flags),
		})
		doc.Value = Value
		rv[i] = doc
		i++

	}

	lw.Debug("", querylog.DATASTORE, "Fetched %d keys ", i)

	return rv, nil
}

func (b *keyspace) FetchOne(key string) (value.Value, errors.Error) {

	item, e := b.Fetch([]string{key})
	if e != nil {
		return nil, e
	}
	// not found
	if len(item) == 0 {
		return nil, nil
	}

	return item[0].Value, e
}

const (
	INSERT = 0x01
	UPDATE = 0x02
	UPSERT = 0x04
)

func opToString(op int) string {

	switch op {
	case INSERT:
		return "insert"
	case UPDATE:
		return "update"
	case UPSERT:
		return "upsert"
	}

	return "unknown operation"
}

func (b *keyspace) performOp(op int, inserts []datastore.Pair) ([]datastore.Pair, errors.Error) {

	if len(inserts) == 0 {
		return nil, errors.NewError(nil, "No keys to insert")
	}

	insertedKeys := make([]datastore.Pair, 0)
	var err error

	for _, kv := range inserts {
		key := kv.Key
		value := kv.Value.Actual()

		// TODO Need to also set meta
		switch op {

		case INSERT:
			// add the key to the backend
			_, err = b.cbbucket.Add(key, 0, value)
		case UPDATE:
			// check if the key exists and if so then use the cas value
			// to update the key
			rv := make([]string, 0, 10)
			var cas uint64

			err = b.cbbucket.Gets(key, &rv, &cas)
			if err == nil {
				err = b.cbbucket.Set(key, 0, value)
			} else {
				lw.Error("", querylog.DATASTORE, "Failed to insert. Key exists %s", key)
			}
		case UPSERT:
			err = b.cbbucket.Set(key, 0, value)
		}

		if err != nil {
			lw.Error("", querylog.DATASTORE, "Failed to perform %s on key %s Error %v", opToString(op), key, err)
		} else {
			insertedKeys = append(insertedKeys, kv)
		}
	}

	if len(insertedKeys) == 0 {
		return nil, errors.NewError(err, "Failed to perform "+opToString(op))
	}

	return insertedKeys, nil

}

func (b *keyspace) Insert(inserts []datastore.Pair) ([]datastore.Pair, errors.Error) {
	return b.performOp(INSERT, inserts)

}

func (b *keyspace) Update(updates []datastore.Pair) ([]datastore.Pair, errors.Error) {
	return b.performOp(UPDATE, updates)
}

func (b *keyspace) Upsert(upserts []datastore.Pair) ([]datastore.Pair, errors.Error) {
	return b.performOp(UPSERT, upserts)
}

func (b *keyspace) Delete(deletes []string) errors.Error {

	failedDeletes := make([]string, 0)
	var err error
	for _, key := range deletes {
		if err = b.cbbucket.Delete(key); err != nil {
			lw.Info("", querylog.DATASTORE, "Failed to delete key %s", key)
			failedDeletes = append(failedDeletes, key)
		}
	}

	if len(failedDeletes) > 0 {
		return errors.NewError(err, "Some keys were not deleted "+fmt.Sprintf("%v", failedDeletes))
	}

	return nil
}

func (b *keyspace) Release() {
	b.cbbucket.Close()
}

// primaryIndex performs full keyspace scans.
type primaryIndex struct {
	viewIndex
}

func (pi *primaryIndex) KeyspaceId() string {
	return pi.keyspace.Id()
}

func (pi *primaryIndex) Id() string {
	return pi.Name()
}

func (pi *primaryIndex) Name() string {
	return pi.name
}

func (pi *primaryIndex) Type() datastore.IndexType {
	return pi.Type()
}

func (pi *primaryIndex) Drop() errors.Error {
	return errors.NewError(nil, "This primary index cannot be dropped.")
}

func (pi *primaryIndex) EqualKey() expression.Expressions {
	return nil
}

func (pi *primaryIndex) RangeKey() expression.Expressions {
	// FIXME
	return nil
}

func (pi *primaryIndex) Condition() expression.Expression {
	return nil
}

func (pi *primaryIndex) Statistics(span *datastore.Span) (datastore.Statistics, errors.Error) {
	return nil, nil
}

func (pi *primaryIndex) Scan(span *datastore.Span, distinct bool, limit int64, conn *datastore.IndexConnection) {
	pi.viewIndex.Scan(span, distinct, limit, conn)

}

func (pi *primaryIndex) ScanEntries(limit int64, conn *datastore.IndexConnection) {
	pi.viewIndex.ScanEntries(limit, conn)

}