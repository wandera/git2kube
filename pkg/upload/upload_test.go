package upload

import (
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"testing"

	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/filemode"
	"github.com/go-git/go-git/v5/plumbing/object"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	testclient "k8s.io/client-go/kubernetes/fake"
	testing2 "k8s.io/client-go/testing"
)

type mockFileIter struct {
	files []*object.File
}

func (m *mockFileIter) ForEach(cb func(*object.File) error) error {
	for _, f := range m.files {
		obj := &plumbing.MemoryObject{}
		content, err := os.ReadFile(filepath.Join("testdata", f.Name))
		if err != nil {
			panic(err)
		}
		obj.Write(content)
		obj.SetType(plumbing.BlobObject)
		blob := &object.Blob{}
		err = blob.Decode(obj)
		if err != nil {
			panic(err)
		}

		cb(object.NewFile(f.Name, f.Mode, blob))
	}
	return nil
}

var basicCases = []struct {
	name        string
	namespace   string
	target      string
	includes    []*regexp.Regexp
	excludes    []*regexp.Regexp
	labels      map[string]string
	annotations map[string]string
	iter        *mockFileIter
	contains    []string
}{
	{
		name:        "Empty JSON in default namespace no include",
		namespace:   "default",
		target:      "git2kube",
		labels:      map[string]string{},
		annotations: map[string]string{},
		iter: &mockFileIter{
			files: []*object.File{
				object.NewFile("test.json", filemode.Regular, &object.Blob{}),
			},
		},
	},
	{
		name:        "Empty JSON in default namespace include all",
		namespace:   "default",
		target:      "git2kube",
		labels:      map[string]string{},
		annotations: map[string]string{},
		includes: []*regexp.Regexp{
			regexp.MustCompile(".*"),
		},
		iter: &mockFileIter{
			files: []*object.File{
				object.NewFile("test.json", filemode.Regular, &object.Blob{}),
				object.NewFile("test.yaml", filemode.Regular, &object.Blob{}),
			},
		},
		contains: []string{"test.json", "test.yaml"},
	},
	{
		name:        "Empty JSON in default namespace include all exclude yaml",
		namespace:   "default",
		target:      "git2kube",
		labels:      map[string]string{},
		annotations: map[string]string{},
		includes: []*regexp.Regexp{
			regexp.MustCompile(".*"),
		},
		excludes: []*regexp.Regexp{
			regexp.MustCompile(`.*\.yaml`),
		},
		iter: &mockFileIter{
			files: []*object.File{
				object.NewFile("test.json", filemode.Regular, &object.Blob{}),
				object.NewFile("test.yaml", filemode.Regular, &object.Blob{}),
			},
		},
		contains: []string{"test.json"},
	},
	{
		name:        "Empty JSON in default namespace no include",
		namespace:   "default",
		target:      "git2kube",
		labels:      map[string]string{},
		annotations: map[string]string{},
		iter: &mockFileIter{
			files: []*object.File{
				object.NewFile("test.json", filemode.Regular, &object.Blob{}),
			},
		},
	},
	{
		name:        "No files in config namespace",
		namespace:   "config",
		target:      "git2kube",
		labels:      map[string]string{},
		annotations: map[string]string{},
		iter: &mockFileIter{
			files: []*object.File{},
		},
	},
	{
		name:      "No files in config namespace with annotations and labels",
		namespace: "config",
		target:    "git2kube",
		labels: map[string]string{
			"test1": "value1",
		},
		annotations: map[string]string{
			"test2": "value2",
		},
		iter: &mockFileIter{
			files: []*object.File{},
		},
	},
}

func TestConfigmapUploader_Upload(t *testing.T) {
	for _, c := range basicCases {
		fakeclient := testclient.NewSimpleClientset()
		cu := &configmapUploader{
			clientset:   fakeclient,
			namespace:   c.namespace,
			name:        c.target,
			labels:      c.labels,
			annotations: c.annotations,
			includes:    c.includes,
			excludes:    c.excludes,
		}
		err := cu.Upload("id", c.iter)
		if err != nil {
			t.Errorf("%s case failed: %v", c.name, err)
		}

		assertAction(fakeclient.Actions()[0], t, c.name, c.namespace, "get", "configmaps")
		assertAction(fakeclient.Actions()[1], t, c.name, c.namespace, "create", "configmaps")
		res, err := fakeclient.CoreV1().ConfigMaps(c.namespace).Get(c.target, v1.GetOptions{})
		if err != nil {
			t.Errorf("%s case failed: %v", c.name, err)
		}
		assertAnnotationsAndLabels(res.Annotations, res.Labels, t, c.name, c.annotations, c.labels)
		assertData(res.Data, t, c.name, c.contains)
	}
}

func TestSecretUploader_Upload(t *testing.T) {
	for _, c := range basicCases {
		fakeclient := testclient.NewSimpleClientset()
		cu := &secretUploader{
			clientset:   fakeclient,
			namespace:   c.namespace,
			name:        c.target,
			labels:      c.labels,
			annotations: c.annotations,
			includes:    c.includes,
			excludes:    c.excludes,
		}
		err := cu.Upload("id", c.iter)
		if err != nil {
			t.Errorf("%s case failed: %v", c.name, err)
		}

		assertAction(fakeclient.Actions()[0], t, c.name, c.namespace, "get", "secrets")
		assertAction(fakeclient.Actions()[1], t, c.name, c.namespace, "create", "secrets")

		res, err := fakeclient.CoreV1().Secrets(c.namespace).Get(c.target, v1.GetOptions{})
		if err != nil {
			t.Errorf("%s case failed: %v", c.name, err)
		}
		assertAnnotationsAndLabels(res.Annotations, res.Labels, t, c.name, c.annotations, c.labels)

		data := make(map[string]string)
		for k, v := range res.Data {
			data[k] = string(v[:])
		}
		assertData(data, t, c.name, c.contains)
	}
}

func TestFolderUploader_Upload(t *testing.T) {
	ex, err := os.Executable()
	if err != nil {
		panic(err)
	}
	exPath := filepath.Dir(ex)

	for _, c := range basicCases {
		cu := &folderUploader{
			sourcePath: exPath,
			name:       c.target,
			includes:   c.includes,
			excludes:   c.excludes,
		}
		err := cu.Upload("id", c.iter)
		if err != nil {
			t.Errorf("%s case failed: %v", c.name, err)
		}
	}
}

func assertAction(action testing2.Action, t *testing.T, name string, namespace string, verb string, resource string) {
	if action.GetNamespace() != namespace {
		t.Errorf("%s case failed: expected '%s' namespace but got '%s' instead", name, namespace, action.GetNamespace())
	}
	if !action.Matches(verb, resource) {
		t.Errorf("%s case failed: expected action '[%s]%s' namespace but got '[%s]%s' instead", name, verb, resource, action.GetVerb(), action.GetResource().Resource)
	}
}

func assertAnnotationsAndLabels(annotations map[string]string, labels map[string]string, t *testing.T, name string, exannotations map[string]string, exlabels map[string]string) {
	if !reflect.DeepEqual(annotations, exannotations) {
		t.Errorf("%s case failed: expected annotations '%s' but got '%s' instead", name, exannotations, annotations)
	}

	if !reflect.DeepEqual(labels, exlabels) {
		t.Errorf("%s case failed: expected labels '%s' but got '%s' instead", name, exlabels, labels)
	}
}

func assertData(data map[string]string, t *testing.T, name string, contains []string) {
	if len(contains) != len(data) {
		t.Errorf("%s case failed: expected data '%s' but got '%s' instead", name, contains, reflect.ValueOf(data).MapKeys())
	}

	for _, k := range contains {
		if v, ok := data[k]; ok {
			content, _ := os.ReadFile(filepath.Join("testdata", k))
			if v != string(content) {
				t.Errorf("%s case failed: content mismatch expected '%s' but got '%s' instead", name, content, v)
			}
		} else {
			t.Errorf("%s case failed: expected data with key '%s' in '%s'", name, k, data)
		}
	}
}
