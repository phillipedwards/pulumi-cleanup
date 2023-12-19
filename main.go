package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/pulumi/pulumi/sdk/v3/go/auto"
	"github.com/pulumi/pulumi/sdk/v3/go/auto/optdestroy"
	"github.com/pulumi/pulumi/sdk/v3/go/auto/optup"
	"github.com/pulumi/pulumi/sdk/v3/go/common/workspace"
)

type RefreshProgram struct {
	Project string
	Stack   string
	Runtime string
}

func main() {
	var stackArgs []string
	argsWithoutProg := os.Args[1:]
	if len(argsWithoutProg) > 0 {
		stackArgs = strings.Split(argsWithoutProg[0], ",")
	}

	ctx := context.Background()
	ws, err := auto.NewLocalWorkspace(ctx)
	if err != nil {
		fmt.Println("Unable to create locate workspace")
		panic(err)
	}

	stacks, err := ws.ListStacks(ctx)
	if err != nil {
		fmt.Sprintln("unable to list all stacks")
		panic(err)
	}

	var refreshStacks []RefreshProgram
	for _, stack := range stacks {
		resourceCount := stack.ResourceCount
		if resourceCount != nil && *resourceCount <= 0 {
			fmt.Sprintln("ignore stack %s with no resources", stack.Name)
			continue
		}

		tags, err := ws.ListTags(ctx, stack.Name)
		if err != nil {
			fmt.Sprintln("unable to retrieve tags for stack %s", stack.Name)
		}

		proj := tags["pulumi:project"]
		runtime := tags["pulumi:runtime"]

		refreshStacks = append(refreshStacks, RefreshProgram{
			Project: proj,
			Runtime: runtime,
			Stack:   stack.Name,
		})
	}

	pwd, err := os.Getwd()
	if err != nil {
		fmt.Println("unable to get working directory")
		panic(err)
	}

	for _, refresh := range refreshStacks {
		tmpDir, err := os.MkdirTemp(pwd, "")
		if err != nil {
			panic(err)
		}

		yamlString := fmt.Sprintf("name: %s\nruntime: %s", refresh.Project, refresh.Runtime)
		data := []byte(yamlString)
		err = os.WriteFile("Pulumi.yaml", data, 0644)
		if err != nil {
			fmt.Sprintln("unable to write Pulumi.yaml for project %s", refresh.Project)
		}

		stack, err := auto.SelectStackLocalSource(ctx, refresh.Stack, tmpDir)
		if err != nil {
			fmt.Sprintln("unable to select stack %s", refresh.Stack)
			panic(err)
		}

		_, err = stack.RefreshConfig(ctx)
		if err != nil {
			fmt.Sprintln("unable to refresh config for stack %s", refresh.Stack)
			panic(err)
		}

		_, err = stack.Refresh(ctx)
		if err != nil {
			fmt.Sprintln("unable to refresh stack %s", refresh.Stack)
			panic(err)
		}
	}
}
