package edge

import "go.mongodb.org/mongo-driver/bson/primitive"

const (
	RoleBindLabel = "ROLE_BIND"
)

type roleBindGroup struct {
	PermissionSet primitive.ObjectID `bson:"_id" json:"permission_set"`
}
