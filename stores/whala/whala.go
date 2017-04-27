package whala

import (
	"errors"

	"github.com/leeola/kala"
	"github.com/leeola/kala/util/kalautil"
	"github.com/leeola/whereis"
)

// Id prefix is used to make the kala id more unique.
const KalaIdPrefix = "whereis-finder-"

type Config struct {
	// Kala is the Kala interface to use as the data store.
	Kala kala.Kala
}

// Whala implements a whereis.Store interface for the Kala datastore.
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

func (k *Whala) Add(i whereis.Item) error {
	c := kala.Commit{
		Id: KalaId(i),
	}

	j, err := kalautil.ToJson(i)
	if err != nil {
		return err
	}

	if _, err := k.kala.Write(c, j, nil); err != nil {
		return err
	}

	return nil
}

// KalaId returns the Kala id for the given item.
func KalaId(i whereis.Item) string {
	return KalaIdPrefix + i.Id
}
