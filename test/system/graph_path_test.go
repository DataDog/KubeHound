package system

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TODO create a test suite with a graph client
// TODO query for TOKEN_STEAL paths
// TODO check expected paths exists!

func TestPath_TOKEN_STEAL(t *testing.T) {
	// TODO assert that the expected node vertices exist in the graph DB
	assert.False(t, true)
}

// TODO ensure identities are not created twice e.g system:kube-scheduler
// WHen writing an identity, check it is not already in the cache!!
