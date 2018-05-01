package upload

import (
	"encoding/json"
	"fmt"
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
	"encoding/base64"
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
	restconfig  *rest.Config
	clientset   *kubernetes.Clientset
	namespace   string
	name        string
	mergeType   MergeType
	labels      map[string]string
	annotations map[string]string
	includes    []*regexp.Regexp
	excludes    []*regexp.Regexp
}

type configmapUploader uploader
type secretUploader uploader

// UploaderOptions uploader options
type UploaderOptions struct {
	Kubeconfig    bool
	ConfigMapName string
	Namespace     string
	MergeType     MergeType
	Includes      []string
	Excludes      []string
	Labels        []string
	Annotations   []string
}

// NewConfigMapUploader creates new ConfigMapUploader
func NewConfigMapUploader(o *UploaderOptions) (Uploader, error) {
	restconfig, err := restConfig(o.Kubeconfig)
	if err != nil {
		return nil, err
	}

	clientset, err := kubernetes.NewForConfig(restconfig)
	if err != nil {
		return nil, err
	}

	includesRegex, err := stringsToRegExp(o.Includes)
	if err != nil {
		return nil, err
	}
	log.Infof("Loaded include rules %s", includesRegex)

	excludesRegex, err := stringsToRegExp(o.Excludes)
	if err != nil {
		return nil, err
	}
	log.Infof("Loaded exclude rules %s", excludesRegex)

	labelsParsed, err := stringsToMap(o.Labels)
	if err != nil {
		return nil, err
	}

	annotationsParsed, err := stringsToMap(o.Annotations)
	if err != nil {
		return nil, err
	}

	return &configmapUploader{
		mergeType:   o.MergeType,
		includes:    includesRegex,
		excludes:    excludesRegex,
		labels:      labelsParsed,
		annotations: annotationsParsed,
		restconfig:  restconfig,
		clientset:   clientset,
		namespace:   o.Namespace,
		name:        o.ConfigMapName,
	}, nil
}

func (u *configmapUploader) Upload(commitID string, iter *object.FileIter) error {
	configMaps := u.clientset.CoreV1().ConfigMaps(u.namespace)

	data, err := u.iterToConfigMapData(iter)
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

func (u *configmapUploader) patchConfigMap(oldMap *corev1.ConfigMap, configMaps typedcore.ConfigMapInterface, data map[string]string, commitID string) error {
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

	if err := mergo.Merge(&newMap.Annotations, u.annotations, mergo.WithOverride); err != nil {
		if err != nil {
			return err
		}
	}
	newMap.Annotations[refAnnotation] = commitID

	if err := mergo.Merge(&newMap.Labels, u.labels, mergo.WithOverride); err != nil {
		if err != nil {
			return err
		}
	}

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

func (u *configmapUploader) createConfigMap(configMaps typedcore.ConfigMapInterface, data map[string]string, commitID string) error {
	log.Infof("Creating ConfigMap '%s.%s'", u.namespace, u.name)

	annotations := u.annotations
	annotations[refAnnotation] = commitID

	_, err := configMaps.Create(&corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:        u.name,
			Namespace:   u.namespace,
			Annotations: annotations,
			Labels:      u.labels,
		},
		Data: data,
	})
	if err != nil {
		return err
	}

	log.Infof("Successfully created ConfigMap '%s.%s'", u.namespace, u.name)
	return nil
}

func (u *configmapUploader) iterToConfigMapData(iter *object.FileIter) (map[string]string, error) {
	var data = make(map[string]string)
	err := iter.ForEach(func(file *object.File) error {
		if filterFile(file, u.includes, u.excludes) {
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

// NewConfigMapUploader creates new SecretUploader
func NewSecretUploader(o *UploaderOptions) (Uploader, error) {
	restconfig, err := restConfig(o.Kubeconfig)
	if err != nil {
		return nil, err
	}

	clientset, err := kubernetes.NewForConfig(restconfig)
	if err != nil {
		return nil, err
	}

	includesRegex, err := stringsToRegExp(o.Includes)
	if err != nil {
		return nil, err
	}
	log.Infof("Loaded include rules %s", includesRegex)

	excludesRegex, err := stringsToRegExp(o.Excludes)
	if err != nil {
		return nil, err
	}
	log.Infof("Loaded exclude rules %s", excludesRegex)

	labelsParsed, err := stringsToMap(o.Labels)
	if err != nil {
		return nil, err
	}

	annotationsParsed, err := stringsToMap(o.Annotations)
	if err != nil {
		return nil, err
	}

	return &configmapUploader{
		mergeType:   o.MergeType,
		includes:    includesRegex,
		excludes:    excludesRegex,
		labels:      labelsParsed,
		annotations: annotationsParsed,
		restconfig:  restconfig,
		clientset:   clientset,
		namespace:   o.Namespace,
		name:        o.ConfigMapName,
	}, nil
}

func (u *secretUploader) Upload(commitID string, iter *object.FileIter) error {
	secrets := u.clientset.CoreV1().Secrets(u.namespace)

	data, err := u.iterToSecretData(iter)
	if err != nil {
		return err
	}

	oldSecret, err := secrets.Get(u.name, metav1.GetOptions{})
	if err == nil {
		err = u.patchSecret(oldSecret, secrets, data, commitID)
		if err != nil {
			return err
		}
	} else {
		err = u.createSecret(secrets, data, commitID)
		if err != nil {
			return err
		}
	}

	return nil
}

func (u *secretUploader) patchSecret(oldSecret *corev1.Secret, secrets typedcore.SecretInterface, data map[string][]byte, commitID string) error {
	log.Infof("Patching Secret '%s.%s'", oldSecret.Namespace, oldSecret.Name)
	newSecret := oldSecret.DeepCopy()

	switch u.mergeType {
	case Delete:
		newSecret.Data = data
	case Upsert:
		if err := mergo.Merge(&newSecret.Data, data, mergo.WithOverride); err != nil {
			if err != nil {
				return err
			}
		}
	}

	if err := mergo.Merge(&newSecret.Annotations, u.annotations, mergo.WithOverride); err != nil {
		if err != nil {
			return err
		}
	}
	newSecret.Annotations[refAnnotation] = commitID

	if err := mergo.Merge(&newSecret.Labels, u.labels, mergo.WithOverride); err != nil {
		if err != nil {
			return err
		}
	}

	oldData, err := json.Marshal(oldSecret)
	if err != nil {
		return err
	}

	newData, err := json.Marshal(newSecret)
	if err != nil {
		return err
	}

	patchBytes, err := strategicpatch.CreateTwoWayMergePatch(oldData, newData, corev1.ConfigMap{})
	if err != nil {
		return err
	}

	_, err = secrets.Patch(u.name, types.StrategicMergePatchType, patchBytes)
	if err != nil {
		return err
	}

	log.Infof("Successfully patched ConfigMap '%s.%s'", oldSecret.Namespace, oldSecret.Name)
	return nil
}

func (u *secretUploader) createSecret(secrets typedcore.SecretInterface, data map[string][]byte, commitID string) error {
	log.Infof("Creating ConfigMap '%s.%s'", u.namespace, u.name)

	annotations := u.annotations
	annotations[refAnnotation] = commitID

	_, err := secrets.Create(&corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:        u.name,
			Namespace:   u.namespace,
			Annotations: annotations,
			Labels:      u.labels,
		},
		Data: data,
	})
	if err != nil {
		return err
	}

	log.Infof("Successfully created ConfigMap '%s.%s'", u.namespace, u.name)
	return nil
}

func (u *secretUploader) iterToSecretData(iter *object.FileIter) (map[string][]byte, error) {
	var data = make(map[string][]byte)
	err := iter.ForEach(func(file *object.File) error {
		if filterFile(file, u.includes, u.excludes) {
			content, err := file.Contents()
			if err != nil {
				return err
			}
			src := []byte(content)
			buf := make([]byte, base64.StdEncoding.EncodedLen(len(src)))
			base64.StdEncoding.Encode(buf, src)

			data[strings.Replace(file.Name, "/", ".", -1)] = buf
		}
		return nil
	})

	return data, err
}

func filterFile(file *object.File, includes []*regexp.Regexp, excludes []*regexp.Regexp) bool {
	pass := false
	for _, inc := range includes {
		if inc.MatchString(file.Name) {
			pass = true
			break
		}
	}

	for _, exc := range excludes {
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

func stringsToRegExp(strs []string) ([]*regexp.Regexp, error) {
	result := make([]*regexp.Regexp, len(strs))
	for i, str := range strs {
		regex, err := regexp.Compile(str)
		if err != nil {
			return nil, err
		}
		result[i] = regex
	}

	return result, nil
}

func stringsToMap(strs []string) (map[string]string, error) {
	result := make(map[string]string)
	for _, str := range strs {
		if !strings.Contains(str, "=") {
			return nil, fmt.Errorf("argument '%s' does not contain required char '='", str)
		}
		split := strings.Split(str, "=")
		result[split[0]] = split[1]
	}

	return result, nil
}
