package broker

// Backend describes the location for a backend that can run Patroni nodes
type Backend struct {
	URI      string
	Username string
	Password string
}
