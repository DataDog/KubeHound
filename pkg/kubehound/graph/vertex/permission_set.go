package vertex

import (
	"database/sql"
	"encoding/json"
	"strings"

	"github.com/DataDog/KubeHound/pkg/kubehound/graph/types"
	"github.com/DataDog/KubeHound/pkg/kubehound/models/store"
	"github.com/DataDog/KubeHound/pkg/kubehound/risk"
	rbacv1 "k8s.io/api/rbac/v1"
)

const (
	PermissionSetLabel = "PermissionSet"
)

var _ Builder = (*PermissionSet)(nil)

type PermissionSet struct {
	BaseVertex
}

func (v *PermissionSet) Label() string {
	return PermissionSetLabel
}

func (v *PermissionSet) BatchSize() int {
	return v.cfg.BatchSizeSmall
}

func (v *PermissionSet) Query(runID, clusterName string) string {
	return "SELECT id, name, role_name, role_binding_name, is_namespaced, namespace, rules, app, team, service, run_id, cluster_name FROM permissionsets WHERE run_id = '" + runID + "' AND cluster_name = '" + clusterName + "'"
}

func (v *PermissionSet) Scanner(rows *sql.Rows) (map[string]any, error) {
	var id int64
	var name, roleName, roleBindingName, namespace, rulesJSON, app, team, service, runID, cluster string
	var isNamespaced int
	if err := rows.Scan(&id, &name, &roleName, &roleBindingName, &isNamespaced, &namespace, &rulesJSON, &app, &team, &service, &runID, &cluster); err != nil {
		return nil, err
	}

	// Flatten policy rules to string array
	var policyRules []rbacv1.PolicyRule
	_ = json.Unmarshal([]byte(rulesJSON), &policyRules)
	rules := flattenPolicyRules(policyRules)

	// Risk engine check
	critical := risk.Engine().IsCritical(&store.PermissionSet{
		RoleName:    roleName,
		IsNamespaced: isNamespaced == 1,
	})

	return map[string]any{
		"storeID":      store.Hex(id),
		"name":         name,
		"role":         roleName,
		"roleBinding":  roleBindingName,
		"isNamespaced": isNamespaced == 1,
		"namespace":    namespace,
		"rules":        rules,
		"critical":     critical,
		"app":          app,
		"team":         team,
		"service":      service,
		"cluster":      cluster,
		"runID":        runID,
	}, nil
}

func (v *PermissionSet) Traversal() types.VertexTraversal {
	return v.DefaultTraversal(v.Label())
}

// flattenPolicyRules flattens the policy rule array into a string array.
func flattenPolicyRules(input []rbacv1.PolicyRule) []string {
	rules := make([]string, 0, len(input))
	for _, i := range input {
		var sb strings.Builder
		sb.WriteString("API(")
		sb.WriteString(strings.Join(i.APIGroups, ","))
		sb.WriteString(")::")
		sb.WriteString("R(")
		sb.WriteString(strings.Join(i.Resources, ","))
		sb.WriteString(")::")
		sb.WriteString("N(")
		sb.WriteString(strings.Join(i.ResourceNames, ","))
		sb.WriteString(")::")
		sb.WriteString("V(")
		sb.WriteString(strings.Join(i.Verbs, ","))
		sb.WriteString(")")
		rules = append(rules, sb.String())
	}
	return rules
}
