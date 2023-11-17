package analyze

import (
	"encoding/json"
	"fmt"
	"unsafe"

	v1 "k8s.io/api/core/v1"
)

const (
	saAnnotation = "kubernetes.io/service-account.name"
)

type AnalysisResults struct {
	NumTotalSecrets     int
	NumTotalPullSecrets int

	NumSecretsByNsHost             int
	NumSecretsByNsHostWithNoSA     int
	NumSecretsByNsNameHost         int
	NumSecretsByNsNameHostWithNoSA int

	SizeBytesByNsHost             int
	SizeBytesByNsHostWithNoSA     int
	SizeBytesByNsNameHost         int
	SizeBytesByNsNameHostWithNoSA int
}

type DockerConfig map[string]DockerConfigEntry

type DockerConfigJSON struct {
	Auths DockerConfig `json:"auths"`
}

// DockerConfigEntry is an entry in the DockerConfig.
type DockerConfigEntry struct {
	Username string
	Password string
	Email    string
}

type secretHolder struct {
	nsHostEntry     map[string]DockerConfigEntry
	nsNameHostEntry map[string]DockerConfigEntry

	nsHostEntryNoSA     map[string]DockerConfigEntry
	nsNameHostEntryNoSA map[string]DockerConfigEntry
}

// PullSecrets WILL modify the secrets map in place
func PullSecrets(secrets map[string]*v1.Secret) (*AnalysisResults, error) {
	result := &AnalysisResults{}
	result.NumTotalSecrets = len(secrets)

	deleted := filterNonPullSecrets(secrets)
	result.NumTotalPullSecrets = result.NumTotalSecrets - deleted

	secretHolder, err := populateSecretHolder(secrets)
	if err != nil {
		return nil, err
	}

	m := secretHolder.nsHostEntry
	result.NumSecretsByNsHost = len(m)
	result.SizeBytesByNsHost = calcSizeInBytes(m)

	m = secretHolder.nsHostEntryNoSA
	result.NumSecretsByNsHostWithNoSA = len(m)
	result.SizeBytesByNsHostWithNoSA = calcSizeInBytes(m)

	m = secretHolder.nsNameHostEntry
	result.NumSecretsByNsNameHost = len(m)
	result.SizeBytesByNsNameHost = calcSizeInBytes(m)

	m = secretHolder.nsNameHostEntryNoSA
	result.NumSecretsByNsNameHostWithNoSA = len(m)
	result.SizeBytesByNsNameHostWithNoSA = calcSizeInBytes(m)

	return result, nil
}

// filterNonPullSecrets returns number of secrets deleted from map
func filterNonPullSecrets(secrets map[string]*v1.Secret) int {
	count := 0
	for id, secret := range secrets {
		switch secret.Type {
		case v1.SecretTypeDockercfg:
			continue
		case v1.SecretTypeDockerConfigJson:
			continue
		}

		count++
		delete(secrets, id)
	}

	return count
}

func populateSecretHolder(secrets map[string]*v1.Secret) (*secretHolder, error) {
	holder := &secretHolder{
		nsHostEntry:         map[string]DockerConfigEntry{},
		nsNameHostEntry:     map[string]DockerConfigEntry{},
		nsHostEntryNoSA:     map[string]DockerConfigEntry{},
		nsNameHostEntryNoSA: map[string]DockerConfigEntry{},
	}

	for _, secret := range secrets {
		dockerConfig, err := convertSecretToDockerConfig(secret)
		if err != nil {
			return nil, err
		}

		// fmt.Printf("%-65v %-50v %v\n", v.Name, v.Namespace, saName)
		for host, entry := range dockerConfig {
			keyNsHostEntry := genKeyNsHostEntry(secret.Namespace, host)
			holder.nsHostEntry[keyNsHostEntry] = entry

			keyNsNameHostEntry := genKeyNsNameHostEntry(secret.Namespace, secret.Name, host)
			holder.nsNameHostEntry[keyNsNameHostEntry] = entry

			saName := secret.GetAnnotations()[saAnnotation]
			if saName == "default" || saName == "" {
				holder.nsHostEntryNoSA[keyNsHostEntry] = entry
				holder.nsNameHostEntryNoSA[keyNsNameHostEntry] = entry
			}

		}

	}

	return holder, nil
}

func convertSecretToDockerConfig(secret *v1.Secret) (DockerConfig, error) {
	var dockerConfig DockerConfig
	switch secret.Type {
	case v1.SecretTypeDockercfg:
		data, ok := secret.Data[v1.DockerConfigKey]
		if !ok {
			return nil, fmt.Errorf("invalid secret, DockerConfigKey not found")
		}

		if err := json.Unmarshal(data, &dockerConfig); err != nil {
			return nil, err
		}
	case v1.SecretTypeDockerConfigJson:
		data, ok := secret.Data[v1.DockerConfigJsonKey]
		if !ok {
			return nil, fmt.Errorf("invalid secret, DockerConfigJsonKey not found")
		}
		var dockerConfigJSON DockerConfigJSON
		if err := json.Unmarshal(data, &dockerConfigJSON); err != nil {
			return nil, err
		}
		dockerConfig = dockerConfigJSON.Auths
	default:
		return nil, fmt.Errorf("unknown secret type: %v", secret.Type)
	}

	return dockerConfig, nil
}

func genKeyNsHostEntry(namespace, host string) string {
	return fmt.Sprintf("^ns:%s-host:%s$", namespace, host)
}

func genKeyNsNameHostEntry(namespace, secretName, host string) string {
	return fmt.Sprintf("^ns:%s-name:%s-host:%s$", namespace, secretName, host)
}

// calcSizeInBytes will not be 100% accurate, the outcome is intended
// to measure relative size increase/decrease.
func calcSizeInBytes(m map[string]DockerConfigEntry) int {
	size := 0 // does not include size of m's header
	for _, v := range m {
		size += int(unsafe.Sizeof(v))
		size += len(v.Username)
		size += len(v.Password)
		size += len(v.Email)
	}

	return size
}
