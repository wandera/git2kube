package upload

import (
	"encoding/json"
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
	"strings"
)

const refAnnotation = "git2kube.github.com/ref"

// Uploader uploading date to K8s configmap
type Uploader interface {
	Upload(commitId string, iter *object.FileIter) error
}

type uploader struct {
	restconfig *rest.Config
	clientset  *kubernetes.Clientset
	namespace  string
	name       string
}

// NewUploader creates new Uploader
func NewUploader(configpath string, name string, namespace string) (Uploader, error) {
	restconfig, err := restConfig(configpath)
	if err != nil {
		return nil, err
	}

	clientset, err := kubernetes.NewForConfig(restconfig)
	if err != nil {
		return nil, err
	}

	return &uploader{
		restconfig: restconfig,
		clientset:  clientset,
		namespace:  namespace,
		name:       name,
	}, nil
}

func (u *uploader) Upload(commitId string, iter *object.FileIter) error {
	configMaps := u.clientset.CoreV1().ConfigMaps(u.namespace)

	data, err := iterToData(iter)
	if err != nil {
		return err
	}

	oldMap, err := configMaps.Get(u.name, metav1.GetOptions{})
	if err == nil {
		err = u.patchConfigMap(oldMap, configMaps, data, commitId)
		if err != nil {
			return err
		}
	} else {
		err = u.createConfigMap(configMaps, data, commitId)
		if err != nil {
			return err
		}
	}

	return nil
}

func (u *uploader) patchConfigMap(oldMap *corev1.ConfigMap, configMaps typedcore.ConfigMapInterface, data map[string]string, commitId string) error {
	log.Infof("Patching ConfigMap '%s.%s'", oldMap.Namespace, oldMap.Name)
	newMap := oldMap.DeepCopy()
	newMap.Data = data

	if newMap.Annotations == nil {
		newMap.Annotations = make(map[string]string)
	}
	newMap.Annotations[refAnnotation] = commitId

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

func (u *uploader) createConfigMap(configMaps typedcore.ConfigMapInterface, data map[string]string, commitId string) error {
	log.Infof("Creating ConfigMap '%s.%s'", u.namespace, u.name)
	_, err := configMaps.Create(&corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      u.name,
			Namespace: u.namespace,
			Annotations: map[string]string{
				refAnnotation: commitId,
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

func iterToData(iter *object.FileIter) (map[string]string, error) {
	var data = make(map[string]string)
	err := iter.ForEach(func(file *object.File) error {
		if !strings.HasPrefix(file.Name, ".") {
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

func restConfig(kubeconfig string) (*rest.Config, error) {
	if kubeconfig != "" {
		loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
		configOverrides := &clientcmd.ConfigOverrides{}
		kubeConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, configOverrides)
		return kubeConfig.ClientConfig()
	}
	return rest.InClusterConfig()
}
