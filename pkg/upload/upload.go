package upload

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/imdario/mergo"
	log "github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/strategicpatch"
	"k8s.io/client-go/kubernetes"
	typedcore "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

const (
	refAnnotation = "git2kube.github.com/ref"
	bufferSize    = 1024
)

// LoadType upload type.
type LoadType int

// MergeType how to merge ConfigMap data.
type MergeType string

const (
	// Delete merge all keys (files) including removal of missing keys.
	Delete MergeType = "delete"
	// Upsert merge all keys (files) but don't remove missing keys from the repository.
	Upsert = "upsert"
)

// LoadType options enum.
const (
	ConfigMap LoadType = iota
	Secret
	Folder
)

// FileIter provides an iterator for the files in a tree.
type FileIter interface {
	ForEach(cb func(*object.File) error) error
}

// UploaderFactory factory constructing Uploaders.
type UploaderFactory func(o UploaderOptions) (Uploader, error)

// Uploader uploading data to target.
type Uploader interface {
	// Upload files into config map tagged by commitID
	Upload(commitID string, iter FileIter) error
}

type uploader struct {
	restconfig  *rest.Config
	clientset   kubernetes.Interface
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

type folderUploader struct {
	name       string
	includes   []*regexp.Regexp
	excludes   []*regexp.Regexp
	sourcePath string
}

// UploaderOptions uploader options.
type UploaderOptions struct {
	Kubeconfig  bool
	Source      string
	Target      string
	Namespace   string
	MergeType   MergeType
	Includes    []string
	Excludes    []string
	Labels      []string
	Annotations []string
}

var uploaderFactories = make(map[LoadType]UploaderFactory)

func register(loadType LoadType, factory UploaderFactory) {
	_, registered := uploaderFactories[loadType]
	if registered {
		log.Errorf("Uploader factory %d already registered. Ignoring.", loadType)
	}
	uploaderFactories[loadType] = factory
}

// NewUploader create uploader of specific type.
func NewUploader(lt LoadType, o UploaderOptions) (Uploader, error) {
	engineFactory, ok := uploaderFactories[lt]
	if !ok {
		return nil, errors.New("invalid uploader name")
	}

	// Run the factory with the configuration.
	return engineFactory(o)
}

func newConfigMapUploader(o UploaderOptions) (Uploader, error) {
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
		name:        o.Target,
	}, nil
}

func (u *configmapUploader) Upload(commitID string, iter FileIter) error {
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

func (u *configmapUploader) iterToConfigMapData(iter FileIter) (map[string]string, error) {
	data := make(map[string]string)
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

func newSecretUploader(o UploaderOptions) (Uploader, error) {
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

	return &secretUploader{
		mergeType:   o.MergeType,
		includes:    includesRegex,
		excludes:    excludesRegex,
		labels:      labelsParsed,
		annotations: annotationsParsed,
		restconfig:  restconfig,
		clientset:   clientset,
		namespace:   o.Namespace,
		name:        o.Target,
	}, nil
}

func (u *secretUploader) Upload(commitID string, iter FileIter) error {
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

	patchBytes, err := strategicpatch.CreateTwoWayMergePatch(oldData, newData, corev1.Secret{})
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

func (u *secretUploader) iterToSecretData(iter FileIter) (map[string][]byte, error) {
	data := make(map[string][]byte)
	err := iter.ForEach(func(file *object.File) error {
		if filterFile(file, u.includes, u.excludes) {
			content, err := file.Contents()
			if err != nil {
				return err
			}
			data[strings.Replace(file.Name, "/", ".", -1)] = []byte(content)
		}
		return nil
	})

	return data, err
}

func newFolderUploader(o UploaderOptions) (Uploader, error) {
	err := os.RemoveAll(o.Target)
	if err != nil {
		return nil, err
	}
	err = os.MkdirAll(o.Target, os.ModePerm)
	if err != nil {
		return nil, err
	}
	log.Infof("Created empty folder %s", o.Target)

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

	return &folderUploader{
		includes:   includesRegex,
		excludes:   excludesRegex,
		name:       o.Target,
		sourcePath: o.Source,
	}, nil
}

func (u *folderUploader) Upload(commitID string, iter FileIter) error {
	filesToKeep := make(map[string]bool)
	err := iter.ForEach(func(file *object.File) error {
		if filterFile(file, u.includes, u.excludes) {
			src := path.Join(u.sourcePath, file.Name)
			if _, err := os.Lstat(src); err == nil {
				src, _ = filepath.Abs(src)
			}
			dst := path.Join(u.name, file.Name)
			filesToKeep[dst] = true

			source, err := os.Open(src)
			if err != nil {
				return err
			}
			defer source.Close()

			if dir, _ := filepath.Split(dst); dir != "" {
				err = os.MkdirAll(dir, 0o777)
				if err != nil {
					return err
				}
			}

			destination, err := os.Create(dst)
			if err != nil {
				return err
			}
			defer destination.Close()

			buf := make([]byte, bufferSize)
			for {
				n, err := source.Read(buf)
				if err != nil && err != io.EOF {
					return err
				}
				if n == 0 {
					break
				}
				if _, err := destination.Write(buf[:n]); err != nil {
					return err
				}
			}
		}
		return nil
	})
	if err != nil {
		return err
	}

	err = filepath.Walk(u.name, func(path string, info os.FileInfo, err error) error {
		if _, exists := filesToKeep[path]; info != nil && !info.IsDir() && !exists {
			err := os.Remove(path)
			if err != nil {
				return err
			}
		}
		return nil
	})

	return err
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

func init() {
	register(ConfigMap, newConfigMapUploader)
	register(Secret, newSecretUploader)
	register(Folder, newFolderUploader)
}
