package node

import "testing"

func TestParseFromNodeString(t *testing.T) {
	nodestring :="127.0.0.1:12000"

	n, err := ParseNodeString(nodestring)
	if err != nil {
		t.Error(err)
	}

	if n == nil {
		t.Error("Nil result")
	}

	if n.Host != "127.0.0.1" {
		t.Error("Unexpected Host", n.Host)
	}

	if n.Port != 12000 {
		t.Error("Unexpected Port")
	}
}

func TestParseFromBadNodeString(t *testing.T) {
	nodestring := "127.0.0.1:12000:asd"
	_, err := ParseNodeString(nodestring)
	if err == nil {
		t.Error(err)
	}
}