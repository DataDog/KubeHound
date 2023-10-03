//nolint:all
package main

import (
	"bytes"
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
	"gopkg.in/yaml.v3"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"

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

var (
	Containers = make(map[string]graph.Container)
	Pods       = make(map[string]graph.Pod)
	Nodes      = make(map[string]graph.Node)
	Roles      = make(map[string]graph.PermissionSet)
	Volumes    = make(map[string]graph.Volume)
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
		ProcessFile(attackPath, file)
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
	err = WriteTemplatesToFile(codegenPath, globalHeaders, outPods, outNodes, outVolumes, outContainers)
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

func ProcessFile(basePath string, file os.FileInfo) {
	fmt.Println("Processing: " + file.Name())
	data, err := os.ReadFile(filepath.Join(basePath, file.Name()))
	if err != nil {
		fmt.Printf("failed to read file: %v", err)
		return
	}
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
		case *v1.Node:
			err = AddNodeToList(o)
			if err != nil {
				fmt.Println("Failed to add node to list:", err)
			}
		case *v1.Pod:
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
		//TODO:
		// case *v1beta1.Role, *v1beta1.RoleBinding, *v1beta1.ClusterRole, *v1beta1.ClusterRoleBinding:
		default:
			fmt.Printf("(TODO) %T object has not yet been implememented: %+v", o, o)
		}
	}
}

func AddPodToList(pod *corev1.Pod) error {
	fmt.Printf("pod name: %s\n", pod.Name)
	pod.Namespace = defaultNamespace
	storePod := store.Pod{
		K8: *pod,
	}
	conv := converter.GraphConverter{}
	convertedPod, err := conv.Pod(&storePod)
	// if we haven't defined the service account in the yaml file, k8s will do it for us.
	if convertedPod.ServiceAccount == "" {
		convertedPod.ServiceAccount = defaultServiceAccount
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
			StoreID:      "",
			Name:         "{{.Name}}",
			IsNamespaced: {{.IsNamespaced}},
			Namespace:    "{{.Namespace}}",
			Compromised:  shared.CompromiseNone,
			ServiceAccount: "{{.ServiceAccount}}",
			ShareProcessNamespace: {{.ShareProcessNamespace}},
			Critical:     false,
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
			StoreID: 	"",
			Name:    	"{{.Name}}",
			Type:    	"{{.Type}}",
			SourcePath: "{{.SourcePath}}",
			MountPath:  "{{.MountPath}}",
			Readonly:   {{.Readonly}},
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
