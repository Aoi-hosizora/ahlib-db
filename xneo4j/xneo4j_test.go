package xneo4j

import (
	"github.com/neo4j/neo4j-go-driver/neo4j"
	"github.com/sirupsen/logrus"
	"log"
	"os"
	"testing"
	"time"
)

/*

func TestLogrus(t *testing.T) {
	authParam := neo4j.BasicAuth("neo4j", "123", "")
	driver, err := neo4j.NewDriver("bolt://localhost:7687", authParam)
	if err != nil {
		log.Fatalln("Failed to connect neo4j: ", err)
	}

	session, err := driver.Session(neo4j.AccessModeRead)
	if err != nil {
		log.Fatalln("Failed to create neo4j session: ", err)
	}

	session = NewLogrusNeo4j(session, logrus.New(), true)

	cypher := "MATCH p = ()-[r :FRIEND]->(n) RETURN r, n"
	rec, _ := neo4j.Collect(session.Run(cypher, nil))
	for _, r := range rec {
		rel := GetRel(r.Values()[0])
		node := GetNode(r.Values()[1])
		// log.Println(rel.Id(), rel.Type(), node.Id(), node.Props())
		_, _ = rel, node
	}

	cypher = "MATCH p = (n)-[r :FRIEND]->() WHERE n.uid > $uid RETURN n"
	rec, _ = neo4j.Collect(session.Run(cypher, map[string]interface{}{"uid": 3}))
	for _, r := range rec {
		node := GetNode(r.Values()[0])
		// log.Println(node.Id(), node.Props())
		_ = node
	}
}

func TestLogger(t *testing.T) {
	authParam := neo4j.BasicAuth("neo4j", "123", "")
	driver, err := neo4j.NewDriver("bolt://localhost:7687", authParam)
	if err != nil {
		log.Fatalln("Failed to connect neo4j: ", err)
	}

	session, err := driver.Session(neo4j.AccessModeRead)
	if err != nil {
		log.Fatalln("Failed to create neo4j session: ", err)
	}

	logger := log.New(os.Stderr, "", log.LstdFlags)
	session = NewLoggerNeo4j(session, logger, true)

	cypher := "MATCH p = ()-[r :FRIEND]->(n) RETURN r, n"
	rec, _ := neo4j.Collect(session.Run(cypher, nil))
	for _, r := range rec {
		rel := GetRel(r.Values()[0])
		node := GetNode(r.Values()[1])
		// log.Println(rel.Id(), rel.Type(), node.Id(), node.Props())
		_, _ = rel, node
	}

	cypher = "MATCH p = (n)-[r :FRIEND]->() WHERE n.uid > $uid RETURN n"
	rec, _ = neo4j.Collect(session.Run(cypher, map[string]interface{}{"uid": 3}))
	for _, r := range rec {
		node := GetNode(r.Values()[0])
		// log.Println(node.Id(), node.Props())
		_ = node
	}
}

*/

func TestXXX(t *testing.T) {
	driver, err := neo4j.NewDriver("bolt://localhost:7687", neo4j.BasicAuth("neo4j", "123", ""))
	if err != nil {
		log.Fatalln(1, err)
	}
	session, err := driver.Session(neo4j.AccessModeRead)
	if err != nil {
		log.Fatalln(2, err)
	}

	l1 := logrus.New()
	l1.SetFormatter(&logrus.TextFormatter{ForceColors: true, FullTimestamp: true, TimestampFormat: time.RFC3339})
	l2 := log.New(os.Stderr, "", log.LstdFlags)
	session = NewLogrusLogger(session, l1, WithSkip(2))
	session = NewLoggerLogger(session, l2)

	result, err := session.Run(`match (n {uid: 8}) return n limit 1`, nil)
	if err != nil {
		// Connection error: dial tcp [::1]:7688: connectex: No connection could be made because the target machine actively refused it.
		log.Fatalln(3, err)
	}
	summary, err := result.Summary()
	if err != nil {
		// Server error: [Neo.ClientError.Statement.SyntaxError] Invalid input 'n' (line 1, column 26 (offset: 25))
		log.Fatalln(4, err)
	}
	record, err := neo4j.Single(result, err)
	if err != nil {
		log.Fatalln(5, err)
	}
	log.Println(summary.ResultAvailableAfter(), summary.ResultConsumedAfter())
	log.Println(record.GetByIndex(0).(neo4j.Node).Props())
}
