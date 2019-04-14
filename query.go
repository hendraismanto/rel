package grimoire

import (
	"reflect"
	"strings"
	"time"

	"github.com/Fs02/grimoire/c"
	"github.com/Fs02/grimoire/changeset"
	"github.com/Fs02/grimoire/errors"
	"github.com/Fs02/grimoire/internal"
	"github.com/azer/snakecase"
)

// Query defines information about query generated by query builder.
type Query struct {
	repo            *Repo
	Collection      string
	Fields          []string
	AggregateField  string
	AggregateMode   string
	AsDistinct      bool
	JoinClause      []c.Join
	Condition       c.Condition
	GroupFields     []string
	HavingCondition c.Condition
	OrderClause     []c.Order
	OffsetResult    int
	LimitResult     int
	LockClause      string
	Changes         map[string]interface{}
}

// Select filter fields to be selected from database.
func (query Query) Select(fields ...string) Query {
	query.Fields = fields
	return query
}

// Distinct add distinct option to select query.
func (query Query) Distinct() Query {
	query.AsDistinct = true
	return query
}

// Join current collection with other collection.
func (query Query) Join(collection string, condition ...c.Condition) Query {
	return query.JoinWith("JOIN", collection, condition...)
}

// JoinWith current collection with other collection with custom join mode.
func (query Query) JoinWith(mode string, collection string, condition ...c.Condition) Query {
	if len(condition) == 0 {
		query.JoinClause = append(query.JoinClause, c.Join{
			Mode:       mode,
			Collection: collection,
			Condition: c.And(c.Eq(
				c.I(query.Collection+"."+strings.TrimSuffix(collection, "s")+"_id"),
				c.I(collection+".id"),
			)),
		})
	} else {
		query.JoinClause = append(query.JoinClause, c.Join{
			Mode:       mode,
			Collection: collection,
			Condition:  c.And(condition...),
		})
	}

	return query
}

// Where expressions are used to filter the result set. If there is more than one where expression, they are combined with an and operator.
func (query Query) Where(condition ...c.Condition) Query {
	query.Condition = query.Condition.And(condition...)
	return query
}

// OrWhere behaves exactly the same as where except it combines with any previous expression by using an OR.
func (query Query) OrWhere(condition ...c.Condition) Query {
	query.Condition = query.Condition.Or(c.And(condition...))
	return query
}

// Group query using fields.
func (query Query) Group(fields ...string) Query {
	query.GroupFields = fields
	return query
}

// Having adds condition for group query.
func (query Query) Having(condition ...c.Condition) Query {
	query.HavingCondition = query.HavingCondition.And(condition...)
	return query
}

// OrHaving behaves exactly the same as having except it combines with any previous expression by using an OR.
func (query Query) OrHaving(condition ...c.Condition) Query {
	query.HavingCondition = query.HavingCondition.Or(c.And(condition...))
	return query
}

// Order the result returned by database.
func (query Query) Order(order ...c.Order) Query {
	query.OrderClause = append(query.OrderClause, order...)
	return query
}

// Offset the result returned by database.
func (query Query) Offset(offset int) Query {
	query.OffsetResult = offset
	return query
}

// Limit result returned by database.
func (query Query) Limit(limit int) Query {
	query.LimitResult = limit
	return query
}

// Lock query using pessimistic locking.
// Lock expression can be specified as first parameter, default to FOR UPDATE.
func (query Query) Lock(lock ...string) Query {
	if !query.repo.inTransaction {
		return query
	}

	if len(lock) > 0 {
		query.LockClause = lock[0]
	} else {
		query.LockClause = "FOR UPDATE"
	}

	return query
}

// Find adds where id=? into query.
// This is short cut for Where(Eq(I("id"), 1))
func (query Query) Find(id interface{}) Query {
	return query.FindBy("id", id)
}

// FindBy adds where col=? into query.
func (query Query) FindBy(col string, val interface{}) Query {
	return query.Where(c.Eq(c.I(query.Collection+"."+col), val))
}

// Set value for insert or update operation that will replace changeset value.
func (query Query) Set(field string, value interface{}) Query {
	if query.Changes == nil {
		query.Changes = make(map[string]interface{})
	}

	query.Changes[field] = value
	return query
}

// One retrieves one result that match the query.
// // If no result found, it'll return not found error.
// func (query Query) One(record interface{}) error {
// 	query.LimitResult = 1
// 	count, err := query.repo.adapter.All(query, record, query.repo.logger...)

// 	if err != nil {
// 		return transformError(err)
// 	} else if count == 0 {
// 		return errors.New("no result found", "", errors.NotFound)
// 	} else {
// 		return nil
// 	}
// }

// MustOne retrieves one result that match the query.
// If no result found, it'll panic.
// func (query Query) MustOne(record interface{}) {
// 	must(query.One(record))
// }

// All retrieves all results that match the query.
// func (query Query) All(record interface{}) error {
// 	_, err := query.repo.adapter.All(query, record, query.repo.logger...)
// 	return err
// }

// MustAll retrieves all results that match the query.
// It'll panic if any error eccured.
// func (query Query) MustAll(record interface{}) {
// 	must(query.All(record))
// }

// Aggregate calculate aggregate over the given field.
// func (query Query) Aggregate(mode string, field string, out interface{}) error {
// 	query.AggregateMode = mode
// 	query.AggregateField = field
// 	return query.repo.adapter.Aggregate(query, out, query.repo.logger...)
// }

// // MustAggregate calculate aggregate over the given field.
// // It'll panic if any error eccured.
// func (query Query) MustAggregate(mode string, field string, out interface{}) {
// 	must(query.Aggregate(mode, field, out))
// }

// Count retrieves count of results that match the query.
// func (query Query) Count() (int, error) {
// 	var out struct {
// 		Count int
// 	}

// 	err := query.Aggregate("count", "*", &out)
// 	return out.Count, err
// }

// // MustCount retrieves count of results that match the query.
// // It'll panic if any error eccured.
// func (query Query) MustCount() int {
// 	count, err := query.Count()
// 	must(err)
// 	return count
// }

// // Insert records to database.
// func (query Query) Insert(record interface{}, chs ...*changeset.Changeset) error {
// 	var err error
// 	var ids []interface{}

// 	if len(chs) == 1 {
// 		// single insert
// 		ch := chs[0]
// 		changes := make(map[string]interface{})
// 		cloneChangeset(changes, ch.Changes())
// 		putTimestamp(changes, "created_at", ch.Types())
// 		putTimestamp(changes, "updated_at", ch.Types())
// 		cloneQuery(changes, query.Changes)

// 		var id interface{}
// 		id, err = query.repo.adapter.Insert(query, changes, query.repo.logger...)
// 		ids = append(ids, id)
// 	} else if len(chs) > 1 {
// 		// multiple insert
// 		fields := getFields(query, chs)

// 		allchanges := make([]map[string]interface{}, len(chs))
// 		for i, ch := range chs {
// 			changes := make(map[string]interface{})
// 			cloneChangeset(changes, ch.Changes())
// 			putTimestamp(changes, "created_at", ch.Types())
// 			putTimestamp(changes, "updated_at", ch.Types())
// 			cloneQuery(changes, query.Changes)

// 			allchanges[i] = changes
// 		}

// 		ids, err = query.repo.adapter.InsertAll(query, fields, allchanges, query.repo.logger...)
// 	} else if len(query.Changes) > 0 {
// 		// set only
// 		var id interface{}
// 		id, err = query.repo.adapter.Insert(query, query.Changes, query.repo.logger...)
// 		ids = append(ids, id)
// 	}

// 	if err != nil {
// 		return transformError(err, chs...)
// 	} else if record == nil || len(ids) == 0 {
// 		return nil
// 	} else if len(ids) == 1 {
// 		return transformError(query.Find(ids[0]).One(record))
// 	}

// 	return transformError(query.Where(c.In(c.I("id"), ids...)).All(record))
// }

// // MustInsert records to database.
// // It'll panic if any error occurred.
// func (query Query) MustInsert(record interface{}, chs ...*changeset.Changeset) {
// 	must(query.Insert(record, chs...))
// }

// Update records in database.
// It'll panic if any error occurred.
// func (query Query) Update(record interface{}, chs ...*changeset.Changeset) error {
// 	changes := make(map[string]interface{})

// 	// only take the first changeset if any
// 	if len(chs) != 0 {
// 		cloneChangeset(changes, chs[0].Changes())
// 		putTimestamp(changes, "updated_at", chs[0].Types())
// 	}

// 	cloneQuery(changes, query.Changes)

// 	// nothing to update
// 	if len(changes) == 0 {
// 		return nil
// 	}

// 	// perform update
// 	err := query.repo.adapter.Update(query, changes, query.repo.logger...)
// 	if err != nil {
// 		return transformError(err, chs...)
// 	}

// 	// should not fetch updated record(s) if not necessery
// 	if record != nil {
// 		return transformError(query.All(record))
// 	}

// 	return nil
// }

// // MustUpdate records in database.
// // It'll panic if any error occurred.
// func (query Query) MustUpdate(record interface{}, chs ...*changeset.Changeset) {
// 	must(query.Update(record, chs...))
// }

func cloneChangeset(out map[string]interface{}, changes map[string]interface{}) {
	for k, v := range changes {
		// skip if not scannable
		if v == nil || !internal.Scannable(reflect.TypeOf(v)) {
			continue
		}

		out[k] = v
	}
}

func cloneQuery(out map[string]interface{}, changes map[string]interface{}) {
	for k, v := range changes {
		out[k] = v
	}
}

func putTimestamp(out map[string]interface{}, field string, types map[string]reflect.Type) {
	if typ, ok := types[field]; ok && typ == reflect.TypeOf(time.Time{}) {
		out[field] = time.Now().Round(time.Second)
	}
}

func getFields(chs []*changeset.Changeset) []string {
	fields := make([]string, 0, len(chs[0].Types()))

	for f := range chs[0].Types() {
		if f == "created_at" || f == "updated_at" {
			fields = append(fields, f)
			continue
		}

		// if _, exist := query.Changes[f]; exist {
		// 	fields = append(fields, f)
		// 	continue
		// }

		for _, ch := range chs {
			if _, exist := ch.Changes()[f]; exist {
				// skip if not scannable
				if !internal.Scannable(ch.Types()[f]) {
					break
				}

				fields = append(fields, f)
				break
			}
		}
	}

	return fields
}

// Save a record to database.
// If condition exist, it will try to update the record, otherwise it'll insert it.
// Save ignores id from record.
// func (query Query) Save(record interface{}) error {
// 	rv := reflect.ValueOf(record)
// 	rt := rv.Type()
// 	if rt.Kind() == reflect.Ptr && rt.Elem().Kind() == reflect.Slice {
// 		// Put multiple records
// 		rv = rv.Elem()

// 		// if it's an empty slice, do nothing
// 		if rv.Len() == 0 {
// 			return nil
// 		}

// 		if query.Condition.None() {
// 			// InsertAll
// 			chs := []*changeset.Changeset{}

// 			for i := 0; i < rv.Len(); i++ {
// 				ch := changeset.Convert(rv.Index(i).Interface())
// 				changeset.DeleteChange(ch, "id")
// 				chs = append(chs, ch)
// 			}

// 			return query.Insert(record, chs...)
// 		}

// 		// Update only with first record definition.
// 		ch := changeset.Convert(rv.Index(0).Interface())
// 		changeset.DeleteChange(ch, "id")
// 		changeset.DeleteChange(ch, "created_at")
// 		return query.Update(record, ch)
// 	}

// 	// Put single records
// 	ch := changeset.Convert(record)
// 	changeset.DeleteChange(ch, "id")

// 	if query.Condition.None() {
// 		return query.Insert(record, ch)
// 	}

// 	// remove created_at from changeset
// 	changeset.DeleteChange(ch, "created_at")

// 	return query.Update(record, ch)
// }

// // MustSave puts a record to database.
// // It'll panic if any error eccured.
// func (query Query) MustSave(record interface{}) {
// 	must(query.Save(record))
// }

// Delete deletes all results that match the query.
// func (query Query) Delete() error {
// 	return transformError(query.repo.adapter.Delete(query, query.repo.logger...))
// }

// MustDelete deletes all results that match the query.
// It'll panic if any error eccured.
// func (query Query) MustDelete() {
// 	must(query.Delete())
// }

type preloadTarget struct {
	schema reflect.Value
	field  reflect.Value
}

// // Preload loads association with given query.
// func (query Query) Preload(record interface{}, field string) error {
// 	path := strings.Split(field, ".")

// 	rv := reflect.ValueOf(record)
// 	if rv.Kind() != reflect.Ptr || rv.IsNil() {
// 		panic("grimoire: record parameter must be a pointer.")
// 	}

// 	preload := traversePreloadTarget(rv.Elem(), path)
// 	if len(preload) == 0 {
// 		return nil
// 	}

// 	schemaType := preload[0].schema.Type()
// 	refIndex, fkIndex, column := getPreloadInfo(schemaType, path[len(path)-1])

// 	addrs, ids := collectPreloadTarget(preload, refIndex)
// 	if len(ids) == 0 {
// 		return nil
// 	}

// 	// prepare temp result variable for querying
// 	rt := preload[0].field.Type()
// 	if rt.Kind() == reflect.Slice || rt.Kind() == reflect.Array || rt.Kind() == reflect.Ptr {
// 		rt = rt.Elem()
// 	}

// 	slice := reflect.MakeSlice(reflect.SliceOf(rt), 0, len(ids))
// 	result := reflect.New(slice.Type())
// 	result.Elem().Set(slice)

// 	// query all records using collected ids.
// 	err := query.Where(c.In(c.I(column), ids...)).All(result.Interface())
// 	if err != nil {
// 		return err
// 	}

// 	// map results.
// 	result = result.Elem()
// 	for i := 0; i < result.Len(); i++ {
// 		curr := result.Index(i)
// 		id := getPreloadID(curr.FieldByIndex(fkIndex))

// 		for _, addr := range addrs[id] {
// 			if addr.Kind() == reflect.Slice {
// 				addr.Set(reflect.Append(addr, curr))
// 			} else if addr.Kind() == reflect.Ptr {
// 				currP := reflect.New(curr.Type())
// 				currP.Elem().Set(curr)
// 				addr.Set(currP)
// 			} else {
// 				addr.Set(curr)
// 			}
// 		}
// 	}

// 	return nil
// }

// // MustPreload loads association with given query.
// // It'll panic if any error occurred.
// func (query Query) MustPreload(record interface{}, field string) {
// 	must(query.Preload(record, field))
// }

func traversePreloadTarget(rv reflect.Value, path []string) []preloadTarget {
	result := []preloadTarget{}
	rt := rv.Type()

	if rt.Kind() == reflect.Slice || rt.Kind() == reflect.Array {
		for i := 0; i < rv.Len(); i++ {
			result = append(result, traversePreloadTarget(rv.Index(i), path)...)
		}
		return result
	}

	// forward to next path.
	fv := rv.FieldByName(path[0])
	if !fv.IsValid() || (fv.Kind() != reflect.Struct && fv.Kind() != reflect.Slice && fv.Kind() != reflect.Ptr) {
		panic("grimoire: field (" + path[0] + ") is not a struct, a slice or a pointer.")
	}

	if fv.Kind() == reflect.Ptr && len(path) != 1 {
		if fv.IsNil() {
			return result
		}

		fv = fv.Elem()
	}

	if len(path) == 1 {
		result = append(result, preloadTarget{
			schema: rv,
			field:  fv,
		})
	} else {
		result = append(result, traversePreloadTarget(fv, path[1:])...)
	}

	return result
}

func getPreloadInfo(rt reflect.Type, field string) ([]int, []int, string) {
	sft, _ := rt.FieldByName(field)
	ft := sft.Type

	ref := sft.Tag.Get("references")
	fk := sft.Tag.Get("foreign_key")
	column := ""

	if ft.Kind() == reflect.Ptr || ft.Kind() == reflect.Slice || ft.Kind() == reflect.Array {
		ft = ft.Elem()
	}

	// Try to guess ref and fk if not defined.
	if ref == "" || fk == "" {
		if _, isBelongsTo := rt.FieldByName(sft.Name + "ID"); isBelongsTo {
			ref = sft.Name + "ID"
			fk = "ID"
		} else {
			ref = "ID"
			fk = rt.Name() + "ID"
		}
	}

	var refIndex []int
	if idv, exist := rt.FieldByName(ref); !exist {
		panic("grimoire: references (" + ref + ") field not found ")
	} else {
		refIndex = idv.Index
	}

	var fkIndex []int
	if sf, exist := ft.FieldByName(fk); !exist {
		panic("grimoire: foreign_key (" + fk + ") field not found " + fk)
	} else {
		fkIndex = sf.Index

		if tag := sf.Tag.Get("db"); tag != "" {
			column = tag
		} else {
			column = snakecase.SnakeCase(fk)
		}
	}

	return refIndex, fkIndex, column
}

func collectPreloadTarget(preload []preloadTarget, refIndex []int) (map[interface{}][]reflect.Value, []interface{}) {
	addrs := make(map[interface{}][]reflect.Value)
	ids := []interface{}{}

	for i := range preload {
		refv := preload[i].schema.FieldByIndex(refIndex)
		fv := preload[i].field

		// Skip if nil
		if refv.Kind() == reflect.Ptr && refv.IsNil() {
			continue
		}

		id := getPreloadID(refv)

		// reset to zero if slice.
		if fv.Kind() == reflect.Slice || fv.Kind() == reflect.Array {
			fv.Set(reflect.Zero(fv.Type()))
		}

		addrs[id] = append(addrs[id], fv)

		// add to ids if not yet added.
		if len(addrs[id]) == 1 {
			ids = append(ids, id)
		}
	}

	return addrs, ids
}

func getPreloadID(fv reflect.Value) interface{} {
	if fv.Kind() == reflect.Ptr {
		return fv.Elem().Interface()
	}

	return fv.Interface()
}

func transformError(err error, chs ...*changeset.Changeset) error {
	if err == nil {
		return nil
	} else if e, ok := err.(errors.Error); ok {
		if len(chs) > 0 {
			return chs[0].Constraints().GetError(e)
		}
		return e
	} else {
		return errors.NewUnexpected(err.Error())
	}
}

// must is grimoire version of paranoid.Panic without context, but only original error.
func must(err error) {
	if err != nil {
		panic(err)
	}
}
