package cmd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/lijiang2014/cwl.go"
	"github.com/lijiang2014/cwl.go/intergration/client"
	"github.com/lijiang2014/cwl.go/intergration/sfs"
	"github.com/lijiang2014/cwl.go/intergration/slex"
	"github.com/lijiang2014/cwl.go/runner"
	"github.com/spf13/cobra"
	"golang.org/x/term"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"path"
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

// RemoteConfig describe a remote task
type RemoteConfig struct {
	Doc        string                   `json:"doc" yaml:"doc"`
	Job        string                   `json:"job" yaml:"job"`
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
	err := clientConfig.SetDefault()
	if errors.Is(err, client.ErrorNoToken) {
		fmt.Printf("Need to login to starlight( %s )\n", clientConfig.BaseURL)
		if clientConfig.Username != "" {
			fmt.Printf("Try login as: %s\n", clientConfig.Username)
		} else {
			// Username
			fmt.Printf("(username) Login as: ")
			_, err = fmt.Scan(&clientConfig.Username)
			if err != nil {
				return err
			}
			// Password
			fmt.Printf("(password) Input password: ")
			raw, err := term.ReadPassword(int(os.Stdin.Fd()))
			if err != nil {
				return err
			}
			clientConfig.Password = string(raw)
			fmt.Println("Login...")
			err = clientConfig.SetDefault()
			if err != nil {
				return err
			}
			fmt.Println("Login Success")
		}
	} else {
		return err
	}
	// New Remote Importer
	tmpClient, err := generateStarlightClient()
	if err != nil {
		return err
	}
	importer, err := sfs.New(context.TODO(), clientConfig.Token, tmpClient, "", false)
	if err != nil {
		return err
	}

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
	fmt.Printf("Using remote importer to read Job: %s\n", config.Job)
	rawJob, err := importer.Load(config.Job)
	if err != nil {
		return err
	}
	job, err := cwl.Y2J(rawJob)
	if err != nil {
		return err
	}

	// DocDir
	var docDir string
	if sfs.IsFileSystemFile(config.Doc) {
		docDir = path.Dir(config.Doc)
	} else {
		docDir = config.Allocation.Default.WorkDir.HostPath
	}

	// New engine
	engineConf := runner.EngineConfig{
		DocumentID: "",
		RunID:      "cwl.go",
		Importer:   importer,
		NewFSMethod: func(workdir string) (runner.Filesystem, error) {
			tmpClient, err := generateStarlightClient()
			if err != nil {
				return nil, err
			}
			fs, err := sfs.New(context.TODO(), clientConfig.Token, tmpClient, "", true)
			if err != nil {
				return nil, err
			}
			return fs, nil
		},
		Process:      doc,
		Params:       job,
		DocImportDir: docDir,
		RootHost:     config.Allocation.Default.WorkDir.HostPath,
		InputsDir:    "inputs",
		WorkDir:      "run",
	}
	engine, err := runner.NewEngine(engineConf)
	if err != nil {
		return err
	}

	// New Executor
	tmpClient, err = generateStarlightClient()
	if err != nil {
		return err
	}
	id, err := uuid.NewRandom()
	if err != nil {
		return err
	}
	exec, err := slex.New(context.TODO(), id.String(), tmpClient, clientConfig.Token, config.Allocation)
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

func generateStarlightClient() (*client.StarlightClient, error) {
	// TODO
	return nil, fmt.Errorf("TODO")
}
