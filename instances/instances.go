package instances

import (
	"encoding/json"
	"errors"
	"os/exec"
	"time"

	"github.com/cloudfoundry/cli/plugin"
)

type Instances interface {
	IsAnyInstancesStarted(string, time.Duration) error
}

type InstancesModel struct {
	ErrorCode   string `json:"error_code"`
	Description string `json:"description"`
}

type instances struct {
	cliCon plugin.CliConnection
}

func NewInstances(cliCon plugin.CliConnection) Instances {
	return &instances{
		cliCon: cliCon,
	}
}

func (cmd *instances) IsAnyInstancesStarted(appGuid string, timeout time.Duration) error {
	var info InstancesModel
	stagingStartTime := time.Now()

	for time.Since(stagingStartTime) < timeout {
		info = InstancesModel{}

		output, err := exec.Command("cf", "curl", "/v2/apps/"+appGuid+"/instances").Output()
		if err != nil {
			return err
		}

		err = json.Unmarshal(output, &info)
		if err != nil {
			return err
		}

		if info.Description == "" && info.ErrorCode == "" {
			return nil
		}

		time.Sleep(5 * time.Second)
	}

	return errors.New(info.Description)
}
