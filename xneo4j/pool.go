package xneo4j

import (
	"github.com/neo4j/neo4j-go-driver/neo4j"
	"net/url"
)

// DialHandler represents a neo4j.Session dial function, used to get a session from neo4j.Driver.
type DialHandler func(driver neo4j.Driver, config *neo4j.SessionConfig) (neo4j.Session, error)

// Pool represents a neo4j.Session pool, implements neo4j.Driver interface. Actually this is a neo4j.Driver wrapper container with a DialHandler with
// custom neo4j.Session config.
//
// Some tips:
//
// 1. Each neo4j.Driver instance maintains a pool of connections inside, as a result, it is recommended to only use one driver per application.
//
// 2. It is considerably cheap to create new neo4j.Session and neo4j.Transaction, as neo4j.Session and neo4j.Transaction do not create new connections
// as long as there are free connections available in the connection pool.
//
// 3. The neo4j.Driver is thread-safe, while the neo4j.Session or the neo4j.Transaction is not thread-safe.
type Pool struct {
	driver neo4j.Driver
	dial   DialHandler
}

// DefaultDatabase is a marker for using the default database instance.
const DefaultDatabase = ""

const (
	panicNilDriver = "xneo4j: nil driver"
)

// NewPool creates a Pool using given neo4j.Driver and DialHandler, panics when using nil neo4j.Driver. Also see neo4j.NewDriver, neo4j.Config and
// neo4j.SessionConfig.
//
// Example:
// 	driver, _ := neo4j.NewDriver(uri, neo4j.BasicAuth(username, password, realm), xneo4j.WithMaxConnectionPoolSize(10))
// 	pool := xneo4j.NewPool(driver, func(driver neo4j.Driver, config *neo4j.SessionConfig) (neo4j.Session, error) {
// 		return driver.Session(config.AccessMode, config.Bookmarks...) // use default database
// 	})
func NewPool(driver neo4j.Driver, dial DialHandler) *Pool {
	if driver == nil {
		panic(panicNilDriver)
	}
	if dial == nil {
		dial = func(driver neo4j.Driver, config *neo4j.SessionConfig) (neo4j.Session, error) {
			return driver.NewSession(*config)
		}
	}
	return &Pool{driver: driver, dial: dial}
}

// Dial dials and returns a new neo4j.Session from neo4j.Driver's pool.
func (p *Pool) Dial(mode neo4j.AccessMode, bookmarks ...string) (neo4j.Session, error) {
	config := &neo4j.SessionConfig{AccessMode: mode, Bookmarks: bookmarks, DatabaseName: DefaultDatabase}
	return p.dial(p.driver, config)
}

// NewSession dials and returns a new neo4j.Session from neo4j.Driver's pool.
func (p *Pool) NewSession(config neo4j.SessionConfig) (neo4j.Session, error) {
	return p.dial(p.driver, &config)
}

// Session dials and returns a new neo4j.Session from neo4j.Driver's pool.
func (p *Pool) Session(mode neo4j.AccessMode, bookmarks ...string) (neo4j.Session, error) {
	return p.Dial(mode, bookmarks...)
}

// DialReadMode dials and returns a new neo4j.Session with neo4j.AccessModeRead from neo4j.Driver's pool.
func (p *Pool) DialReadMode(bookmarks ...string) (neo4j.Session, error) {
	return p.Dial(neo4j.AccessModeRead, bookmarks...)
}

// DialWriteMode dials and returns a new neo4j.Session with neo4j.AccessModeWrite from neo4j.Driver's pool.
func (p *Pool) DialWriteMode(bookmarks ...string) (neo4j.Session, error) {
	return p.Dial(neo4j.AccessModeWrite, bookmarks...)
}

// Driver returned the neo4j.Driver from Pool.
func (p *Pool) Driver() neo4j.Driver {
	return p.driver
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
