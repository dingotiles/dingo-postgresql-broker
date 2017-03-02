package config

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

// ContainerStartupRequest is the expected inbound data when each Dingo Agent starts
type ContainerStartupRequest struct {
	ImageVersion string `json:"image_version"`
	ClusterName  string `json:"cluster"`
	NodeName     string `json:"node"`
	Account      string `json:"account"`
}

// ClusterSpecification describes the cluster configuration provided by central API
type ClusterSpecification struct {
	Cluster struct {
		Name      string `json:"name"`
		Scope     string `json:"scope"`
		Namespace string `json:"namespace"`
	} `json:"cluster"`
	Archives Archives `json:"archives"`
	Etcd     struct {
		URI string `json:"uri"`
	} `json:"etcd"`
	Postgresql struct {
		Admin struct {
			Password string `json:"password"`
		} `json:"admin"`
		Appuser struct {
			Password string `json:"password"`
			Username string `json:"username"`
		} `json:"appuser"`
		Superuser struct {
			Password string `json:"password"`
			Username string `json:"username"`
		} `json:"superuser"`
	} `json:"postgresql"`
}

// Archives describes the different supported backends for patroni/wal-e
type Archives struct {
	Method string `json:"method"`
	S3     struct {
		AWSAccessKeyID    string `json:"aws_access_key_id,omitempty"`
		AWSSecretAccessID string `json:"aws_secret_access_id,omitempty"`
		S3Bucket          string `json:"s3_bucket,omitempty"`
		S3Endpoint        string `json:"s3_endpoint,omitempty"`
	} `json:"s3,omitempty"`
	Local struct {
		LocalBackupVolume string `json:"local_backup_volume,omitempty"`
	} `json:"local,omitempty"`
	SSH struct {
		Host       string `json:"host,omitempty"`
		Port       string `json:"port,omitempty"`
		User       string `json:"user,omitempty"`
		PrivateKey string `json:"private_key,omitempty"`
		BasePath   string `json:"base_path,omitempty"`
	} `json:"ssh,omitempty"`
}

// TODO: POST ClusterName & OrgAuthToken to API

// FetchClusterSpec retrieves the new/existing configuration for a cluster from central API
// If agent does not have $DINGO_NODE/APISpec().NodeName, then
// construct from Host:Port5432 so as to be unique
func FetchClusterSpec() (cluster *ClusterSpecification, err error) {
	apiSpec := APISpec()
	apiClusterSpec := fmt.Sprintf("%s/api", apiSpec.APIURI)
	fmt.Printf("Loading configuration from %s...\n", apiClusterSpec)
	var netClient = &http.Client{
		Timeout: time.Second * 10,
	}

	nodeName := apiSpec.NodeName
	if nodeName == "" {
		hostDiscovery := HostDiscoverySpec()
		nodeName = fmt.Sprintf("%s-%s", hostDiscovery.IP, hostDiscovery.Port5432)
		nodeName = strings.Replace(nodeName, ".", "-", -1)
	}
	startupReq := ContainerStartupRequest{
		ImageVersion: apiSpec.ImageVersion,
		ClusterName:  apiSpec.ClusterName,
		NodeName:     nodeName,
		Account:      apiSpec.Account,
	}
	b := new(bytes.Buffer)
	json.NewEncoder(b).Encode(startupReq)

	resp, err := netClient.Post(apiClusterSpec, "application/json", b)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	json.Unmarshal(body, &cluster)

	return
}

// UsingWaleS3 summarizes whether central API wants patroni to be configured for wal-e to ship backups/WAL
func (cluster *ClusterSpecification) UsingWaleS3() bool {
	return cluster.Archives.Method == "s3"
}

// UsingWaleLocal true if wal-e will push/fetch files to a local filesystem volume
func (cluster *ClusterSpecification) UsingWaleLocal() bool {
	return cluster.Archives.Method == "local"
}

// UsingWaleSSH true if wal-e will push/fetch files to a ssh server filesystem
func (cluster *ClusterSpecification) UsingWaleSSH() bool {
	return cluster.Archives.Method == "ssh"
}

func (cluster *ClusterSpecification) waleS3Prefix() string {
	return fmt.Sprintf("s3://%s/backups/%s/wal/", cluster.Archives.S3.S3Bucket, cluster.Cluster.Scope)
}
