package whala

import (
	"errors"
	"fmt"

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

// incrementId iterates over the id to attempt and create a unique item.
//
// Note that this just queries over the existing matching IDs, and that
// this does not ensure a unique id. Currently Kala does not enforce a central
// method to ensure unique id's, so we have to use an unsafe method like this
// attempt to create a unique id.
func (k *Whala) incrementId(name string) (string, error) {
	var increment int

	pageSize := 10
	for page := 0; increment < 100; page++ {
		q := q.New().
			Limit(pageSize).
			Skip(page * pageSize).
			Const(q.Eq("name", name))
		hashes, err := k.kala.Search(q)
		if err != nil {
			return "", err
		}
		total := len(hashes)
		increment = increment + total

		if total < pageSize {
			break
		}
	}

	if increment >= 100 {
		return "", errors.New("name is too ambiguous")
	}

	return fmt.Sprintf("%s_%d", name, increment+1), nil
}

func (k *Whala) Add(i rubbish.Item) (string, error) {
	var c kala.Commit
	if i.Id == "" {
		id, err := k.incrementId(i.Name)
		if err != nil {
			return "", err
		}
		c.Id = id
	} else {
		c.Id = i.Id
	}

	j, err := kalautil.MarshalJson(i)
	if err != nil {
		return "", err
	}

	j.Meta.IndexedFields.Append(kala.Field{
		Field: "name",
	})
	if i.ContainerId != "" {
		j.Meta.IndexedFields.Append(kala.Field{
			Field: "container-id",
		})
	}
	if i.Description != "" {
		j.Meta.IndexedFields.Append(kala.Field{
			Field: "description",
		})
	}

	if _, err := k.kala.Write(c, j, nil); err != nil {
		return "", err
	}

	return c.Id, nil
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
