package config

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"

	"github.com/hashicorp/errwrap"

	"gopkg.in/yaml.v1"
)

// PatroniV12Specification is constructed based on ClusterSpecification provided by the API.
// It is converted to a patroni.yml and used by Patroni to configure & run PostgreSQL.
// The scheme is for Patroni v1.2
type PatroniV12Specification struct {
	Scope     string `yaml:"scope"`
	Namespace string `yaml:"namespace"`
	Name      string `yaml:"name"`
	Restapi   struct {
		Listen         string `yaml:"listen"`
		ConnectAddress string `yaml:"connect_address"`
	} `yaml:"restapi"`
	Etcd struct {
		URL      string `yaml:"url"`
		Host     string `yaml:"host,omitempty"`
		Port     int    `yaml:"port,omitempty"`
		Protocol string `yaml:"protocol,omitempty"`
		Username string `yaml:"username,omitempty"`
		Password string `yaml:"password,omitempty"`
		Cacert   string `yaml:"cacert,omitempty"`
		Cert     string `yaml:"cert,omitempty"`
		Key      string `yaml:"key,omitempty"`
		Srv      string `yaml:"srv,omitempty"`
		Proxy    string `yaml:"proxy,omitempty"`
	} `yaml:"etcd"`
	Bootstrap struct {
		Dcs struct {
			TTL                  int `yaml:"ttl"`
			LoopWait             int `yaml:"loop_wait"`
			RetryTimeout         int `yaml:"retry_timeout"`
			MaximumLagOnFailover int `yaml:"maximum_lag_on_failover"`
			Postgresql           struct {
				UsePgRewind bool `yaml:"use_pg_rewind"`
				UseSlots    bool `yaml:"use_slots"`
				Parameters  struct {
					WalLevel            string `yaml:"wal_level"`
					HotStandby          string `yaml:"hot_standby"`
					WalKeepSegments     int    `yaml:"wal_keep_segments"`
					MaxWalSenders       int    `yaml:"max_wal_senders"`
					MaxReplicationSlots int    `yaml:"max_replication_slots"`
					WalLogHints         string `yaml:"wal_log_hints"`
					ArchiveMode         string `yaml:"archive_mode"`
					ArchiveTimeout      string `yaml:"archive_timeout"`
					ArchiveCommand      string `yaml:"archive_command"`
				} `yaml:"parameters"`
				RecoveryConf struct {
					RestoreCommand string `yaml:"restore_command"`
				} `yaml:"recovery_conf"`
			} `yaml:"postgresql"`
		} `yaml:"dcs"`
		Initdb []interface{} `yaml:"initdb"`
		PgHba  []string      `yaml:"pg_hba"`
		Users  struct {
			Postgres struct {
				Password string   `yaml:"password"`
				Options  []string `yaml:"options"`
			} `yaml:"postgres"`
		} `yaml:"users"`
	} `yaml:"bootstrap"`
	Postgresql struct {
		Listen         string `yaml:"listen"`
		ConnectAddress string `yaml:"connect_address"`
		DataDir        string `yaml:"data_dir"`
		Pgpass         string `yaml:"pgpass"`
		Authentication struct {
			Replication struct {
				Username string `yaml:"username"`
				Password string `yaml:"password"`
			} `yaml:"replication"`
			Superuser struct {
				Username string `yaml:"username"`
				Password string `yaml:"password"`
			} `yaml:"superuser"`
		} `yaml:"authentication"`
		Parameters struct {
			UnixSocketDirectories string `yaml:"unix_socket_directories"`
		} `yaml:"parameters"`
		Callbacks struct {
			OnStart      string `yaml:"on_start"`
			OnStop       string `yaml:"on_stop"`
			OnRestart    string `yaml:"on_restart"`
			OnRoleChange string `yaml:"on_role_change"`
		} `yaml:"callbacks"`
		CreateReplicaMethod []string `yaml:"create_replica_method"`
		WalE                struct {
			Command                       string `yaml:"command"`
			Envdir                        string `yaml:"envdir"`
			ThresholdMegabytes            int    `yaml:"threshold_megabytes"`
			ThresholdBackupSizePercentage int    `yaml:"threshold_backup_size_percentage"`
			Retries                       int    `yaml:"retries"`
			UseIam                        int    `yaml:"use_iam"`
			NoMaster                      int    `yaml:"no_master"`
		} `yaml:"wal_e"`
	} `yaml:"postgresql"`
	Tags struct {
		Nofailover    bool `yaml:"nofailover"`
		Noloadbalance bool `yaml:"noloadbalance"`
		Clonefrom     bool `yaml:"clonefrom"`
	} `yaml:"tags"`
}

var defaultPatroniSpec *PatroniV12Specification

// BuildPatroniSpec merges cluster config with defaults
func BuildPatroniSpec(clusterSpec *ClusterSpecification, hostDiscoverySpec *HostDiscoverySpecification) (patroniSpec *PatroniV12Specification, err error) {
	patroniSpec, err = DefaultPatroniSpec()
	if err != nil {
		return
	}
	patroniSpec.MergeClusterSpec(clusterSpec, hostDiscoverySpec)
	return
}

// DefaultPatroniSpec provides default patroni v1.1 config
func DefaultPatroniSpec() (*PatroniV12Specification, error) {
	if defaultPatroniSpec == nil {
		filename, err := filepath.Abs(APISpec().PatroniDefaultPath)
		if err != nil {
			return nil, err
		}
		yamlFile, err := ioutil.ReadFile(filename)
		if err != nil {
			return nil, err
		}
		defaultPatroniSpec = &PatroniV12Specification{}
		err = yaml.Unmarshal(yamlFile, defaultPatroniSpec)
		if err != nil {
			return nil, err
		}
	}
	return defaultPatroniSpec, nil
}

// MergeClusterSpec builds patroni v1.1 config specification
func (patroniSpec *PatroniV12Specification) MergeClusterSpec(clusterSpec *ClusterSpecification, hostDiscoverySpec *HostDiscoverySpecification) {
	appuserName := clusterSpec.Postgresql.Appuser.Username
	replicationUsername := appuserName
	patroniSpec.Etcd.URL = clusterSpec.Etcd.URI
	patroniSpec.Scope = clusterSpec.Cluster.Scope
	patroniSpec.Name = clusterSpec.Cluster.Name
	if clusterSpec.Cluster.Namespace == "" {
		clusterSpec.Cluster.Namespace = "/service/"
		fmt.Fprintln(os.Stderr, "Using default namespace:", clusterSpec.Cluster.Namespace)
	}
	patroniSpec.Namespace = clusterSpec.Cluster.Namespace
	patroniSpec.Bootstrap.PgHba = []string{
		fmt.Sprintf("host replication %s 0.0.0.0/0 md5", replicationUsername),
		"host postgres all 0.0.0.0/0 md5",
	}
	patroniSpec.Bootstrap.Users.Postgres.Password = clusterSpec.Postgresql.Admin.Password
	patroniSpec.Postgresql.Authentication.Replication.Username = clusterSpec.Postgresql.Appuser.Username
	patroniSpec.Postgresql.Authentication.Replication.Password = clusterSpec.Postgresql.Appuser.Password
	patroniSpec.Postgresql.Authentication.Superuser.Username = clusterSpec.Postgresql.Superuser.Username
	patroniSpec.Postgresql.Authentication.Superuser.Password = clusterSpec.Postgresql.Superuser.Password

	patroniSpec.Postgresql.ConnectAddress = fmt.Sprintf("%s:%s", hostDiscoverySpec.IP, hostDiscoverySpec.Port5432)
	patroniSpec.Restapi.ConnectAddress = fmt.Sprintf("%s:%s", hostDiscoverySpec.IP, hostDiscoverySpec.Port8008)
}

func (patroniSpec *PatroniV12Specification) String() string {
	bytes, err := yaml.Marshal(patroniSpec)
	if err != nil {
		panic(err)
	}
	return string(bytes[:])
}

// CreateConfigFile creates a config file from patroni specification
func (patroniSpec *PatroniV12Specification) CreateConfigFile(path string) (err error) {
	data, err := yaml.Marshal(patroniSpec)
	if err != nil {
		return err
	}
	err = os.MkdirAll(filepath.Dir(path), 0755)
	if err != nil {
		return
	}
	err = ioutil.WriteFile(path, data, 0644)
	return
}

// CreateURIFile creates a file containing superuser URI
func (patroniSpec *PatroniV12Specification) CreateURIFile(createPath string) (err error) {
	err = os.MkdirAll(path.Dir(createPath), 0755)
	if err != nil {
		return errwrap.Wrapf("Cannot mkdir: {{err}}", err)
	}

	username := patroniSpec.Postgresql.Authentication.Superuser.Username
	password := patroniSpec.Postgresql.Authentication.Superuser.Password
	host := os.Getenv("DOCKER_HOST_IP")
	port := os.Getenv("DOCKER_HOST_PORT_5432")
	uri := fmt.Sprintf("postgres://%s:%s@%s:%s/postgres", username, password, host, port)
	err = ioutil.WriteFile(createPath, []byte(uri), 0644)
	return
}
