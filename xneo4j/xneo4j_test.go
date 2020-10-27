package xneo4j

import (
	"github.com/Aoi-hosizora/ahlib/xproperty"
	"github.com/Aoi-hosizora/ahlib/xtesting"
	"github.com/neo4j/neo4j-go-driver/neo4j"
	"github.com/sirupsen/logrus"
	"log"
	"os"
	"testing"
)

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
	rec, _ := GetRecords(session.Run(cypher, nil))
	for _, r := range rec {
		rel := GetRel(r.Values()[0])
		node := GetNode(r.Values()[1])
		// log.Println(rel.Id(), rel.Type(), node.Id(), node.Props())
		_, _ = rel, node
	}

	cypher = "MATCH p = (n)-[r :FRIEND]->() WHERE n.uid > $uid RETURN n"
	rec, _ = GetRecords(session.Run(cypher, map[string]interface{}{"uid": 3}))
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
	rec, _ := GetRecords(session.Run(cypher, nil))
	for _, r := range rec {
		rel := GetRel(r.Values()[0])
		node := GetNode(r.Values()[1])
		// log.Println(rel.Id(), rel.Type(), node.Id(), node.Props())
		_, _ = rel, node
	}

	cypher = "MATCH p = (n)-[r :FRIEND]->() WHERE n.uid > $uid RETURN n"
	rec, _ = GetRecords(session.Run(cypher, map[string]interface{}{"uid": 3}))
	for _, r := range rec {
		node := GetNode(r.Values()[0])
		// log.Println(node.Id(), node.Props())
		_ = node
	}
}

func TestOrderBy(t *testing.T) {
	m := xproperty.PropertyDict{
		"a": xproperty.NewValue(false, "at"),
		"b": xproperty.NewValue(false, "bt1", "bt2"),
		"c": xproperty.NewValue(true, "ct"),
	}
	p := xproperty.VariableDict{"at": 1, "bt1": 2, "bt2": 1, "ct": 2}

	xtesting.Equal(t, OrderByFunc(m)("a asc, b desc, c asc", "o"), "o.at ASC, o.bt1 DESC, o.bt2 DESC, o.ct DESC")
	xtesting.Equal(t, OrderByFunc(m)("", ""), "")
	xtesting.Equal(t, OrderByFunc(m)("h", ""), "")
	xtesting.Equal(t, OrderByFunc(xproperty.PropertyDict{})("", ""), "")

	xtesting.Equal(t, OrderByFunc2(m, p)("a asc, b desc, c asc", "r1", "n1"), "r1.at ASC, n1.bt1 DESC, r1.bt2 DESC, n1.ct DESC")
	xtesting.Equal(t, OrderByFunc2(m, p)(""), "")
	xtesting.Equal(t, OrderByFunc2(m, p)("h"), "")
	xtesting.Equal(t, OrderByFunc2(xproperty.PropertyDict{}, xproperty.VariableDict{})(""), "")

}
