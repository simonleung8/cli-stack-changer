package stacks

import (
	"encoding/json"
	"errors"
	"os/exec"

	"github.com/cloudfoundry/cli/plugin"
)

const (
	lucid64    = "lucid64"
	cflinuxfs2 = "cflinuxfs2"
)

type StacksModel struct {
	Resources []StackModel `json:"resources"`
}

type StackModel struct {
	Metadata MetadataModel `json:"metadata"`
	Entity   EntityModel   `json:"entity"`
}

type MetadataModel struct {
	Guid string `json:"guid"`
}

type EntityModel struct {
	Name string `json:"name"`
}

type Stacks interface {
	GetLucid64Guid() (string, error)
	GetCflinuxfs2Guid() (string, error)
}

type stacks struct {
	cliCon plugin.CliConnection
}

func NewStacks(cliConnection plugin.CliConnection) Stacks {
	return &stacks{
		cliCon: cliConnection,
	}
}

func (s *stacks) GetLucid64Guid() (string, error) {
	return s.getStackGuid(lucid64)
}

func (s *stacks) GetCflinuxfs2Guid() (string, error) {
	return s.getStackGuid(cflinuxfs2)
}

func (s *stacks) getStackGuid(name string) (string, error) {
	// output, err := s.cliCon.CliCommandWithoutTerminalOutput("curl", "/v2/stacks")
	output, err := exec.Command("cf", "curl", "/v2/stacks").Output()
	if err != nil {
		return "", err
	}

	model := StacksModel{}
	// err = json.Unmarshal([]byte(output[0]), &model)
	err = json.Unmarshal(output, &model)
	if err != nil {
		return "", err
	}

	if len(model.Resources) == 0 {
		return "", errors.New("No stacks named " + name + " is found")
	}

	for _, s := range model.Resources {
		if s.Entity.Name == name {
			return s.Metadata.Guid, nil
		}
	}

	return "", nil
}
