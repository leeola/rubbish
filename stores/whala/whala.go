package whala

import (
	"errors"
	"fmt"

	"github.com/leeola/fixity"
	"github.com/leeola/fixity-rubbish"
	"github.com/leeola/fixity/q"
)

const (
	// IdPrefix is used to make the fixity id more unique.
	IdPrefix = "rubbish-finder-"

	// MultiJsonKey is used to unmarshal our item document from the MultiJson.
	MultiJsonKey = "item"
)

type Config struct {
	// Fixity is the Fixity interface to use as the data store.
	Fixity fixity.Fixity
}

// Whala implements a rubbish.Store interface for the Fixity datastore.
//
// The strange name is a shortened Whereis Fixity combination, joined so that
// annoying import conflicts between fixity.Fixity and this package are reduced.
type Whala struct {
	fixity fixity.Fixity
}

func New(c Config) (*Whala, error) {
	if c.Fixity == nil {
		return nil, errors.New("missing required field: Fixity")
	}

	return &Whala{
		fixity: c.Fixity,
	}, nil
}

// incrementId iterates over the id to attempt and create a unique item.
//
// Note that this just queries over the existing matching IDs, and that
// this does not ensure a unique id. Currently Fixity does not enforce a central
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
		hashes, err := k.fixity.Search(q)
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
	var c fixity.Commit
	if i.Id == "" {
		id, err := k.incrementId(i.Name)
		if err != nil {
			return "", err
		}
		c.Id = id
	} else {
		c.Id = i.Id
	}

	// Do not store the Id, as Version already stores it.
	i.Id = ""

	var fields fixity.Fields
	fields.Append(fixity.Field{
		Field:   "name",
		Options: (fixity.FieldOptions{}).FullTextSearch(),
	})
	if i.ContainerId != "" {
		fields.Append(fixity.Field{
			Field: "containerId",
		})
	}
	if i.Description != "" {
		fields.Append(fixity.Field{
			Field:   "description",
			Options: (fixity.FieldOptions{}).FullTextSearch(),
		})
	}
	if len(i.Tags) > 0 {
		fields.Append(fixity.Field{
			Field: "tags",
		})
	}

	multiJson := fixity.MultiJson{}
	if err := multiJson.MarshalWithFields(MultiJsonKey, i, fields); err != nil {
		return "", err
	}

	if _, err := k.fixity.Write(c, multiJson, nil); err != nil {
		return "", err
	}

	return c.Id, nil
}

func (k *Whala) Search(s string, ts []string) ([]rubbish.Item, error) {
	var cs q.Constraints
	if s != "" {
		cs.Fts("*", s)
	}
	for _, t := range ts {
		cs.In("tags", t)
	}
	q := q.New().And(cs...).Limit(25)

	if len(cs) == 0 {
		return nil, errors.New("no queries specified")
	}

	hashes, err := k.fixity.Search(q)
	if err != nil {
		return nil, err
	}

	items := make([]rubbish.Item, len(hashes))
	for i, h := range hashes {
		v, err := k.fixity.ReadHash(h)
		if err != nil {
			return nil, err
		}

		var item rubbish.Item
		if err := v.MultiJson.Unmarshal(MultiJsonKey, &item); err != nil {
			return nil, err
		}

		item.Id = v.Id

		items[i] = item
	}

	return items, nil
}

func (k *Whala) SearchDescription(s string, ts []string) ([]rubbish.Item, error) {
	var cs q.Constraints
	if s != "" {
		cs.Fts("*", s)
	}
	for _, t := range ts {
		cs.In("tags", t)
	}
	q := q.New().And(cs...).Limit(25)

	if len(cs) == 0 {
		return nil, errors.New("no queries specified")
	}

	hashes, err := k.fixity.Search(q)
	if err != nil {
		return nil, err
	}

	items := make([]rubbish.Item, len(hashes))
	for i, h := range hashes {
		v, err := k.fixity.ReadHash(h)
		if err != nil {
			return nil, err
		}

		var item rubbish.Item
		if err := v.MultiJson.Unmarshal(MultiJsonKey, &item); err != nil {
			return nil, err
		}

		item.Id = v.Id

		items[i] = item
	}

	return items, nil
}

// FixityId returns the Fixity id for the given item.
func FixityId(i rubbish.Item) string {
	return IdPrefix + i.Id
}
