package patroni

// ServiceMemberData contains the data advertised by a patroni member
type ServiceMemberData struct {
	ConnURL  string `json:"conn_url"`
	HostPort string `json:"conn_address"`
	Role     string `json:"role"`
	State    string `json:"state"`
}
