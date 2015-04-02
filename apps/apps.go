package apps

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"

	"github.com/cloudfoundry/cli/plugin"
	"github.com/simonleung8/cli-stack-changer/stacks"
)

type Apps interface {
	UpdateStack(string) error
	UpdateStackAndStopApp(string) error
	RestartApp(string) error
	GetLucid64Apps() (AppsModel, error)
	GetLucid64AppsFromOrg(string) (AppsModel, error)
	GetLucid64AppsFromSpace(string) (AppsModel, error)
}

type AppsModel struct {
	Resources []AppModel `json:"resources"`
}

type MetadataModel struct {
	Guid string `json:"guid"`
}

type EntityModel struct {
	Name      string `json:"name"`
	StackGuid string `json:"stack_guid"`
	State     string `json:"state"`
}

type AppModel struct {
	Metadata MetadataModel `json:"metadata"`
	Entity   EntityModel   `json:"entity"`
}

type apps struct {
	cliCon plugin.CliConnection
}

func NewApps(cliConnection plugin.CliConnection) Apps {
	return &apps{
		cliCon: cliConnection,
	}
}

func (a *apps) UpdateStack(appGuid string) error {
	s := stacks.NewStacks(a.cliCon)
	stackGuid, err := s.GetCflinuxfs2Guid()
	if err != nil {
		return err
	}

	_, err = a.cliCon.CliCommandWithoutTerminalOutput("curl", "/v2/apps/"+appGuid, "-X", "PUT", `-d={"stack_guid":"`+stackGuid+`"}`)

	return err
}

func (a *apps) UpdateStackAndStopApp(appGuid string) error {
	s := stacks.NewStacks(a.cliCon)
	stackGuid, err := s.GetCflinuxfs2Guid()
	if err != nil {
		return err
	}

	_, err = a.cliCon.CliCommandWithoutTerminalOutput("curl", "/v2/apps/"+appGuid, "-X", "PUT", `-d={"stack_guid":"`+stackGuid+`","state":"STOPPED"}`)

	return err
}

func (a *apps) RestartApp(appGuid string) error {
	_, err := a.cliCon.CliCommandWithoutTerminalOutput("curl", "/v2/apps/"+appGuid, "-X", "PUT", `-d={"state":"STARTED"}`)
	return err
}

func (a *apps) GetLucid64Apps() (AppsModel, error) {
	output, err := a.cliCon.CliCommandWithoutTerminalOutput("curl", "/v2/apps")
	if err != nil {
		return AppsModel{}, err
	}

	allApps := AppsModel{}
	err = json.Unmarshal([]byte(output[0]), &allApps)

	return a.filterLucid64App(allApps), err
}

func (a *apps) GetLucid64AppsFromOrg(orgGuid string) (AppsModel, error) {
	output, err := a.cliCon.CliCommandWithoutTerminalOutput("curl", fmt.Sprintf("/v2/apps?q=%s", url.QueryEscape("organization_guid:"+orgGuid)))
	if err != nil {
		return AppsModel{}, err
	}

	allApps := AppsModel{}
	err = json.Unmarshal([]byte(output[0]), &allApps)

	return a.filterLucid64App(allApps), err
}

func (a *apps) GetLucid64AppsFromSpace(spaceGuid string) (AppsModel, error) {
	output, err := a.cliCon.CliCommandWithoutTerminalOutput("curl", fmt.Sprintf("/v2/apps?q=%s", url.QueryEscape("space_guid:"+spaceGuid)))
	if err != nil {
		return AppsModel{}, err
	}

	allApps := AppsModel{}
	err = json.Unmarshal([]byte(output[0]), &allApps)

	return a.filterLucid64App(allApps), err
}

func (a *apps) getLucid64Guid() string {
	s := stacks.NewStacks(a.cliCon)
	guid, err := s.GetLucid64Guid()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	return guid
}

func (a *apps) filterLucid64App(allApps AppsModel) AppsModel {
	stackGuid := a.getLucid64Guid()
	filtered := AppsModel{}

	for _, a := range allApps.Resources {
		if a.Entity.StackGuid == stackGuid {
			filtered.Resources = append(filtered.Resources, a)
		}
	}

	return filtered
}
