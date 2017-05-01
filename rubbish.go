package rubbish

import "errors"

type Item struct {
	// Id is a unique id for the given item.
	Id string `json:"id"`

	// Name of the item being stored.
	Name string `json:"name"`

	// ContainerId is the id of the item this item is within.
	ContainerId string `json:"containerId"`

	// Description of the item in question.
	Description string `json:"description"`
}

// Store implements basic storing and indexing of inventory items.
type Store interface {
	Add(Item) (string, error)
	SearchName(string) ([]Item, error)
}

type Config struct {
	Store Store
}

type Whereis struct {
	store Store
}

func New(c Config) (*Whereis, error) {
	if c.Store == nil {
		return nil, errors.New("missing required config value: Store")
	}

	return &Whereis{
		store: c.Store,
	}, nil
}
