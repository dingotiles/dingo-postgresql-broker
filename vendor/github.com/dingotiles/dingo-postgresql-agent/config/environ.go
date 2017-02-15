package config

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	"path"
	"strconv"
	"strings"

	"github.com/hashicorp/errwrap"
)

// Environ generates a folder for consumption by envdir,
// which is used by wal-e to look up secrets without
// exposing them into PostgreSQL itself.
type Environ map[string]string

// NewPatroniEnvironFromClusterSpec setups up environment variables for patroni + scripts
func NewPatroniEnvironFromClusterSpec(clusterSpec *ClusterSpecification) *Environ {
	environ := Environ{}
	environ["REPLICATION_USER"] = clusterSpec.Postgresql.Appuser.Username
	environ["DINGO_NODE"] = clusterSpec.Cluster.Name
	environ["PATRONI_SCOPE"] = clusterSpec.Cluster.Scope
	environ["PG_DATA_DIR"] = "/data/postgres0"

	environ["ETCD_URI"] = clusterSpec.Etcd.URI

	environ["ARCHIVE_METHOD"] = clusterSpec.Archives.Method
	if clusterSpec.UsingWaleS3() {
		environ["AWS_ACCESS_KEY_ID"] = clusterSpec.Archives.S3.AWSAccessKeyID
		environ["AWS_SECRET_ACCESS_KEY"] = clusterSpec.Archives.S3.AWSSecretAccessID
		environ["WAL_S3_BUCKET"] = clusterSpec.Archives.S3.S3Bucket
		environ["WALE_S3_PREFIX"] = clusterSpec.waleS3Prefix()
		environ["WALE_S3_ENDPOINT"] = clusterSpec.Archives.S3.S3Endpoint
	}
	if clusterSpec.UsingWaleLocal() {
		volume := clusterSpec.Archives.Local.LocalBackupVolume
		environ["WALE_LOCAL_PREFIX"] = fmt.Sprintf("local://%s", volume)
		environ["LOCAL_BACKUP_VOLUME"] = volume
	}
	if clusterSpec.UsingWaleSSH() {
		host := clusterSpec.Archives.SSH.Host
		basePath := clusterSpec.Archives.SSH.BasePath
		environ["SSH_HOST"] = host
		environ["SSH_BASE_PATH"] = basePath
		environ["SSH_PORT"] = clusterSpec.Archives.SSH.Port
		environ["SSH_USER"] = clusterSpec.Archives.SSH.User
		environ["SSH_PRIVATE_KEY"] = clusterSpec.Archives.SSH.PrivateKey
		environ["WALE_SSH_PREFIX"] = fmt.Sprintf("ssh://%s%s", host, basePath)
		environ["SSH_IDENTITY_FILE"] = "/home/postgres/.ssh/ssh_backup_storage"
		// contents of $SSH_IDENTITY_FILE created by agent_wrapper.sh
	}

	return &environ
}

// AddEnv adds an addition KEY=VALUE pair
func (environ *Environ) AddEnv(envvar string) {
	if !strings.Contains(envvar, "=") {
		fmt.Fprintf(os.Stderr, "Format error for env var '%s', must be 'KEY=VALUE'", envvar)
		return
	}
	parts := strings.Split(envvar, "=")
	key := parts[0]
	value := parts[1]
	if len(key) == 0 {
		fmt.Fprintf(os.Stderr, "Missing env variable name in '%s'\n", envvar)
		return
	}
	if len(value) == 0 {
		fmt.Fprintf(os.Stderr, "Missing env variable value in '%s'\n", envvar)
		return
	}
	(*environ)[key] = value
}

// CreateEnvDirFiles creates a directory with one file per env var
func (environ *Environ) CreateEnvDirFiles(dir string) (err error) {
	err = os.RemoveAll(dir)
	if err != nil {
		return errwrap.Wrapf("Cannot delete directory: {{err}}", err)
	}

	err = os.MkdirAll(dir, 0755)
	if err != nil {
		return
	}
	for name, value := range *environ {
		data := []byte(value)
		err = ioutil.WriteFile(path.Join(dir, name), data, 0644)
		if err != nil {
			return
		}
	}
	return
}

// CreateEnvScript creates a script that exports env vars
func (environ *Environ) CreateEnvScript(filePath string, chownUser string) (err error) {
	err = os.MkdirAll(path.Dir(filePath), 0755)
	if err != nil {
		return errwrap.Wrapf("Cannot mkdir: {{err}}", err)
	}

	var f *os.File
	f, err = os.Create(filePath)
	if err != nil {
		return errwrap.Wrapf("Cannot create file: {{err}}", err)
	}

	for name, value := range *environ {
		env := fmt.Sprintf("export %s=\"%s\"\n", name, value)
		_, err = f.WriteString(env)
		if err != nil {
			return errwrap.Wrapf("Cannot create write string to file: {{err}}", err)
		}
	}
	f.Sync()

	if chownUser != "" {
		u, err := user.Lookup(chownUser)
		if err != nil {
			return errwrap.Wrapf("Cannot lookup user: {{err}}", err)
		}
		uid, err := strconv.Atoi(u.Uid)
		if err != nil {
			return errwrap.Wrapf("Cannot get user Uid: {{err}}", err)
		}
		gid, err := strconv.Atoi(u.Gid)
		if err != nil {
			return errwrap.Wrapf("Cannot get user group Gid: {{err}}", err)
		}
		err = os.Chown(filePath, uid, gid)
		if err != nil {
			return errwrap.Wrapf("Cannot chown file: {{err}}", err)
		}
	}

	return
}
