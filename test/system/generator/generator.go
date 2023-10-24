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
	"text/template"
	"time"

	"github.com/DataDog/KubeHound/pkg/kubehound/models/converter"
	"github.com/DataDog/KubeHound/pkg/kubehound/models/graph"
	"github.com/DataDog/KubeHound/pkg/kubehound/models/store"
	"github.com/DataDog/KubeHound/pkg/kubehound/storage/cache"
	"github.com/DataDog/KubeHound/pkg/kubehound/storage/cache/cachekey"
	"gopkg.in/yaml.v3"
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
	defaultNamespace = "default"
)

var (
	Containers     = make(map[string]graph.Container)
	Pods           = make(map[string]graph.Pod)
	Nodes          = make(map[string]graph.Node)
	PermissionSets = make(map[string]graph.PermissionSet)
	Identities     = make(map[string]graph.Identity)
	Volumes        = make(map[string]graph.Volume)
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
    "github.com/DataDog/KubeHound/pkg/kubehound/models/graph"
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
	cp, err := cache.Factory(ctx, nil)
	if err != nil {
		fmt.Printf("cache client creation: %v", err)
		return
	}
	defer cp.Close(ctx)

	cacheRole, err := cp.BulkWriter(ctx, cache.WithExpectedOverwrite())
	if err != nil {
		fmt.Printf("cache bulk writer: %v", err)
		return
	}
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
		ProcessFile(ctx, attackPath, file, cacheRole, &cacheRoleBinding, &cacheClusterRoleBinding)
	}

	// Generate permissionsets
	ConvertRoleBindings(ctx, cacheRoleBinding, cacheClusterRoleBinding, cp)
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

		Nodes[nodeName] = graph.Node{
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

func ProcessFile(ctx context.Context, basePath string, file os.FileInfo, cacheRole cache.AsyncWriter, cacheRoleBinding *[]*rbacv1.RoleBinding, cacheClusterRoleBinding *[]*rbacv1.ClusterRoleBinding) {
	fmt.Println("Processing: " + file.Name())
	data, err := os.ReadFile(filepath.Join(basePath, file.Name()))
	if err != nil {
		fmt.Printf("failed to read file: %v", err)
		return
	}

	conv := converter.StoreConverter{}
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
				K8: *o,
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
			cacheRole.Queue(ctx, cachekey.Role(role.Name, role.Namespace), *role)
		case *rbacv1.ClusterRole:
			clusterRole, err := conv.ClusterRole(ctx, o)
			if err != nil {
				fmt.Println("Failed to convert role:", err)
			}
			cacheRole.Queue(ctx, cachekey.Role(clusterRole.Name, clusterRole.Namespace), *clusterRole)
		case *rbacv1.ClusterRoleBinding:
			*cacheClusterRoleBinding = append(*cacheClusterRoleBinding, o)
		case *rbacv1.RoleBinding:
			*cacheRoleBinding = append(*cacheRoleBinding, o)
		default:
			fmt.Printf("(TODO) %T object has not yet been implememented: %+v", o, o)
		}
	}
}

func AddPermissionSetToList(ctx context.Context, roleBinding *store.RoleBinding, convStore *converter.StoreConverter, convGraph *converter.GraphConverter) error {
	AddIdentityToList(roleBinding)
	permissionSetStore, err := convStore.PermissionSet(ctx, roleBinding)
	if err != nil {
		return err
	}
	permissionSetGraph, err := convGraph.PermissionSet(permissionSetStore)
	PermissionSets[permissionSetGraph.Name] = *permissionSetGraph
	if err != nil {
		return err
	}
	return nil
}

func ConvertRoleBindings(ctx context.Context, cacheRoleBinding []*rbacv1.RoleBinding, cacheClusterRoleBinding []*rbacv1.ClusterRoleBinding, cp cache.CacheReader) error {

	convStore := converter.NewStoreWithCache(cp)
	convGraph := &converter.GraphConverter{}
	var errConvert error
	for _, rb := range cacheRoleBinding {
		roleBinding, err := convStore.RoleBinding(ctx, rb)
		if err != nil {
			errConvert = errors.Join(errConvert, err)
			fmt.Println("Failed to convert role:", err)
			continue
		}
		err = AddPermissionSetToList(ctx, roleBinding, convStore, convGraph)
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
		err = AddPermissionSetToList(ctx, clusterRoleBinding, convStore, convGraph)
		if err != nil {
			errConvert = errors.Join(errConvert, err)
			fmt.Printf("Failed to add permission set[rb:%s]:  %v\n", clusterRoleBinding.Name, err)
			continue
		}
	}

	return errConvert
}

func AddIdentityToList(rb *store.RoleBinding) error {
	convStore := converter.StoreConverter{}
	convGraph := converter.GraphConverter{}
	for _, subj := range rb.Subjects {
		sid, err := convStore.Identity(context.Background(), &subj, rb)
		if err != nil {
			return err
		}
		// Transform store model to vertex input
		insert, err := convGraph.Identity(sid)
		if err != nil {
			return err
		}
		Identities[insert.Name] = *insert
	}
	return nil
}

func AddPodToList(pod *corev1.Pod) error {
	fmt.Printf("pod name: %s\n", pod.Name)
	if pod.Namespace == "" {
		pod.Namespace = defaultNamespace
	}
	storePod := store.Pod{
		K8: *pod,
	}
	conv := converter.GraphConverter{}
	convertedPod, err := conv.Pod(&storePod)
	// if we haven't defined the service account in the yaml file, k8s will do it for us.
	if convertedPod.ServiceAccount == "" {
		convertedPod.ServiceAccount = pod.Namespace
	}
	if err != nil {
		return err
	}
	Pods[pod.Name] = *convertedPod

	return nil
}

func AddNodeToList(node *corev1.Node) error {
	fmt.Printf("Node name: %s\n", node.Name)
	storeNode := store.Node{
		K8: *node,
	}
	conv := converter.GraphConverter{}
	convertedNode, err := conv.Node(&storeNode)
	if err != nil {
		return err
	}
	fmt.Printf("Adding %+v to nodes", convertedNode)
	Nodes[convertedNode.Name] = *convertedNode

	return nil
}

func AddContainerToList(Container *corev1.Container, storePod *store.Pod) error {
	fmt.Printf("Container name: %s\n", Container.Name)
	storeContainer := store.Container{
		K8: *Container,
	}
	conv := converter.GraphConverter{}
	convertedContainer, err := conv.Container(&storeContainer, storePod)
	if err != nil {
		return err
	}
	convertedContainer.Pod = storePod.K8.Name
	Containers[Container.Name] = *convertedContainer

	return nil
}

func AddVolumeToList(volume *corev1.VolumeMount, storePod *store.Pod) error {
	fmt.Printf("Volume name: %s\n", volume.Name)
	storeVolume := store.Volume{
		Name:      volume.Name,
		MountPath: volume.MountPath,
		ReadOnly:  volume.ReadOnly,
	}
	conv := converter.GraphConverter{}
	convertedVolume, err := conv.Volume(&storeVolume, storePod)
	if err != nil {
		return err
	}
	Volumes[volume.Name] = *convertedVolume

	return nil
}

func GeneratePermissionSetTemplate() ([]byte, error) {
	tmpl := `var expectedPermissionSets = map[string]graph.PermissionSet{
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
	tmpl := `var expectedIdentities = map[string]graph.Identity{
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
	tmpl := `var expectedNodes = map[string]graph.Node{
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
	tmpl := `var expectedPods = map[string]graph.Pod{
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
	tmpl := `var expectedContainers = map[string]graph.Container{
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
	tmpl := `var expectedVolumes = map[string]graph.Volume{
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
