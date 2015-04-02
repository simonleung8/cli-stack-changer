package orgs

import (
	"encoding/json"
	"strings"

	"github.com/cloudfoundry/cli/plugin"
)

type Orgs interface {
	GetAllOrgs() ([]OrgModel, error)
	GetOrg(name string) (OrgModel, error)
}

type OrgsModel struct {
	Resources []OrgModel `json:"resources"`
}

type MetadataModel struct {
	Guid string `json:"guid"`
}

type EntityModel struct {
	Name string `json:"name"`
}

type OrgModel struct {
	Metadata MetadataModel `json:"metadata"`
	Entity   EntityModel   `json:"entity"`
}

type orgs struct {
	cliCon plugin.CliConnection
}

func NewOrgs(cliConnection plugin.CliConnection) Orgs {
	return &orgs{
		cliCon: cliConnection,
	}
}

func (o *orgs) GetAllOrgs() ([]OrgModel, error) {
	output, err := o.cliCon.CliCommandWithoutTerminalOutput("curl", "/v2/organizations")
	if err != nil {
		return []OrgModel{}, err
	}

	model := OrgsModel{}
	err = json.Unmarshal([]byte(output[0]), &model)
	return model.Resources, err
}

func (o *orgs) GetOrg(name string) (OrgModel, error) {
	output, err := o.cliCon.CliCommandWithoutTerminalOutput("curl", "/v2/organizations")
	if err != nil {
		return OrgModel{}, err
	}

	model := OrgsModel{}
	err = json.Unmarshal([]byte(output[0]), &model)
	if err != nil {
		return OrgModel{}, err
	}

	for _, o := range model.Resources {
		if strings.ToLower(o.Entity.Name) == strings.ToLower(name) {
			return o, nil
		}
	}

	return OrgModel{}, err
}
