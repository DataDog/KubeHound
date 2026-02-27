package edge

const (
	RoleBindLabel = "ROLE_BIND"
)

type roleBindGroup struct {
	PermissionSet int64 `json:"permission_set"`
}
