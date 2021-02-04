package xneo4j

import (
	"github.com/neo4j/neo4j-go-driver/neo4j"
	"net/url"
)

// DialHandler represents a neo4j.Session dial function, used to get a session from neo4j.Driver.
type DialHandler func(driver neo4j.Driver, accessMode neo4j.AccessMode, bookmarks ...string) (neo4j.Session, error)

// Pool represents a neo4j.Session pool, actually this is a neo4j.Driver wrapper container with a custom DialHandler.
//
// Some notes:
//
// 1. Each neo4j.Driver instance maintains a pool of connections inside, as a result, it is recommended to only use one driver
// per application.
//
// 2. It is considerably cheap to create new neo4j.Session and neo4j.Transaction, as neo4j.Session and neo4j.Transaction do not
// create new connections as long as there are free connections available in the connection pool.
//
// 3. The neo4j.Driver is thread-safe, while the neo4j.Session or the neo4j.Transaction is not thread-safe.
type Pool struct {
	// driver represents the wrapped neo4j.Driver, we only expose the Target, VerifyConnectivity, Close methods,
	// and add Dial, DialReadMode, DialWriteMode methods as the replacement of Session, NewSession methods.
	driver neo4j.Driver

	// dial represents the custom DialHandler.
	dial DialHandler
}

const (
	panicNilDriver = "xneo4j: using nil driver"
)

// NewPool creates a Pool using given neo4j.Driver and DialHandler, panics when using nil neo4j.Driver, uses default DialHandler when giving nil value.
// Also see neo4j.NewDriver, neo4j.Config, neo4j.SessionConfig.
//
// Example:
//	driver, _ := neo4j.NewDriver(uri, neo4j.BasicAuth(username, password, realm), func(config *neo4j.Config) {
// 		config.MaxConnectionPoolSize = 10
// 	})
// 	pool := xneo4j.NewPool(driver, func(driver neo4j.Driver, accessMode neo4j.AccessMode, bookmarks ...string) (neo4j.Session, error) {
// 		return driver.NewSession(neo4j.SessionConfig{
// 			AccessMode:   accessMode,
// 			DatabaseName: database, // custom config
// 		})
// 	})
func NewPool(driver neo4j.Driver, dial DialHandler) *Pool {
	if driver == nil {
		panic(panicNilDriver)
	}
	if dial == nil {
		dial = func(driver neo4j.Driver, accessMode neo4j.AccessMode, bookmarks ...string) (neo4j.Session, error) {
			return driver.Session(accessMode, bookmarks...)
		}
	}
	return &Pool{driver: driver, dial: dial}
}

// Dial dials and returns a new neo4j.Session from neo4j.Driver's pool.
func (p *Pool) Dial(mode neo4j.AccessMode, bookmarks ...string) (neo4j.Session, error) {
	return p.dial(p.driver, mode, bookmarks...)
}

// DialReadMode dials and returns a new neo4j.Session with neo4j.AccessModeRead from neo4j.Driver's pool.
func (p *Pool) DialReadMode(bookmarks ...string) (neo4j.Session, error) {
	return p.dial(p.driver, neo4j.AccessModeRead, bookmarks...)
}

// DialWriteMode dials and returns a new neo4j.Session with neo4j.AccessModeWrite from neo4j.Driver's pool.
func (p *Pool) DialWriteMode(bookmarks ...string) (neo4j.Session, error) {
	return p.dial(p.driver, neo4j.AccessModeWrite, bookmarks...)
}

// Target returned the url.URL that this driver is bootstrapped.
func (p *Pool) Target() url.URL {
	return p.driver.Target()
}

// VerifyConnectivity verifies the driver can connect to a remote server or cluster by establishing a network connection with the remote.
func (p *Pool) VerifyConnectivity() error {
	return p.driver.VerifyConnectivity()
}

// Close closes the neo4j.Driver and all underlying connections.
func (p *Pool) Close() error {
	return p.driver.Close()
}
