// Package testing defines utils for testing
package testing

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/util/diff"
	kubetesting "k8s.io/client-go/testing"
)

// AssertActionCounts can count all kinds of actions and compare with expected count
func AssertActionCounts(t *testing.T, expected map[string]int, actions []kubetesting.Action) {
	actual := map[string]int{}
	for k := range expected {
		actual[k] = 0
	}

	for _, a := range actions {
		verb := a.GetVerb()

		v, ok := actual[verb]
		if ok {
			actual[verb] = v + 1
		}
	}

	assert.Equal(t, expected, actual, "action counts are not equal, expected: %v, actual: %v", expected, actual)
}

// AssertWriteAction will assert that expected action should be equal with
// actual action
// AssertWriteAction will ignore all read actions
// TODO(bo.liub): now only support create, update and patch verb
func AssertWriteAction(t *testing.T, expected, actual kubetesting.Action) {
	matched := expected.Matches(actual.GetVerb(), actual.GetResource().Resource)
	require.Equal(t,
		true,
		matched,
		"Expected\n\t%#v\ngot\n\t%#v",
		expected,
		actual,
	)
	require.Equal(t,
		expected.GetSubresource(),
		actual.GetSubresource(),
		"Expected\n\t%#v\ngot\n\t%#v",
		expected,
		actual,
	)
	require.Equal(t,
		reflect.TypeOf(expected),
		reflect.TypeOf(actual),
		"Action has wrong type. Expected: %t, actual: %t",
		expected,
		actual,
	)

	switch actual.GetVerb() {
	case "create":
		a, _ := actual.(kubetesting.CreateAction)
		e, _ := expected.(kubetesting.CreateAction)
		expObject := e.GetObject()
		object := a.GetObject()
		assert.Equal(t,
			expObject,
			object,
			"Action %s %s has wrong object\nDiff: \n %s",
			a.GetVerb(),
			a.GetResource().Resource,
			diff.ObjectDiff(expObject, object),
		)

	case "update":
		a, _ := actual.(kubetesting.UpdateAction)
		e, _ := expected.(kubetesting.UpdateAction)
		expObject := e.GetObject()
		object := a.GetObject()
		assert.Equal(t,
			expObject,
			object,
			"Action %s %s has wrong object\nDiff: \n %s",
			a.GetVerb(),
			a.GetResource().Resource,
			diff.ObjectDiff(expObject, object),
		)

	case "patch":
		a, _ := actual.(kubetesting.PatchAction)
		e, _ := expected.(kubetesting.PatchAction)
		expPatch := e.GetPatch()
		patch := a.GetPatch()
		assert.Equal(t,
			expPatch,
			patch,
			"Action %s %s has wrong patch\nDiff: \n %s",
			a.GetVerb(),
			a.GetResource().Resource,
			diff.ObjectDiff(expPatch, patch),
		)
	}
}

// AssertWriteActions will assert a collection of actions
func AssertWriteActions(t *testing.T, expected, actual []kubetesting.Action) {
	filtered := []kubetesting.Action{}

	for _, action := range actual {
		switch action.GetVerb() {
		case "create":
			filtered = append(filtered, action)
		case "update":
			filtered = append(filtered, action)
		case "patch":
			filtered = append(filtered, action)
		}
	}

	for i, a := range filtered {
		if len(expected) < i+1 {
			require.Fail(t,
				"unexpected actions", "%d unexpected actions: %+v",
				len(filtered)-len(expected),
				filtered[i:],
			)
		}

		AssertWriteAction(t, expected[i], a)
	}

	if len(expected) > len(filtered) {
		assert.Fail(t,
			"additional expected actions",
			"%d additional expected actions: %+v",
			len(expected)-len(filtered),
			expected[len(filtered):],
		)
	}
}
