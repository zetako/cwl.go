package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/lijiang2014/cwl.go"
	"github.com/lijiang2014/cwl.go/intergration/slex"
	"github.com/lijiang2014/cwl.go/runner"
	"github.com/spf13/cobra"
	"io/ioutil"
	"os"
	"path"
	"starlight/common/model"
	"strings"
)

var (
	overallFeatureSwitch bool
	useRemote            bool
	remoteTokenFile      string
	flags                runner.EngineFlags
)

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:   "run [doc] [job]",
	Short: "Run a cwl job",
	Long:  `cwl.go can run directly with a cwl document and an inputs`,
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		// set flag
		if !overallFeatureSwitch {
			flags.DisableLoopFeature = true
			flags.DisableLastNonNull = true
		}
		// The args is exact 2 element
		return run(args[0], args[1], flags)
	},
}

func init() {
	rootCmd.AddCommand(runCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// runCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// runCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	runCmd.Flags().BoolVar(&useRemote, "remote", false, "Use remote Executor")
	runCmd.Flags().StringVar(&remoteTokenFile, "token", "starlight.token", "Remote Token File")
}

func splitPackedFile(raw string) (fileName, fragID string) {
	tmp := strings.IndexByte(raw, '#')
	if tmp < 0 {
		return raw, ""
	}
	return raw[:tmp], raw[tmp+1:]
}

func openFileAsJSON(pathlike string) ([]byte, error) {
	f, err := os.Open(pathlike)
	if err != nil {
		return nil, err
	}
	data, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, err
	}
	data, err = cwl.Y2J(data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func run(doc, job string, engineFlags runner.EngineFlags) error {
	fmt.Printf("doc is %s, job is %s", doc, job)

	// read data
	docName, fragID := splitPackedFile(doc)
	docData, err := openFileAsJSON(docName)
	if err != nil {
		return err
	}
	jobData, err := openFileAsJSON(job)
	if err != nil {
		return err
	}

	// create engine
	pwd, err := os.Getwd()
	if err != nil {
		return err
	}
	e, err := runner.NewEngine(runner.EngineConfig{
		DocumentID:   fragID,
		RunID:        "cwl.go",
		Process:      docData,
		Params:       jobData,
		DocImportDir: pwd,
		RootHost:     pwd,
		InputsDir:    path.Join(pwd, "inputs"),
		WorkDir:      path.Join(pwd, "run"),
	})
	if err != nil {
		return err
	}
	if useRemote {
		tokenFile, err := os.Open(remoteTokenFile)
		if err != nil {
			return err
		}
		tokenRaw, err := ioutil.ReadAll(tokenFile)
		if err != nil {
			return err
		}
		token := string(tokenRaw)
		var (
			tmp1 int = 1
			tmp2 int = 0
			tmp3 int = 4096
		)
		exec, err := slex.New(context.TODO(), token, &slex.JobAllocationModel{
			Default: &slex.SingleJobAllocationModel{
				Cluster:   "k8s_uat",
				Cpu:       &tmp1,
				Gpu:       &tmp2,
				Memory:    &tmp3,
				Partition: "ln15",
				WorkDir:   model.Volume{HostPath: "/HOME/nscc-gz_yfb_2/cwl"},
			},
			Diff: map[string]*slex.SingleJobAllocationModel{},
		})
		if err != nil {
			return err
		}
		e.SetDefaultExecutor(exec)
	} else {
		e.SetDefaultExecutor(&runner.LocalExecutor{})
	}
	e.Flags = engineFlags

	// run
	outputs, err := e.Run()
	if err != nil {
		return err
	}

	// output
	outStr, err := json.MarshalIndent(outputs, "", "  ")
	if err != nil {
		return err
	} else {
		fmt.Println(string(outStr))
	}

	return nil
}
