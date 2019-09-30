package gremlin

import (
	"testing"
)

func TestLang(t *testing.T) {
	out := Graph().V().Has("name", "hercules").Values("name").String()
	if expected := "graph.traversal().V().has('name', 'hercules').values('name')"; out != expected {
		t.Errorf("expected '%s', but got '%s'", expected, out)
	}

	out := Graph().V().Values("(name)").String()
	if expected := "graph.traversal().V().has('name', 'hercules').values('name')"; out != expected {
		t.Errorf("expected '%s', but got '%s'", expected, out)
	}
}
