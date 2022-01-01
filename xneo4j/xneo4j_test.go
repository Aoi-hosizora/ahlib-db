package xneo4j

import (
	"github.com/Aoi-hosizora/ahlib/xtesting"
	"github.com/neo4j/neo4j-go-driver/neo4j"
	"github.com/sirupsen/logrus"
	"log"
	"reflect"
	"testing"
	"time"
)

const (
	neo4jParam      = "bolt://localhost:7687"
	neo4jWrongParam = "bolt://localhost:7688"
	neo4jUser       = "neo4j"
	neo4jPasswd     = "123"
)

func TestPool(t *testing.T) {
	driver, err := neo4j.NewDriver(neo4jParam, neo4j.BasicAuth(neo4jUser, neo4jPasswd, ""), WithEncrypted(false))
	if err != nil {
		log.Println(err)
		t.FailNow()
	}
	l := logrus.New()
	l.SetFormatter(&logrus.TextFormatter{ForceColors: true, FullTimestamp: true, TimestampFormat: time.RFC3339})
	pool := NewPool(driver, func(driver neo4j.Driver, accessMode neo4j.AccessMode, bookmarks ...string) (neo4j.Session, error) {
		session, err := driver.NewSession(neo4j.SessionConfig{
			AccessMode:   accessMode,
			Bookmarks:    bookmarks,
			DatabaseName: "",
		})
		if err != nil {
			return nil, err
		}
		return NewLogrusLogger(session, l), nil
	})
	check := func(session neo4j.Session, err error) neo4j.Session {
		if err != nil {
			log.Println(err)
			t.FailNow()
		}
		return session
	}

	// match
	session := check(pool.Dial(neo4j.AccessModeRead))
	records, summary, err := Collect(session.Run(`MATCH (n :XNEO4J_TEST) RETURN n`, nil))
	xtesting.Nil(t, err)
	xtesting.Empty(t, records)
	xtesting.Equal(t, summary.Counters().NodesCreated(), 0)

	// merge
	session = check(pool.DialWriteMode())
	records, summary, err = Collect(session.Run(`MERGE (n :XNEO4J_TEST { id: $id, name: $name }) RETURN n`, P{"id": 2, "name": "name2"}))
	xtesting.Nil(t, err)
	xtesting.Equal(t, len(records), 1)
	xtesting.Equal(t, summary.Counters().NodesCreated(), 1)
	xtesting.Equal(t, summary.Counters().LabelsAdded(), 1)
	xtesting.Equal(t, summary.Counters().PropertiesSet(), 2)
	node := GetNode(records[0].GetByIndex(0))
	xtesting.Equal(t, node.Props(), map[string]interface{}{"id": int64(2), "name": "name2"})

	// match
	session = check(pool.DialReadMode())
	records, summary, err = Collect(session.Run(`MATCH (n :XNEO4J_TEST { id: $id }) RETURN n`, P{"id": 2}))
	xtesting.Nil(t, err)
	xtesting.Equal(t, len(records), 1)
	xtesting.Equal(t, summary.Counters().NodesCreated(), 0)
	node = GetNode(records[0].GetByIndex(0))
	xtesting.Equal(t, node.Props(), map[string]interface{}{"id": int64(2), "name": "name2"})

	// delete
	session = check(pool.DialWriteMode())
	_, summary, err = Collect(session.Run(`MATCH (n :XNEO4J_TEST { id: $id }) DELETE n`, P{"id": 2}))
	xtesting.Nil(t, err)
	xtesting.Equal(t, summary.Counters().NodesDeleted(), 1)

	// others
	xtesting.Panic(t, func() { NewPool(nil, nil) })
	xtesting.NotPanic(t, func() {
		pool := NewPool(driver, nil)
		session := check(pool.Dial(neo4j.AccessModeRead))
		records, _, err := Collect(session.Run(`MATCH (n :XNEO4J_TEST) RETURN n`, nil))
		xtesting.Nil(t, err)
		xtesting.Empty(t, records)
	})
	target := pool.Target()
	xtesting.Equal(t, target.String(), neo4jParam)
	target = pool.Driver().Target()
	xtesting.Equal(t, target.String(), neo4jParam)
	xtesting.Nil(t, pool.VerifyConnectivity())
	xtesting.Nil(t, pool.Close())
	xtesting.NotNil(t, pool.VerifyConnectivity())
}

type (
	mockNode struct{}
	mockRel  struct{}
	mockPath struct{}
)

func (m *mockNode) Id() int64                           { return 0 }
func (m *mockNode) Labels() []string                    { return nil }
func (m *mockNode) Props() map[string]interface{}       { return nil }
func (m *mockRel) Id() int64                            { return 0 }
func (m *mockRel) StartId() int64                       { return 0 }
func (m *mockRel) EndId() int64                         { return 0 }
func (m *mockRel) Type() string                         { return "" }
func (m *mockRel) Props() map[string]interface{}        { return nil }
func (m *mockPath) Nodes() []neo4j.Node                 { return nil }
func (m *mockPath) Relationships() []neo4j.Relationship { return nil }

func TestHelper(t *testing.T) {
	t.Run("Collect", func(t *testing.T) {
		l := logrus.New()
		l.SetFormatter(&logrus.TextFormatter{ForceColors: true, FullTimestamp: true, TimestampFormat: time.RFC3339})

		// 1
		driver, err := neo4j.NewDriver(neo4jWrongParam, neo4j.BasicAuth(neo4jUser, neo4jPasswd, ""), WithEncrypted(false))
		if err != nil {
			log.Println(err)
			t.FailNow()
		}
		session, err := driver.Session(neo4j.AccessModeWrite)
		if err != nil {
			log.Println(err)
			t.FailNow()
		}
		session = NewLogrusLogger(session, l)

		_, _, err = Collect(session.Run(`MATCH (n) RETURN n LIMIT 1`, nil)) // error
		xtesting.NotNil(t, err)
		_ = driver.Close()

		// 2
		driver, err = neo4j.NewDriver(neo4jParam, neo4j.BasicAuth(neo4jUser, neo4jPasswd, ""), WithEncrypted(false))
		if err != nil {
			log.Println(err)
			t.FailNow()
		}
		session, err = driver.Session(neo4j.AccessModeWrite)
		if err != nil {
			log.Println(err)
			t.FailNow()
		}
		session = NewLogrusLogger(session, l)

		_, _, err = Collect(session.Run(`MATCH (n) RETURN m`, nil)) // error
		xtesting.NotNil(t, err)
		_, _, err = Collect(session.Run(`MATCH (n) RETURN n LIMIT 1`, nil)) // not error
		xtesting.Nil(t, err)
		_ = driver.Close()
	})

	t.Run("GetXXX", func(t *testing.T) {
		now := time.Now()
		for _, tc := range []struct {
			giveFn  interface{}
			giveArg interface{}
			wantArg interface{}
		}{
			{GetInteger, int64(1), int64(1)},
			{GetFloat, 1.5, 1.5},
			{GetString, "t", "t"},
			{GetBoolean, true, true},
			{GetByteArray, []byte("test"), []byte("test")},
			{GetList, []interface{}{0, true, 2.3, "t"}, []interface{}{0, true, 2.3, "t"}},
			{GetMap, map[string]interface{}{"a": 0, "b": 1.1}, map[string]interface{}{"a": 0, "b": 1.1}},
			{GetNode, &mockNode{}, &mockNode{}},
			{GetRel, &mockRel{}, &mockRel{}},
			{GetPath, &mockPath{}, &mockPath{}},
			{GetPoint, neo4j.NewPoint2D(11, 1, 2), neo4j.NewPoint2D(11, 1, 2)},
			{GetPoint, *neo4j.NewPoint2D(11, 1, 2), neo4j.NewPoint2D(11, 1, 2)},
			{GetDate, neo4j.DateOf(now), neo4j.DateOf(now)},
			{GetTime, neo4j.OffsetTimeOf(now), neo4j.OffsetTimeOf(now)},
			{GetDateTime, now, now},
			{GetLocalTime, neo4j.LocalTimeOf(now), neo4j.LocalTimeOf(now)},
			{GetLocalDateTime, neo4j.LocalDateTimeOf(now), neo4j.LocalDateTimeOf(now)},
			{GetDuration, neo4j.DurationOf(0, 1, 1, 0), neo4j.DurationOf(0, 1, 1, 0)},
		} {
			result := reflect.ValueOf(tc.giveFn).Call([]reflect.Value{reflect.ValueOf(tc.giveArg)})[0].Interface()
			if w, ok := tc.wantArg.(*neo4j.Point); ok {
				if g, ok := tc.giveArg.(neo4j.Point); ok {
					xtesting.Equal(t, g.String(), w.String())
				} else {
					xtesting.Equal(t, tc.giveArg.(*neo4j.Point).String(), w.String())
				}
			} else {
				xtesting.Equal(t, result, tc.wantArg)
			}
		}
	})

	t.Run("Order", func(t *testing.T) {
		dict := PropertyDict{
			"uid":      NewPropertyValue(false, "n.uid"),
			"username": NewPropertyValue(false, "n.firstname", "n.lastname"),
			"age":      NewPropertyValue(true, "r.birthday"),
		}
		for _, tc := range []struct {
			giveSource string
			giveDict   PropertyDict
			want       string
		}{
			{"uid, xxx", dict, "n.uid ASC"},
			{"uid desc xxx", dict, "n.uid DESC"},
			{"uid, username", dict, "n.uid ASC, n.firstname ASC, n.lastname ASC"},
			{"username desc, age desc", dict, "n.firstname DESC, n.lastname DESC, r.birthday ASC"},
		} {
			xtesting.Equal(t, GenerateOrderByExp(tc.giveSource, tc.giveDict), tc.want)
		}
	})
}

func TestLogger(t *testing.T) {
	l1 := logrus.New()
	l1.SetFormatter(&logrus.TextFormatter{ForceColors: true, FullTimestamp: true, TimestampFormat: time.RFC3339})
	l2 := log.Default()

	for _, tc := range []struct {
		name      string
		sessionFn func(neo4j.Session) neo4j.Session
	}{
		{"default", func(s neo4j.Session) neo4j.Session { return s }},
		{"logrus", func(s neo4j.Session) neo4j.Session { return NewLogrusLogger(s, l1) }},
		{"logrus_no_err", func(s neo4j.Session) neo4j.Session { return NewLogrusLogger(s, l1, WithLogErr(false)) }},
		{"logrus_no_cypher", func(s neo4j.Session) neo4j.Session { return NewLogrusLogger(s, l1, WithLogCypher(false)) }},
		{"logrus_field", func(s neo4j.Session) neo4j.Session {
			return NewLogrusLogger(s, l1, WithSkip(1), WithCounterField(true))
		}},
		{"logger", func(s neo4j.Session) neo4j.Session { return NewStdLogger(s, l2) }},
		{"logger_no_xxx", func(s neo4j.Session) neo4j.Session {
			return NewStdLogger(s, l2, WithLogErr(false), WithLogCypher(false))
		}},
		{"disable", func(s neo4j.Session) neo4j.Session { return NewLogrusLogger(s, l1) }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			driver, err := neo4j.NewDriver(neo4jParam, neo4j.BasicAuth(neo4jUser, neo4jPasswd, ""), WithEncrypted(false))
			if err != nil {
				log.Println(err)
				t.FailNow()
			}
			session, err := driver.Session(neo4j.AccessModeWrite)
			if err != nil {
				log.Println(err)
				t.FailNow()
			}
			session = tc.sessionFn(session)
			if tc.name != "disable" {
				EnableLogger()
			} else {
				DisableLogger()
			}

			_, _ = session.Run(`MATCH n RETURN n`, nil) // error
			_, _ = session.Run(`MATCH (n) RETURN n LIMIT 1`, nil)
			_, _ = session.Run(`MATCH (n :NOT_FOUND { id: $id }) RETURN n`, P{"id": 1})                                 // $ num
			_, _ = session.Run(`MATCH (n :NOT_FOUND { name: $name }) RETURN n`, P{"name": "name1"})                     // $ str
			_, _ = session.Run(`MATCH (n :NOT_FOUND { obj: $obj }) RETURN n`, P{"obj": nil})                            // nil
			_, _ = session.Run(`MATCH (n :NOT_FOUND { obj: $obj }) RETURN n`, P{"obj": []interface{}{"a"}})             // ...
			_, _ = session.Run(`MATCH (n :NOT_FOUND { obj: $obj }) RETURN n`, P{"obj": map[string]interface{}{"a": 0}}) // ...
			_, _ = session.Run(`MATCH (n :NOT_FOUND { date: $date, time: $time, datetime: $datetime, ltime: $ltime, ldatetime: $ldatetime, duration: $duration }) RETURN n`, P{
				"date":      neo4j.DateOf(time.Now().UTC()),
				"time":      neo4j.OffsetTimeOf(time.Now().UTC()),
				"datetime":  time.Now().UTC(),
				"ltime":     neo4j.LocalTimeOf(time.Now().UTC()),
				"ldatetime": neo4j.LocalDateTimeOf(time.Now().UTC()),
				"duration":  neo4j.DurationOf(0, 1, 1, 0),
			})
		})
	}
}
