package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"text/template"

	"github.com/DataDog/KubeHound/pkg/kubehound/models/converter"
	"github.com/DataDog/KubeHound/pkg/kubehound/models/graph"
	"github.com/DataDog/KubeHound/pkg/kubehound/models/store"
	"gopkg.in/yaml.v3"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/api/rbac/v1beta1"

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

var (
	Containers = make(map[string]graph.Container)
	Pods       = make(map[string]graph.Pod)
	Nodes      = make(map[string]graph.Node)
	Roles      = make(map[string]graph.Role)
	Volumes    = make(map[string]graph.Volume)
)

var (
	globalHeaders = []byte(`
// PLEASE DO NOT EDIT
// THIS HAS BEEN GENERATED AUTOMATICALLY
// 
// FIXME:
// cd test/system/generator && go build && ./generator ../../setup/test-cluster/

package system

import (
	"github.com/DataDog/KubeHound/pkg/kubehound/models/graph"
	"github.com/DataDog/KubeHound/pkg/kubehound/models/shared"
)

`)
)

func main() {
	// todo: check length
	path := os.Args[1]

	clusterFile, err := ioutil.ReadFile(filepath.Join(path, "cluster.yaml"))
	if err != nil {
		log.Fatal(err)
	}
	err = ProcessCluster(clusterFile)
	if err != nil {
		log.Fatal(err)
	}

	attackPath := filepath.Join(path, "attacks")
	filesAttack, err := ioutil.ReadDir(attackPath)
	if err != nil {
		log.Fatal(err)
	}
	for _, file := range filesAttack {
		ProcessFile(attackPath, file)
	}

	fmt.Printf("Pod stored: %+v\n", Pods)
	fmt.Printf("Node stored: %+v\n", Nodes)
	fmt.Printf("Volume stored: %+v\n", Volumes)
	outPods, err := GeneratePodTemplate()
	if err != nil {
		fmt.Println("failed to write pods: ", err)
	}
	outNodes, err := GenerateNodeTemplate()
	if err != nil {
		fmt.Println("failed to write pods: ", err)
	}
	WriteTemplatesToFile("outfile.gen.go", globalHeaders, outPods, outNodes)
}

func ProcessCluster(content []byte) error {
	var cluster Cluster
	err := yaml.Unmarshal(content, &cluster)
	if err != nil {
		return err
	}
	for _, n := range cluster.Nodes {
		node := n.Role
		Nodes[node] = graph.Node{
			StoreID:      "",
			Name:         node,
			IsNamespaced: false,
			Namespace:    "",
			Compromised:  0,
			Critical:     false,
		}
	}
	return nil
}

func ProcessFile(basePath string, file os.FileInfo) {
	fmt.Println(file.Name(), file.IsDir())
	data, err := os.ReadFile(filepath.Join(basePath, file.Name()))
	if err != nil {
		fmt.Printf("failed to read file: %v", err)
		return
	}
	decode := scheme.Codecs.UniversalDeserializer().Decode
	obj, _, err := decode(data, nil, nil)
	if err != nil {
		fmt.Println("Error while decoding YAML object. Err was: ", err)
		return
	}

	// now use switch over the type of the object
	// and match each type-case
	switch o := obj.(type) {
	case *v1.Node:
		AddNodeToList(o)
	case *v1.Pod:
		AddPodToList(o)
		for _, vol := range o.Spec.Volumes {
			p := store.Pod{
				K8: *o,
			}
			AddVolumeToList(&vol, &p)
		}
	case *v1beta1.Role:
		// TODO
	case *v1beta1.RoleBinding:
		// TODO
	case *v1beta1.ClusterRole:
		// TODO
	case *v1beta1.ClusterRoleBinding:
		// TODO
	case *v1.ServiceAccount:
	default:
		fmt.Printf("Unknown object type: %+v\n", o)
		//o is unknown for us
	}
}

func AddPodToList(pod *corev1.Pod) error {
	fmt.Printf("pod name: %s\n", pod.Name)
	storePod := store.Pod{
		K8: *pod,
	}
	conv := converter.GraphConverter{}
	convertedPod, _ := conv.Pod(&storePod)
	Pods[pod.Name] = *convertedPod

	return nil
}

func AddNodeToList(node *corev1.Node) error {
	fmt.Printf("Node name: %s\n", node.Name)
	storeNode := store.Node{
		K8: *node,
	}
	conv := converter.GraphConverter{}
	convertedNode, _ := conv.Node(&storeNode)
	Nodes[node.Name] = *convertedNode

	return nil
}

func AddVolumeToList(volume *corev1.Volume, storePod *store.Pod) error {
	fmt.Printf("Volume name: %s\n", volume.Name)
	storeVolume := store.Volume{
		Source: *volume,
	}
	conv := converter.GraphConverter{}
	convertedVolume, _ := conv.Volume(&storeVolume, storePod)
	Volumes[volume.Name] = *convertedVolume

	return nil
}

func GenerateNodeTemplate() ([]byte, error) {
	tmpl := `var expectedNodeNames = map[string]graph.Node{
		{{range $val := . -}}
		"{{.Name}}": {
			StoreID:      "",
			Name:         "{{.Name}}",
			IsNamespaced: {{.IsNamespaced}},
			Namespace:    "{{.Namespace}}",
			Compromised:  shared.CompromiseNone,
			Critical:     false,
		},
		{{- end}}
	}
`

	t := template.Must(template.New("tmpl").Parse(tmpl))
	outbuf := bytes.NewBuffer([]byte{})
	t.Execute(outbuf, Nodes)

	return outbuf.Bytes(), nil
}

func GeneratePodTemplate() ([]byte, error) {
	tmpl := `var expectedPodNames = map[string]graph.Pod{
		{{range $val := . -}}
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
	fmt.Println("seen pods total:", len(Pods))
	t.Execute(outbuf, Pods)

	return outbuf.Bytes(), nil
}

func GenerateContainerTemplate() ([]byte, error) {
	tmpl := `var expectedContainerNames = map[string]graph.Container{
		{{range $val := .}}
		"{{.Name}}": {
			StoreID:      "",
			Name:         "{{.Name}}",
			IsNamespaced: {{.IsNamespaced}},
			Namespace:    "{{.Namespace}}",
			Compromised:  shared.CompromiseNone,
			Critical:     false,
		},
		{{end}}
	}
`

	t := template.Must(template.New("tmpl").Parse(tmpl))
	outbuf := bytes.NewBuffer([]byte{})
	t.Execute(outbuf, Nodes)

	return outbuf.Bytes(), nil
}

func WriteTemplatesToFile(path string, templates ...[]byte) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}

	for _, t := range templates {
		_, err := f.Write(t)
		if err != nil {
			return err
		}
	}
	return nil
}
