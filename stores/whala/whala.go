package whala

import (
	"errors"

	"github.com/leeola/kala"
	"github.com/leeola/kala/q"
	"github.com/leeola/kala/util/kalautil"
	"github.com/leeola/rubbish"
)

// Id prefix is used to make the kala id more unique.
const KalaIdPrefix = "rubbish-finder-"

type Config struct {
	// Kala is the Kala interface to use as the data store.
	Kala kala.Kala
}

// Whala implements a rubbish.Store interface for the Kala datastore.
//
// The strange name is a shortened Whereis Kala combination, joined so that
// annoying import conflicts between kala.Kala and this package are reduced.
type Whala struct {
	kala kala.Kala
}

func New(c Config) (*Whala, error) {
	if c.Kala == nil {
		return nil, errors.New("missing required field: Kala")
	}

	return &Whala{
		kala: c.Kala,
	}, nil
}

func (k *Whala) Add(i rubbish.Item) error {
	c := kala.Commit{
		Id: KalaId(i),
	}

	j, err := kalautil.MarshalJson(i)
	if err != nil {
		return err
	}

	// we have to specify the indexing value here, as kala doesn't yet support
	// automatic value assertion.
	//
	// Seealso: https://github.com/leeola/kala/blob/master/impl/local/local.go#L98
	j.Meta.IndexedFields.Append(kala.Field{
		Field: "name",
		Value: i.Name,
	})
	if i.ContainerId != "" {
		j.Meta.IndexedFields.Append(kala.Field{
			Field: "container-id",
			Value: i.ContainerId,
		})
	}
	if i.Description != "" {
		j.Meta.IndexedFields.Append(kala.Field{
			Field: "description",
			Value: i.Description,
		})
	}

	if _, err := k.kala.Write(c, j, nil); err != nil {
		return err
	}

	return nil
}

func (k *Whala) SearchName(s string) ([]rubbish.Item, error) {
	q := q.New().Const(q.Eq("name", s))
	hashes, err := k.kala.Search(q)
	if err != nil {
		return nil, err
	}

	// faking loading here, for testing
	items := make([]rubbish.Item, len(hashes))
	for i, h := range hashes {
		v, err := k.kala.ReadHash(h)
		if err != nil {
			return nil, err
		}

		var item rubbish.Item
		if err := kalautil.UnmarshalJson(v.Json, &item); err != nil {
			return nil, err
		}

		items[i] = item
	}

	return items, nil
}

// KalaId returns the Kala id for the given item.
func KalaId(i rubbish.Item) string {
	return KalaIdPrefix + i.Id
}
