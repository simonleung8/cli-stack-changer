package main

import (
	"errors"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/cloudfoundry/cli/flags"
	"github.com/cloudfoundry/cli/flags/flag"
	"github.com/cloudfoundry/cli/plugin"
	"github.com/simonleung8/cli-stack-changer/apps"
	"github.com/simonleung8/cli-stack-changer/instances"
	"github.com/simonleung8/cli-stack-changer/orgs"
	"github.com/simonleung8/cli-stack-changer/spaces"
)

type StackChanger struct {
	ui terminal.UI
}

func (c *StackChanger) GetMetadata() plugin.PluginMetadata {
	return plugin.PluginMetadata{
		Name: "Stack-Changer",
		Version: plugin.VersionType{
			Major: 1,
			Minor: 0,
			Build: 0,
		},
		Commands: []plugin.Command{
			{
				Name:     "stack-change",
				HelpText: "Update stacks for apps from lucid64 to cflinuxfs2. Restart started apps.",
				UsageDetails: plugin.Usage{
					Usage: `cf stack-change [-o {org}] [-s {space}] [-p {parallel}]`,
				},
			},
			{
				Name:     "stack-list",
				HelpText: "List all apps running on stack lucid64.",
				UsageDetails: plugin.Usage{
					Usage: "cf stack-list [-o {org}] [-s {space}]",
				},
			},
		},
	}
}

func main() {
	plugin.Start(new(StackChanger))
}

func (cmd *StackChanger) Run(cliConnection plugin.CliConnection, args []string) {
	cmd.ui = terminal.NewUI(os.Stdin, terminal.NewTeePrinter())
	appsObj := apps.NewApps(cliConnection)
	instancesObj := instances.NewInstances(cliConnection)

	fc := flags.NewFlagContext(setupFlags())
	err := fc.Parse(args[1:]...)
	if err != nil {
		cmd.exitWithError(err)
	}

	if args[0] == "stack-change" {
		cmd.ui.Say("Getting all apps with lucid64 stack...")
		allApps := cmd.getApps(cliConnection, fc, appsObj)
		cmd.ui.Say(terminal.SuccessColor("OK"))
		cmd.ui.Say("")

		i := 0
		j := 10 //default throttle
		if fc.IsSet("p") {
			j = fc.Int("p")
			if j < 1 || j > 100 {
				cmd.ui.Failed("`-p` has to be in the range of 1 - 100")
			}
		}

		for i < len(allApps.Resources) {
			if i+j >= len(allApps.Resources) {
				j = len(allApps.Resources) - i
			}

			cmd.ui.Say(fmt.Sprintf("Total %d found, %d processed. Batch processing %d at a time ...", len(allApps.Resources), i, j))

			cmd.updateAndRestart(appsObj, instancesObj, allApps.Resources[i:i+j])
			i = i + j
		}

	} else if args[0] == "stack-list" {
		cmd.ui.Say("Getting all apps with lucid64 stack...")
		allApps := cmd.getApps(cliConnection, fc, appsObj)
		cmd.ui.Say(terminal.SuccessColor("OK"))
		cmd.ui.Say("")
		cmd.ui.Say(fmt.Sprintf("Total %d found ...", len(allApps.Resources)))
		cmd.printTable(allApps.Resources)
	}
}

func setupFlags() map[string]flags.FlagSet {
	fs := make(map[string]flags.FlagSet)
	fs["o"] = &cliFlags.StringFlag{Name: "o", Usage: ""}
	fs["s"] = &cliFlags.StringFlag{Name: "s", Usage: ""}
	fs["p"] = &cliFlags.IntFlag{Name: "p", Usage: ""}
	return fs
}

func (cmd *StackChanger) getOrgs(cliConnection plugin.CliConnection, fc flags.FlagContext) []orgs.OrgModel {
	o := orgs.NewOrgs(cliConnection)

	if fc.IsSet("o") {
		oneOrg, err := o.GetOrg(fc.String("o"))
		if err != nil {
			cmd.exitWithError(err)
		}

		return []orgs.OrgModel{oneOrg}
	} else {
		allOrgs, err := o.GetAllOrgs()
		if err != nil {
			cmd.exitWithError(err)
		}

		return allOrgs
	}
}

func (cmd *StackChanger) getApps(cliConnection plugin.CliConnection, fc flags.FlagContext, a apps.Apps) apps.AppsModel {
	var allApps apps.AppsModel
	var err error
	var oneOrg orgs.OrgModel

	if fc.IsSet("s") {
		if !fc.IsSet("o") {
			cmd.exitWithError(errors.New(fmt.Sprintf("Please provide the organization which space '%s' belongs to\n", fc.String("s"))))
		}

		s := spaces.NewSpaces(cliConnection)
		spaceGuid, err := s.GetSpaceGuid(fc)
		if err != nil {
			cmd.exitWithError(err)
		}

		allApps, err = a.GetLucid64AppsFromSpace(spaceGuid)
		if err != nil {
			cmd.exitWithError(err)
		}

	} else if fc.IsSet("o") {
		o := orgs.NewOrgs(cliConnection)
		oneOrg, err = o.GetOrg(fc.String("o"))
		if err != nil {
			cmd.exitWithError(err)
		}

		if oneOrg.Metadata.Guid == "" {
			cmd.exitWithError(errors.New(fmt.Sprintf("Org %s is not found\n", fc.String("o"))))
		}

		allApps, err = a.GetLucid64AppsFromOrg(oneOrg.Metadata.Guid)
		if err != nil {
			cmd.exitWithError(err)
		}
	} else {
		allApps, err = a.GetLucid64Apps()
		if err != nil {
			cmd.exitWithError(err)
		}
	}

	return allApps
}

func (cmd *StackChanger) updateAndRestart(appsObj apps.Apps, instancesObj instances.Instances, allApps []apps.AppModel) {
	var wg sync.WaitGroup

	cmd.printTable(allApps)

	for _, a := range allApps {
		wg.Add(1)

		go func(app apps.AppModel, allApps []apps.AppModel) {
			defer wg.Done()
			if app.Entity.State == "STARTED" {
				err := appsObj.UpdateStackAndStopApp(app.Metadata.Guid)
				if err != nil {
					cmd.ui.Warn("Error updating stack for app '"+app.Entity.Name+"' ("+app.Metadata.Guid+")", err.Error())
					return
				}
				allApps = updateAppState(allApps, app.Entity.Name, "Updated, Restarting")
				cmd.reprintTable(allApps, app.Entity.Name, "Updated, Restarting")

				err = appsObj.RestartApp(app.Metadata.Guid)
				if err != nil {
					allApps = updateAppState(allApps, app.Entity.Name, "Updated, Error Restarting")
					cmd.reprintTable(allApps, app.Entity.Name, "Updated, Error Restarting")
					return
				}

				err = instancesObj.IsAnyInstancesStarted(app.Metadata.Guid, 600*time.Second)
				if err != nil {
					allApps = updateAppState(allApps, app.Entity.Name, "Updated, Timeout Restarting")
					cmd.reprintTable(allApps, app.Entity.Name, "Updated, Timeout Restarting")
					return
				}

				allApps = updateAppState(allApps, app.Entity.Name, "Updated, Restarted ")
				cmd.reprintTable(allApps, app.Entity.Name, "Updated, Restarted ")
			} else {
				err := appsObj.UpdateStack(app.Metadata.Guid)
				if err != nil {
					cmd.ui.Warn("Error updating stack for app '"+app.Entity.Name+"' ("+app.Metadata.Guid+")", err.Error())
					return
				}

				allApps = updateAppState(allApps, app.Entity.Name, "Updated, Done")
				cmd.reprintTable(allApps, app.Entity.Name, "Updated, Done")
			}
		}(a, allApps)

	}
	wg.Wait()
	fmt.Printf("\033[2B")
	fmt.Printf("\n\n")
}

func (cmd *StackChanger) printTable(allApps []apps.AppModel) {
	table := terminal.NewTable(cmd.ui, []string{"name", "guid", "state"})
	for _, a := range allApps {
		table.Add(a.Entity.Name, a.Metadata.Guid, a.Entity.State)
	}
	table.Print()
}

func (cmd *StackChanger) reprintTable(allApps []apps.AppModel, updatedApp string, msg string) {
	n := len(allApps) + 1
	fmt.Printf("\033[%dA", n)
	table := terminal.NewTable(cmd.ui, []string{"name", "guid", "state"})
	for _, a := range allApps {
		if a.Entity.Name != updatedApp {
			table.Add(a.Entity.Name, a.Metadata.Guid, a.Entity.State)
		} else {
			table.Add(updatedApp, a.Metadata.Guid, msg)
		}
	}
	table.Print()
}

func (c *StackChanger) exitWithError(err error) {
	c.ui.Failed("Error: " + err.Error())
}

func updateAppState(allApps []apps.AppModel, updatedApp string, msg string) []apps.AppModel {
	for i, _ := range allApps {
		if allApps[i].Entity.Name == updatedApp {
			allApps[i].Entity.State = msg
			return allApps
		}
	}
	return allApps
}
