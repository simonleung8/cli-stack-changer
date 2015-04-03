package orgs

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/cloudfoundry/cli/plugin"
)

type Orgs interface {
	GetAllOrgs() ([]OrgModel, error)
	GetOrg(name string) (OrgModel, error)
}

type OrgsModel struct {
	NextUrl   string     `json:"next_url,omitempty"`
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
	allModel := []OrgModel{}
	nextUrl := "/v2/organizations"

	for nextUrl != "" {

		output, err := o.cliCon.CliCommandWithoutTerminalOutput("curl", nextUrl)
		if err != nil {
			return []OrgModel{}, err
		}

		model := OrgsModel{}
		err = json.Unmarshal([]byte(output[0]), &model)
		if err != nil {
			return []OrgModel{}, err
		}

		allModel = append(allModel, model.Resources...)

		if model.NextUrl != "" {
			nextUrl = model.NextUrl
		} else {
			nextUrl = ""
		}
	}

	return allModel, nil
}

func (o *orgs) GetOrg(name string) (OrgModel, error) {
	nextUrl := "/v2/organizations"

	for nextUrl != "" {

		output, err := o.cliCon.CliCommandWithoutTerminalOutput("curl", nextUrl)
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

		if model.NextUrl != "" {
			nextUrl = model.NextUrl
		} else {
			nextUrl = ""
		}

	}
	return OrgModel{}, errors.New(fmt.Sprintf("Org '%s' does not exist.", name))
}
