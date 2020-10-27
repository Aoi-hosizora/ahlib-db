package xneo4j

import (
	"github.com/neo4j/neo4j-go-driver/neo4j"
)

// DialFunc is used to get a session from neo4j.Driver.
type DialFunc func(driver neo4j.Driver, accessMode neo4j.AccessMode, bookmarks ...string) (neo4j.Session, error)

type Pool struct {
	neo4j.Driver
	Dial DialFunc
}

// Each neo4j.Driver instance maintains a pool of connections inside, as a result, it is recommended to only use one driver per application.
//
// It is considerably cheap to create new sessions and transactions, as sessions and transactions do not create new connections
// as long as there are free connections available in the connection pool.
//
// The neo4j.Driver is thread-safe, while the neo4j.Session or the transaction is not thread-safe.
func NewPool(driver neo4j.Driver, dial DialFunc) *Pool {
	return &Pool{
		Driver: driver,
		Dial:   dial,
	}
}

// Get a session from neo4j.Driver.
func (n *Pool) Get(mode neo4j.AccessMode, bookmarks ...string) (neo4j.Session, error) {
	return n.Dial(n.Driver, mode, bookmarks...)
}

// Get a session (neo4j.AccessModeWrite) from neo4j.Driver.
func (n *Pool) GetWriteMode(bookmarks ...string) (neo4j.Session, error) {
	return n.Dial(n.Driver, neo4j.AccessModeWrite, bookmarks...)
}

// Get a session (neo4j.AccessModeRead) from neo4j.Driver.
func (n *Pool) GetReadMode(bookmarks ...string) (neo4j.Session, error) {
	return n.Dial(n.Driver, neo4j.AccessModeRead, bookmarks...)
}
