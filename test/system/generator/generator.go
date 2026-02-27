//nolint:all
package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"go/format"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"text/template"
	"time"

	"database/sql"

	"github.com/DataDog/KubeHound/pkg/config"
	"github.com/DataDog/KubeHound/pkg/kubehound/models/converter"
	"github.com/DataDog/KubeHound/pkg/kubehound/models/shared"
	"github.com/DataDog/KubeHound/pkg/kubehound/models/store"
	"github.com/DataDog/KubeHound/pkg/kubehound/risk"
	"github.com/DataDog/KubeHound/pkg/kubehound/storage/storedb"
	"gopkg.in/yaml.v3"

	_ "modernc.org/sqlite"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"

	"k8s.io/client-go/kubernetes/scheme"
)

type Cluster struct {
	Kind       string `yaml:"kind"`
	APIVersion string `yaml:"apiVersion"`
	Name       string `yaml:"name"`
	Nodes      []struct {
		Role string `yaml:"role"`
	} `yaml:"nodes"`
}

const (
	defaultNamespace      = "default"
	defaultServiceAccount = "default"
)

// Local model types mirroring test/system/models.go for code generation.
type gPod struct {
	StoreID, Name, Namespace, ServiceAccount, Node string
	IsNamespaced, ShareProcessNamespace, Critical   bool
	Compromised                                     shared.CompromiseType
}
type gNode struct {
	StoreID, Name, Namespace string
	IsNamespaced, Critical   bool
	Compromised              shared.CompromiseType
}
type gContainer struct {
	StoreID, Name, Image, Pod, Node, Namespace string
	Command, Args, Capabilities, Ports         []string
	Privileged, PrivEsc, HostPID, HostIPC      bool
	HostNetwork, IsNamespaced                  bool
	RunAsUser                                  int64
	Compromised                                shared.CompromiseType
}
type gVolume struct {
	StoreID, Name, Type, SourcePath, MountPath, Namespace string
	Readonly, IsNamespaced                                bool
}
type gIdentity struct {
	StoreID, Name, Namespace, Type string
	IsNamespaced, Critical         bool
}
type gPermissionSet struct {
	StoreID, Name, Namespace, Role, RoleBinding string
	Rules                                       []string
	IsNamespaced, Critical                      bool
}

var (
	Containers     = make(map[string]gContainer)
	Pods           = make(map[string]gPod)
	Nodes          = make(map[string]gNode)
	PermissionSets = make(map[string]gPermissionSet)
	Identities     = make(map[string]gIdentity)
	Volumes        = make(map[string]gVolume)
)

var (
	GeneratorConfig = &config.KubehoundConfig{
		Dynamic: config.DynamicConfig{
			Cluster: config.DynamicClusterInfo{
				Name: "kind-kubehound",
			},
			RunID: config.NewRunID(),
		},
	}
)

var (
	globalHeaders = []byte(`// PLEASE DO NOT EDIT
// THIS HAS BEEN GENERATED AUTOMATICALLY on ` + time.Now().Format("2006-01-02 15:04") + `
//
// Generate it with "go generate ./..."
//
// currently support only:
// - nodes
// - pods
// - containers
// - volumes
//
// TODO: roles, rolebinding, clusterrole, clusterrolebindings

package system

import (
	"github.com/DataDog/KubeHound/pkg/kubehound/models/shared"
)

`)
)

func usage() {
	fmt.Println(`Usage:
	./generator <k8s_yaml_folder> <destination_file>`)
}

func main() {
	if len(os.Args) != 3 {
		usage()
		return
	}
	k8sDefinitionPath := os.Args[1]
	codegenPath := os.Args[2]

	ctx := context.Background()
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		fmt.Printf("sqlite open: %v", err)
		return
	}
	defer db.Close()

	if err := storedb.InitSchema(db); err != nil {
		fmt.Printf("sqlite schema: %v", err)
		return
	}

	provider := storedb.NewSQLiteProviderFromDB(db)
	cacheRoleBinding := []*rbacv1.RoleBinding{}
	cacheClusterRoleBinding := []*rbacv1.ClusterRoleBinding{}

	clusterFile, err := ioutil.ReadFile(filepath.Join(k8sDefinitionPath, "cluster.yaml"))
	if err != nil {
		log.Fatal(err)
	}
	err = ProcessCluster(clusterFile)
	if err != nil {
		log.Fatal(err)
	}

	attackPath := filepath.Join(k8sDefinitionPath, "attacks")
	filesAttack, err := ioutil.ReadDir(attackPath)
	if err != nil {
		log.Fatal(err)
	}
	for _, file := range filesAttack {
		ProcessFile(ctx, attackPath, file, provider, &cacheRoleBinding, &cacheClusterRoleBinding)
	}

	// Generate permissionsets
	ConvertRoleBindings(ctx, cacheRoleBinding, cacheClusterRoleBinding, db)
	outPermSets, err := GeneratePermissionSetTemplate()
	if err != nil {
		fmt.Println("failed to permission sets: ", err)
	}
	outIdentities, err := GenerateIdentityTemplate()
	if err != nil {
		fmt.Println("failed to permission sets: ", err)
	}

	outPods, err := GeneratePodTemplate()
	if err != nil {
		fmt.Println("failed to write pods: ", err)
	}
	outNodes, err := GenerateNodeTemplate()
	if err != nil {
		fmt.Println("failed to write pods: ", err)
	}
	outContainers, err := GenerateContainerTemplate()
	if err != nil {
		fmt.Println("failed to write pods: ", err)
	}
	outVolumes, err := GenerateVolumeTemplate()
	if err != nil {
		fmt.Println("failed to write pods: ", err)
	}
	fmt.Printf("volumes: %+v\n", Volumes)
	err = WriteTemplatesToFile(codegenPath, globalHeaders, outPods, outNodes, outVolumes, outContainers, outPermSets, outIdentities)
	if err != nil {
		fmt.Println(err)
	}
}

func ProcessCluster(content []byte) error {
	var cluster Cluster
	err := yaml.Unmarshal(content, &cluster)
	if err != nil {
		return err
	}

	for _, n := range cluster.Nodes {
		nodeName := "kubehound.test.local-" + n.Role
		for {
			orig := nodeName
			count := 2
			_, exist := Nodes[nodeName]
			if exist {
				nodeName = fmt.Sprintf("%s%d", orig, count)
				continue
			}
			break
		}

		Nodes[nodeName] = gNode{
			StoreID:      "",
			Name:         nodeName,
			IsNamespaced: false,
			Namespace:    "",
			Compromised:  0,
			Critical:     false,
		}
	}
	return nil
}

func ProcessFile(ctx context.Context, basePath string, file os.FileInfo, provider storedb.Provider, cacheRoleBinding *[]*rbacv1.RoleBinding, cacheClusterRoleBinding *[]*rbacv1.ClusterRoleBinding) {
	fmt.Println("Processing: " + file.Name())
	data, err := os.ReadFile(filepath.Join(basePath, file.Name()))
	if err != nil {
		fmt.Printf("failed to read file: %v", err)
		return
	}

	conv := converter.NewStore(GeneratorConfig)
	for _, subfile := range bytes.Split(data, []byte("\n---\n")) {

		decode := scheme.Codecs.UniversalDeserializer().Decode
		obj, _, err := decode(subfile, nil, nil)
		if err != nil {
			fmt.Println("Error while decoding YAML object. Err was: ", err)
			return
		}

		// now use switch over the type of the object
		// and match each type-case
		switch o := obj.(type) {
		case *corev1.Node:
			err = AddNodeToList(o)
			if err != nil {
				fmt.Println("Failed to add node to list:", err)
			}
		case *corev1.Pod:
			err = AddPodToList(o)
			if err != nil {
				fmt.Println("Failed to add pod to list:", err)
			}
			p := store.Pod{
				Name:           o.Name,
				Namespace:      o.Namespace,
				NodeName:       o.Spec.NodeName,
				ServiceAccount: o.Spec.ServiceAccountName,
				HostPID:        o.Spec.HostPID,
				HostIPC:        o.Spec.HostIPC,
				HostNetwork:    o.Spec.HostNetwork,
			}

			for _, cont := range o.Spec.Containers {
				err = AddContainerToList(&cont, &p)
				if err != nil {
					fmt.Println("Failed to add container to list:", err)
				}

				for _, vol := range cont.VolumeMounts {
					err = AddVolumeToList(&vol, &p)
					if err != nil {
						fmt.Println("Failed to add volume to list:", err)
					}
				}

			}

		case *rbacv1.Role:
			role, err := conv.Role(ctx, o)
			if err != nil {
				fmt.Println("Failed to convert role:", err)
			}
			provider.Write(ctx, role)
		case *rbacv1.ClusterRole:
			clusterRole, err := conv.ClusterRole(ctx, o)
			if err != nil {
				fmt.Println("Failed to convert role:", err)
			}
			provider.Write(ctx, clusterRole)
		case *rbacv1.ClusterRoleBinding:
			*cacheClusterRoleBinding = append(*cacheClusterRoleBinding, o)
		case *rbacv1.RoleBinding:
			*cacheRoleBinding = append(*cacheRoleBinding, o)
		default:
			fmt.Printf("(TODO) %T object has not yet been implememented: %+v", o, o)
		}
	}
}

func AddPermissionSetToList(ctx context.Context, roleBinding *store.RoleBinding, convStore *converter.StoreConverter) error {
	AddIdentityToList(roleBinding)
	ps, err := convStore.PermissionSet(ctx, roleBinding)
	if err != nil {
		return err
	}
	gps := storePermissionSetToGraph(ps)
	PermissionSets[gps.Name] = *gps
	return nil
}

func ConvertRoleBindings(ctx context.Context, cacheRoleBinding []*rbacv1.RoleBinding, cacheClusterRoleBinding []*rbacv1.ClusterRoleBinding, db *sql.DB) error {

	convStore := converter.NewStoreWithDB(GeneratorConfig, db)
	var errConvert error
	for _, rb := range cacheRoleBinding {
		roleBinding, err := convStore.RoleBinding(ctx, rb)
		if err != nil {
			errConvert = errors.Join(errConvert, err)
			fmt.Println("Failed to convert role:", err)
			continue
		}
		err = AddPermissionSetToList(ctx, roleBinding, convStore)
		if err != nil {
			errConvert = errors.Join(errConvert, err)
			fmt.Printf("Failed to add permission set[rb:%s]: %v\n", roleBinding.Name, err)
			continue
		}
	}

	for _, crb := range cacheClusterRoleBinding {
		clusterRoleBinding, err := convStore.ClusterRoleBinding(ctx, crb)
		if err != nil {
			errConvert = errors.Join(errConvert, err)
			fmt.Println("Failed to convert cluster role:", err)
			continue
		}
		err = AddPermissionSetToList(ctx, clusterRoleBinding, convStore)
		if err != nil {
			errConvert = errors.Join(errConvert, err)
			fmt.Printf("Failed to add permission set[rb:%s]:  %v\n", clusterRoleBinding.Name, err)
			continue
		}
	}

	return errConvert
}

func AddIdentityToList(rb *store.RoleBinding) error {
	convStore := converter.NewStore(GeneratorConfig)
	for _, subj := range rb.Subjects {
		sid, err := convStore.Identity(context.Background(), &subj, rb)
		if err != nil {
			return err
		}
		gi := storeIdentityToGraph(sid)
		Identities[gi.Name] = *gi
	}
	return nil
}

func AddPodToList(pod *corev1.Pod) error {
	fmt.Printf("pod name: %s\n", pod.Name)
	if pod.Namespace == "" {
		pod.Namespace = defaultNamespace
	}
	sa := pod.Spec.ServiceAccountName
	if sa == "" {
		sa = defaultServiceAccount
	}
	spn := false
	if pod.Spec.ShareProcessNamespace != nil {
		spn = *pod.Spec.ShareProcessNamespace
	}
	Pods[pod.Name] = gPod{
		StoreID:               "",
		Name:                  pod.Name,
		IsNamespaced:          pod.Namespace != "",
		Namespace:             pod.Namespace,
		ShareProcessNamespace: spn,
		ServiceAccount:        sa,
		Node:                  pod.Spec.NodeName,
		Compromised:           0,
		Critical:              false,
	}

	return nil
}

func AddNodeToList(node *corev1.Node) error {
	fmt.Printf("Node name: %s\n", node.Name)
	n := gNode{
		StoreID:      "",
		Name:         node.Name,
		IsNamespaced: node.Namespace != "",
		Namespace:    node.Namespace,
		Compromised:  0,
		Critical:     false,
	}
	fmt.Printf("Adding %+v to nodes", n)
	Nodes[n.Name] = n

	return nil
}

func AddContainerToList(container *corev1.Container, storePod *store.Pod) error {
	fmt.Printf("Container name: %s\n", container.Name)

	caps := make([]string, 0)
	if container.SecurityContext != nil && container.SecurityContext.Capabilities != nil {
		for _, cap := range container.SecurityContext.Capabilities.Add {
			caps = append(caps, string(cap))
		}
	}

	ports := make([]string, 0, len(container.Ports))
	for _, p := range container.Ports {
		ports = append(ports, strconv.Itoa(int(p.ContainerPort)))
	}

	privileged := false
	if container.SecurityContext != nil && container.SecurityContext.Privileged != nil {
		privileged = *container.SecurityContext.Privileged
	}

	privEsc := false
	if container.SecurityContext != nil && container.SecurityContext.AllowPrivilegeEscalation != nil {
		privEsc = *container.SecurityContext.AllowPrivilegeEscalation
	}

	var runAsUser int64
	if container.SecurityContext != nil && container.SecurityContext.RunAsUser != nil {
		runAsUser = *container.SecurityContext.RunAsUser
	}

	c := gContainer{
		StoreID:      "",
		Name:         container.Name,
		Image:        container.Image,
		Command:      container.Command,
		Args:         container.Args,
		Capabilities: caps,
		Privileged:   privileged,
		PrivEsc:      privEsc,
		HostPID:      storePod.HostPID,
		HostIPC:      storePod.HostIPC,
		HostNetwork:  storePod.HostNetwork,
		RunAsUser:    runAsUser,
		Ports:        ports,
		Pod:          storePod.Name,
		Namespace:    storePod.Namespace,
		Compromised:  0,
	}

	if c.Command == nil {
		c.Command = []string{}
	}
	if c.Args == nil {
		c.Args = []string{}
	}
	if len(storePod.Namespace) != 0 {
		c.IsNamespaced = true
	}

	Containers[container.Name] = c

	return nil
}

func AddVolumeToList(volume *corev1.VolumeMount, storePod *store.Pod) error {
	fmt.Printf("Volume name: %s\n", volume.Name)
	v := gVolume{
		StoreID:    "",
		Name:       volume.Name,
		MountPath:  volume.MountPath,
		Readonly:   volume.ReadOnly,
		Namespace:  storePod.Namespace,
	}
	if v.Namespace != "" {
		v.IsNamespaced = true
	}
	Volumes[volume.Name] = v

	return nil
}

// storePermissionSetToGraph converts a store.PermissionSet to a gPermissionSet.
func storePermissionSetToGraph(input *store.PermissionSet) *gPermissionSet {
	output := &gPermissionSet{
		StoreID:      store.Hex(input.Id),
		Name:         input.Name,
		Namespace:    input.Namespace,
		Role:         input.RoleName,
		RoleBinding:  input.RoleBindingName,
		Rules:        flattenPolicyRules(input.Rules),
		Critical:     risk.Engine().IsCritical(input),
	}
	if output.Namespace != "" {
		output.IsNamespaced = true
	}
	return output
}

// storeIdentityToGraph converts a store.Identity to a gIdentity.
func storeIdentityToGraph(input *store.Identity) *gIdentity {
	output := &gIdentity{
		StoreID:   store.Hex(input.Id),
		Name:      input.Name,
		Namespace: input.Namespace,
		Type:      input.Type,
		Critical:  false,
	}
	if output.Namespace != "" {
		output.IsNamespaced = true
	}
	return output
}

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

func GeneratePermissionSetTemplate() ([]byte, error) {
	tmpl := `var expectedPermissionSets = map[string]gPermissionSet{
		{{- range $val := .}}
		"{{.Name}}": {
			StoreID:      "",
			Name:         "{{.Name}}",
			IsNamespaced: {{.IsNamespaced}},
			Namespace:    "{{.Namespace}}",
			Role:         "{{.Role}}",
			Rules:        []string{ {{range $i, $rule := .Rules}}{{if $i}},{{end}}"{{$rule}}"{{end}} },
			RoleBinding:  "{{.RoleBinding}}",
			Critical:     false,
		},{{ end }}
	}
`
	t := template.Must(template.New("tmpl").Parse(tmpl))
	outbuf := bytes.NewBuffer([]byte{})
	err := t.Execute(outbuf, PermissionSets)
	if err != nil {
		fmt.Print(err)
	}
	return outbuf.Bytes(), nil
}

func GenerateIdentityTemplate() ([]byte, error) {
	tmpl := `var expectedIdentities = map[string]gIdentity{
		{{- range $val := .}}
		"{{.Name}}": {
			StoreID:      "",
			Name:         "{{.Name}}",
			IsNamespaced: {{.IsNamespaced}},
			Namespace:    "{{.Namespace}}",
			Type:         "{{.Type}}",
			Critical:     false,
		},{{ end }}
	}
`

	t := template.Must(template.New("tmpl").Parse(tmpl))
	outbuf := bytes.NewBuffer([]byte{})
	t.Execute(outbuf, Identities)

	return outbuf.Bytes(), nil
}

func GenerateNodeTemplate() ([]byte, error) {
	tmpl := `var expectedNodes = map[string]gNode{
		{{- range $val := .}}
		"{{.Name}}": {
			StoreID:      "",
			Name:         "{{.Name}}",
			IsNamespaced: {{.IsNamespaced}},
			Namespace:    "{{.Namespace}}",
			Compromised:  shared.CompromiseNone,
			Critical:     false,
		},{{ end }}
	}
`

	t := template.Must(template.New("tmpl").Parse(tmpl))
	outbuf := bytes.NewBuffer([]byte{})
	t.Execute(outbuf, Nodes)

	return outbuf.Bytes(), nil
}

func GeneratePodTemplate() ([]byte, error) {
	tmpl := `var expectedPods = map[string]gPod{
		{{- range $val := .}}
		"{{.Name}}": {
			StoreID:                 "",
			Name:                    "{{.Name}}",
			IsNamespaced:            {{.IsNamespaced}},
			Namespace:               "{{.Namespace}}",
			Compromised:             shared.CompromiseNone,
			ServiceAccount:          "{{.ServiceAccount}}",
			ShareProcessNamespace:   {{.ShareProcessNamespace}},
			Critical:                false,
		},{{ end }}
	}
`

	t := template.Must(template.New("tmpl").Parse(tmpl))
	outbuf := bytes.NewBuffer([]byte{})
	fmt.Println("seen pods total:", len(Pods))
	t.Execute(outbuf, Pods)

	return outbuf.Bytes(), nil
}

func GenerateContainerTemplate() ([]byte, error) {
	tmpl := `var expectedContainers = map[string]gContainer{
		{{- range $val := .}}
		"{{.Name}}": {
			StoreID:      "",
			Name:         "{{.Name}}",
			Image:        "{{.Image}}",
			Command:      []string{},
			Args:         []string{},
			Capabilities: []string{},
			Privileged:   {{.Privileged}},
			PrivEsc:      {{.PrivEsc}},
			HostPID:      {{.HostPID}},
			HostIPC:      {{.HostIPC}},
			HostNetwork:  {{.HostNetwork}},
			RunAsUser:    {{.RunAsUser}},
			Namespace:    "{{.Namespace}}",
			Ports:        []string{},
			Pod:          "{{.Pod}}",
			// Node:         "{{.Node}}",
			Compromised:  0,
		},{{ end }}
	}
`

	t := template.Must(template.New("tmpl").Parse(tmpl))
	outbuf := bytes.NewBuffer([]byte{})
	t.Execute(outbuf, Containers)

	return outbuf.Bytes(), nil
}

func GenerateVolumeTemplate() ([]byte, error) {
	tmpl := `var expectedVolumes = map[string]gVolume{
		{{- range $val := .}}
		"{{.Name}}": {
			StoreID:    "",
			Name:       "{{.Name}}",
			Type:       "{{.Type}}",
			SourcePath: "{{.SourcePath}}",
			MountPath:  "{{.MountPath}}",
			Readonly:   {{.Readonly}},
			Namespace:  "{{.Namespace}}",
		},{{ end }}
	}
`

	t := template.Must(template.New("tmpl").Parse(tmpl))
	outbuf := bytes.NewBuffer([]byte{})
	t.Execute(outbuf, Volumes)

	return outbuf.Bytes(), nil
}

func WriteTemplatesToFile(path string, templates ...[]byte) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	in := bytes.Join(templates, []byte("\n"))
	// We run go fmt on it so it's "clean"
	// The formatting is not as strict as our editors config & linter
	clean, err := format.Source(in)
	if err != nil {
		return err
	}
	_, err = f.Write(clean)
	if err != nil {
		return err
	}
	return nil
}
