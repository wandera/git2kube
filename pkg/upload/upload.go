package upload

import (
	"encoding/json"
	"github.com/imdario/mergo"
	log "github.com/sirupsen/logrus"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/strategicpatch"
	"k8s.io/client-go/kubernetes"
	typedcore "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"regexp"
	"strings"
)

const refAnnotation = "git2kube.github.com/ref"

// MergeType how to merge ConfigMap data
type MergeType string

const (
	// Delete merge all keys (files) including removal of missing keys
	Delete MergeType = "delete"
	// Upsert merge all keys (files) but don't remove missing keys from the repository
	Upsert MergeType = "upsert"
)

// Uploader uploading date to K8s configmap
type Uploader interface {
	// Upload files into config map tagged by commitID
	Upload(commitID string, iter *object.FileIter) error
}

type uploader struct {
	restconfig *rest.Config
	clientset  *kubernetes.Clientset
	namespace  string
	name       string
	mergeType  MergeType
	include    []*regexp.Regexp
	exclude    []*regexp.Regexp
}

// NewUploader creates new Uploader
func NewUploader(kubeconfig bool, name string, namespace string, mergeType MergeType, include []string, exclude []string) (Uploader, error) {
	restconfig, err := restConfig(kubeconfig)
	if err != nil {
		return nil, err
	}

	clientset, err := kubernetes.NewForConfig(restconfig)
	if err != nil {
		return nil, err
	}

	includeRegex, err := stringsToRegExp(include)
	if err != nil {
		return nil, err
	}
	log.Infof("Loaded include rules %s", includeRegex)

	excludeRegex, err := stringsToRegExp(exclude)
	if err != nil {
		return nil, err
	}
	log.Infof("Loaded exclude rules %s", excludeRegex)

	return &uploader{
		mergeType:  mergeType,
		include:    includeRegex,
		exclude:    excludeRegex,
		restconfig: restconfig,
		clientset:  clientset,
		namespace:  namespace,
		name:       name,
	}, nil
}

func (u *uploader) Upload(commitID string, iter *object.FileIter) error {
	configMaps := u.clientset.CoreV1().ConfigMaps(u.namespace)

	data, err := u.iterToData(iter)
	if err != nil {
		return err
	}

	oldMap, err := configMaps.Get(u.name, metav1.GetOptions{})
	if err == nil {
		err = u.patchConfigMap(oldMap, configMaps, data, commitID)
		if err != nil {
			return err
		}
	} else {
		err = u.createConfigMap(configMaps, data, commitID)
		if err != nil {
			return err
		}
	}

	return nil
}

func (u *uploader) patchConfigMap(oldMap *corev1.ConfigMap, configMaps typedcore.ConfigMapInterface, data map[string]string, commitID string) error {
	log.Infof("Patching ConfigMap '%s.%s'", oldMap.Namespace, oldMap.Name)
	newMap := oldMap.DeepCopy()

	switch u.mergeType {
	case Delete:
		newMap.Data = data
	case Upsert:
		if err := mergo.Merge(&newMap.Data, data, mergo.WithOverride); err != nil {
			if err != nil {
				return err
			}
		}
	}

	if newMap.Annotations == nil {
		newMap.Annotations = make(map[string]string)
	}
	newMap.Annotations[refAnnotation] = commitID

	oldData, err := json.Marshal(oldMap)
	if err != nil {
		return err
	}

	newData, err := json.Marshal(newMap)
	if err != nil {
		return err
	}

	patchBytes, err := strategicpatch.CreateTwoWayMergePatch(oldData, newData, corev1.ConfigMap{})
	if err != nil {
		return err
	}

	_, err = configMaps.Patch(u.name, types.StrategicMergePatchType, patchBytes)
	if err != nil {
		return err
	}

	log.Infof("Successfully patched ConfigMap '%s.%s'", oldMap.Namespace, oldMap.Name)
	return nil
}

func (u *uploader) createConfigMap(configMaps typedcore.ConfigMapInterface, data map[string]string, commitID string) error {
	log.Infof("Creating ConfigMap '%s.%s'", u.namespace, u.name)
	_, err := configMaps.Create(&corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      u.name,
			Namespace: u.namespace,
			Annotations: map[string]string{
				refAnnotation: commitID,
			},
		},
		Data: data,
	})
	if err != nil {
		return err
	}

	log.Infof("Successfully created ConfigMap '%s.%s'", u.namespace, u.name)
	return nil
}

func (u *uploader) iterToData(iter *object.FileIter) (map[string]string, error) {
	var data = make(map[string]string)
	err := iter.ForEach(func(file *object.File) error {
		if u.filterFile(file) {
			content, err := file.Contents()
			if err != nil {
				return err
			}
			data[strings.Replace(file.Name, "/", ".", -1)] = content
		}
		return nil
	})

	return data, err
}

func (u *uploader) filterFile(file *object.File) bool {
	pass := false
	for _, inc := range u.include {
		if inc.MatchString(file.Name) {
			pass = true
			break
		}
	}

	for _, exc := range u.exclude {
		if exc.MatchString(file.Name) {
			pass = false
			break
		}
	}

	log.Debugf("[%t] '%s'", pass, file.Name)
	return pass
}

func restConfig(kubeconfig bool) (*rest.Config, error) {
	if kubeconfig {
		log.Infof("Loading kubeconfig")
		loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
		configOverrides := &clientcmd.ConfigOverrides{}
		kubeConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, configOverrides)
		return kubeConfig.ClientConfig()
	}
	log.Infof("Loading InCluster config")
	return rest.InClusterConfig()
}

func stringsToRegExp(strings []string) ([]*regexp.Regexp, error) {
	result := make([]*regexp.Regexp, len(strings))
	for i, str := range strings {
		regex, err := regexp.Compile(str)
		if err != nil {
			return nil, err
		}
		result[i] = regex
	}

	return result, nil
}
