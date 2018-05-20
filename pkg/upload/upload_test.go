package upload

import (
	"testing"
	testclient "k8s.io/client-go/kubernetes/fake"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
	"gopkg.in/src-d/go-git.v4/plumbing/filemode"
)

type mockFileIter struct {
	files []*object.File
}

func (m *mockFileIter) ForEach(cb func(*object.File) error) error {
	for _, f := range m.files {
		cb(f)
	}
	return nil
}

func TestConfigmapUploader_Upload(t *testing.T) {
	cases := []struct {
		name        string
		namespace   string
		mapname     string
		labels      map[string]string
		annotations map[string]string
		iter        *mockFileIter
	}{
		{
			name:        "Empty JSON in default namespace",
			namespace:   "default",
			mapname:     "git2kube",
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
			mapname:     "git2kube",
			labels:      map[string]string{},
			annotations: map[string]string{},
			iter: &mockFileIter{
				files: []*object.File{},
			},
		},
	}

	for _, c := range cases {
		fakeclient := testclient.NewSimpleClientset()
		cu := &configmapUploader{
			clientset:   fakeclient,
			namespace:   c.namespace,
			name:        c.mapname,
			labels:      c.labels,
			annotations: c.annotations,
		}
		err := cu.Upload("id", c.iter)
		if err != nil {
			t.Errorf("%s case failed: %v", c.name, err)
		}

		firsta := fakeclient.Actions()[0]
		if firsta.GetNamespace() != c.namespace {
			t.Errorf("%s case failed expected %s namespace but got %s instead", c.name, c.namespace, firsta.GetNamespace())
		}
	}
}

func TestSecretUploader_Upload(t *testing.T) {

}

func TestFolderUploader_Upload(t *testing.T) {

}
