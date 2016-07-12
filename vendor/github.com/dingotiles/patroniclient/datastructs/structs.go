package datastructs

import (
	"encoding/json"
	"io"
	"strings"
)

type DataServiceMember struct {
	Role         string `json:"role"`
	State        string `json:"state"`
	XlogLocation int64  `json:"xlog_location"`
	ConnURL      string `json:"conn_url"`
	APIURL       string `json:"api_url"`
	RootAPIURL   string
}

func NewDataServiceMember(jsonValue string) (dataServiceMember *DataServiceMember, err error) {
	dataServiceMember = &DataServiceMember{}
	dec := json.NewDecoder(strings.NewReader(jsonValue))
	if err = dec.Decode(&dataServiceMember); err == io.EOF {
		return
	} else if err != nil {
		return
	}

	dataServiceMember.RootAPIURL = strings.Replace(dataServiceMember.APIURL, "/patroni", "/", 1)

	return
}
