package pipeline

import (
	"context"

	"github.com/DataDog/KubeHound/pkg/globals"
)

const (
	RoleBindingIngestName = "k8s-role-binding-ingest"
)

type RoleBindingIngest struct {
	BaseObjectIngest
}

var _ ObjectIngest = (*RoleBindingIngest)(nil)

func (i RoleBindingIngest) Name() string {
	return RoleBindingIngestName
}

func (i RoleBindingIngest) Initialize(ctx context.Context, deps *Dependencies) error {
	return globals.ErrNotImplemented
}

func (i RoleBindingIngest) Run(ctx context.Context) error {
	return globals.ErrNotImplemented
}

func (i RoleBindingIngest) Close(ctx context.Context) error {
	return globals.ErrNotImplemented
}
