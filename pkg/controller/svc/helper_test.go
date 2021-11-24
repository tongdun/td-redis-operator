package svc

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/labels"
)

func TestParseSelector(t *testing.T) {
	cases := []struct {
		desc string
		str  string
		s    labels.Selector
		err  error
	}{
		{
			desc: "normal case",
			str:  "app=tdb",
			s: labels.SelectorFromSet(labels.Set{
				"app": "tdb",
			}),
		},
		{
			desc: "multiple keys",
			str:  "app=tdb,name=xxx",
			s: labels.SelectorFromSet(labels.Set{
				"app":  "tdb",
				"name": "xxx",
			}),
		},
		{
			desc: "multiple keys and contains many spaces",
			str:  " app = tdb , name = xxx",
			s: labels.SelectorFromSet(labels.Set{
				"app":  "tdb",
				"name": "xxx",
			}),
		},
		{
			desc: "duplicate key",
			str:  "app=tdb,app=xxx",
			err:  fmt.Errorf("duplicate key app"),
		},
		{
			desc: "parse kv error",
			str:  "app==tdb,app=xxx",
			err:  fmt.Errorf("can't parse kv app==tdb"),
		},
	}

	for _, c := range cases {
		s, err := parseSelector(c.str)
		assert.Equal(t, c.err, err, c.desc)
		assert.Equal(t, c.s, s, c.desc)
	}
}
