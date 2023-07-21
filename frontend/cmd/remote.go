/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/lijiang2014/cwl.go"
	"github.com/lijiang2014/cwl.go/intergration/sfs"
	"github.com/lijiang2014/cwl.go/intergration/slex"
	"github.com/lijiang2014/cwl.go/runner"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
	"io/ioutil"
)

// remoteCmd represents the remote command
var remoteCmd = &cobra.Command{
	Use:   "remote [config]",
	Short: "Executor Workflow remotely (Supported by starlight)",
	Long:  `Work just like cwl.go run, but all CommandLineTool and FileSystem operation will be executed at starlight.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		conf, err := readConfig(args[0])
		if err != nil {
			return err
		}
		return remote(conf)
	},
}

func init() {
	rootCmd.AddCommand(remoteCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// remoteCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// remoteCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

type RemoteConfig struct {
	Doc        string                   `json:"doc" yaml:"doc"`
	Job        string                   `json:"job" yaml:"job"`
	Token      string                   `json:"token" yaml:"token"`
	Allocation *slex.JobAllocationModel `json:"allocation" yaml:"allocation"`
}

func readConfig(file string) (*RemoteConfig, error) {
	raw, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}
	ret := RemoteConfig{}
	err = yaml.Unmarshal(raw, &ret)
	return &ret, err
}

func remote(config *RemoteConfig) error {
	// New Remote Importer
	var (
		importer runner.Importer
		err      error
	)
	importer, err = sfs.New(context.TODO(), config.Token, "")

	// Import Doc and Job
	fmt.Printf("Using remote importer to read Doc: %s\n", config.Doc)
	rawDoc, err := importer.Load(config.Doc)
	if err != nil {
		return err
	}
	doc, err := cwl.Y2J(rawDoc)
	if err != nil {
		return err
	}
	fmt.Printf("Using remote importer to read Job: %s\n", config.Doc)
	rawJob, err := importer.Load(config.Job)
	if err != nil {
		return err
	}
	job, err := cwl.Y2J(rawJob)
	if err != nil {
		return err
	}

	// New engine
	engineConf := runner.EngineConfig{
		DocumentID: "",
		RunID:      "cwl.go",
		Importer:   importer,
		NewFSMethod: func(workdir string) (runner.Filesystem, error) {
			return sfs.New(context.TODO(), config.Token, workdir)
		},
		Process:   doc,
		Params:    job,
		RootHost:  config.Allocation.Default.WorkDir.HostPath,
		InputsDir: "inputs",
		WorkDir:   "run",
	}
	engine, err := runner.NewEngine(engineConf)
	if err != nil {
		return err
	}

	// New Executor
	exec, err := slex.New(context.TODO(), config.Token, config.Allocation)
	if err != nil {
		return err
	}
	engine.SetDefaultExecutor(exec)

	// Run and get Result
	outputs, err := engine.Run()
	if err != nil {
		return err
	}
	outputJSON, err := json.MarshalIndent(outputs, "", "  ")
	if err != nil {
		return err
	} else {
		fmt.Println(string(outputJSON))
	}

	return nil
}
