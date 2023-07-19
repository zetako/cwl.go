package slex

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/lijiang2014/cwl.go"
	"github.com/lijiang2014/cwl.go/runner"
	"path"
	"starlight/common/httpclient"
	"starlight/common/model"
	"time"
)

const (
	StarlightJobNameLimit = 32
	MaxAllowErrorCount    = 5
)

var JobQueryInterval = [6]time.Duration{time.Second, time.Second * 10, time.Minute, time.Minute * 10, time.Minute * 30, time.Hour}

type StarlightExecutor struct {
	alloc  JobAllocationModel
	ctx    context.Context
	token  string
	client *httpclient.BihuClient
	uuid   uuid.UUID
}

func (s StarlightExecutor) Run(process *runner.Process) (runID string, retChan <-chan int, err error) {
	// basic setting
	submit := NewSubmitModelFrom(s.alloc.Get(process.PathID))

	// start migrate
	migrateRecord, err := process.MigrateInputs()
	if err != nil {
		return "", nil, err
	}

	// others
	submit.RuntimeParams.JobName = process.ShortPath(StarlightJobNameLimit)
	submit.RuntimeParams.Env = process.Env()
	submit.RuntimeParams.Stdin, submit.RuntimeParams.Stdout, submit.RuntimeParams.Stderr = process.GetRedirection()

	submit.RuntimeParams.Cmd, err = process.Command()

	// container job
	var (
		dockerReq *cwl.DockerRequirement
	)
	reqs := process.Root().Process.Base().Requirements
	hints := process.Root().Process.Base().Hints
	for _, r := range reqs {
		if r.ClassName() == "DockerRequirement" {
			dockerReq = r.(*cwl.DockerRequirement)
		}
	}
	if dockerReq != nil {
		for _, r := range hints {
			if r.ClassName() == "DockerRequirement" {
				dockerReq = r.(*cwl.DockerRequirement)
			}
		}
	}
	if dockerReq != nil {
		// setup image
		submit.RuntimeParams.Image = dockerReq.DockerPull
		// setup mount
		for _, rec := range migrateRecord {
			if rec.IsSymLink {
				submit.RuntimeParams.Volume = append(submit.RuntimeParams.Volume, model.Volume{
					HostPath:  rec.Source,
					MountPath: rec.Source,
				})
			}
		}
	}

	// send req
	var job model.Job
	_, err = s.client.PostSpec("job/submit", submit, &job)
	if err != nil {
		return "", nil, err
	}

	bidChan := make(chan int)
	go s.waitingJob(job.ClusterName, job.ClusterJobID, bidChan)
	return fmt.Sprint(job.ClusterJobID), bidChan, nil
}

func (s StarlightExecutor) QueryRuntime(p *runner.Process) (runner.Runtime, error) {
	var (
		workdir string
	)
	// 1. get alloc
	allocation := s.alloc.Get(p.PathID)
	// 2. if no workdir, then add a default
	if allocation.WorkDir.HostPath == "" {
		// 2.1 check home dir in cluster/partition
		base, err := s.getPartitionBaseDir(allocation)
		if err != nil {
			return runner.Runtime{}, err
		}
		// 2.2 generate dir path
		workdir = path.Join(base, s.uuid.String(), p.Path())
		// 2.3 save the new generated work dir
		s.alloc.Set(p.PathID, SingleJobAllocationModel{
			WorkDir: model.Volume{
				HostPath: workdir,
			},
		})
	}
	// 3. generate runtime
	rt := runner.Runtime{
		RootHost:   workdir,
		InputsHost: path.Join(workdir, "inputs"),
		ExitCode:   nil,
		Outdir:     path.Join(workdir, "outputs"),
		Tmpdir:     path.Join(workdir, "tmp"),
	}
	if allocation.Cpu != nil {
		rt.Cores = *allocation.Cpu
	}
	if allocation.Memory != nil {
		rt.RAM = runner.Mebibyte(*allocation.Memory)
	}

	return rt, nil
}

func (s StarlightExecutor) getPartitionBaseDir(alloc SingleJobAllocationModel) (string, error) {
	baseDir, ok := globalConfig.BaseDir[alloc.Cluster]
	if !ok {
		return "", fmt.Errorf("no matched cluster base dir")
	}
	return baseDir, nil
}

func (s StarlightExecutor) waitingJob(cluster, jobIdx string, retChan chan<- int) {
	var (
		intervalIdx   int    = 0
		intervalCount int    = 0
		errorCount    int    = 0
		query         string = fmt.Sprintf("job/%s/%s?history=true", cluster, jobIdx)
	)

	for {
		// 1. query
		var job model.Job
		_, err := s.client.GetSpec(query, &job)
		if err != nil {
			errorCount++
			if errorCount > MaxAllowErrorCount {
				// reached timeout, quit as failed
				retChan <- -1
				break
			}
		}
		// 2. end if success or failed
		if job.Status == 2 {
			if job.ExitCode == 0 {
				retChan <- -2
			} else {
				retChan <- job.ExitCode
			}
			break
		} else if job.Status == 3 {
			retChan <- 0
			break
		}
		// 3. set interval
		interval := JobQueryInterval[intervalIdx]
		intervalCount++
		if intervalCount >= 5 {
			intervalIdx++
			if intervalIdx >= 6 {
				// reached timeout, quit as failed
				retChan <- -1
				break
			}
		}
		time.Sleep(interval)
	}
}
