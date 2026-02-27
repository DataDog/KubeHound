package libkube

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sync"
)

const (
	DefaultNodeGroup     = "system:nodes"
	DefaultNodeNamespace = ""
)

var (
	ErrMissingNodeUser = errors.New("unable to resolve node user id")
)

var (
	lookupOnce sync.Once
	lookupNid  int64
	errLookup  error
)

// NodeUser will return the full name of the dedicated node user.
// See reference for details: https://kubernetes.io/docs/reference/access-authn-authz/node/
func NodeUser(nodeName string) string {
	return fmt.Sprintf("system:node:%s", nodeName)
}

// DefaultNodeIdentity will return the store id of the default system:nodes group.
func DefaultNodeIdentity(ctx context.Context, db *sql.DB, runID, clusterName string) (int64, error) {
	lookupOnce.Do(func() {
		err := db.QueryRowContext(ctx,
			"SELECT id FROM identities WHERE name = ? AND namespace = ? AND run_id = ? AND cluster_name = ? LIMIT 1",
			DefaultNodeGroup, DefaultNodeNamespace, runID, clusterName).Scan(&lookupNid)
		switch {
		case err == nil:
			// NOP
		case errors.Is(err, sql.ErrNoRows):
			errLookup = ErrMissingNodeUser
		default:
			errLookup = err
		}
	})

	return lookupNid, errLookup
}

// NodeIdentity will either return the store id of the dedicated node user or the default system:nodes group.
func NodeIdentity(ctx context.Context, db *sql.DB, nodeName, runID, clusterName string) (int64, error) {
	var nid int64
	err := db.QueryRowContext(ctx,
		"SELECT id FROM identities WHERE name = ? AND namespace = ? AND run_id = ? AND cluster_name = ? LIMIT 1",
		NodeUser(nodeName), DefaultNodeNamespace, runID, clusterName).Scan(&nid)
	switch {
	case err == nil:
		return nid, nil
	case errors.Is(err, sql.ErrNoRows):
		return DefaultNodeIdentity(ctx, db, runID, clusterName)
	}

	return 0, fmt.Errorf("resolving node identity (%s): %w", nodeName, err)
}

func ResetOnce() {
	lookupOnce = sync.Once{}
	lookupNid = 0
	errLookup = nil
}
