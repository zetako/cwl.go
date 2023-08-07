package slex

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/lijiang2014/cwl.go"
	"github.com/lijiang2014/cwl.go/runner"
	"path"
	"regexp"
	"starlight/common/httpclient"
	"starlight/common/model"
	"strings"
	"time"
)

const (
	StarlightJobNameLimit  = 32
	MaxAllowErrorCount     = 5
	StarlightJobNameMin    = 5
	StarlightAppDocPattern = `^http.*starlight.*\.nscc-gz\.cn/api/app/app/cwl/[^/]+$`
)

var JobQueryInterval = [6]time.Duration{time.Second, time.Second * 10, time.Minute, time.Minute * 10, time.Minute * 30, time.Hour}

type StarlightExecutor struct {
	alloc      *JobAllocationModel
	ctx        context.Context
	token      string
	username   string // username is needed since we need to determine workdir
	client     *httpclient.BihuClient
	workflowID string
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
	submit.RuntimeParams.JobName = process.ID() // 临时更换为简单ID，解决job服务不支持'/'的问题；job服务修复后可以删掉
	submit.RuntimeParams.Env = process.Env()
	submit.RuntimeParams.Stdin, submit.RuntimeParams.Stdout, submit.RuntimeParams.Stderr = process.GetRedirection()

	submit.RuntimeParams.Cmd, err = process.Command()

	// container job
	var (
		dockerReq *cwl.DockerRequirement
	)
	reqs := process.Root().Process.Base().Requirements
	hints := process.Root().Process.Base().Hints
	for _, r := range hints {
		if r.ClassName() == "DockerRequirement" {
			dockerReq = r.(*cwl.DockerRequirement)
		}
	}
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
	}

	// setup mount
	// workdir itself is needed
	rt, err := s.QueryRuntime(process)
	if err != nil {
		return "", nil, err
	}
	submit.RuntimeParams.WorkDir.HostPath = rt.RootHost
	// 不需要主动设置workdir挂载，星光会挂
	//submit.RuntimeParams.Volumes = append(submit.RuntimeParams.Volumes, model.Volume{
	//	HostPath:  strings.TrimPrefix(rt.RootHost, "file://"),
	//	MountPath: strings.TrimPrefix(rt.RootHost, "file://"),
	//})
	// migrated source is needed
	for _, rec := range migrateRecord {
		if rec.IsSymLink {
			tmpID, err := uuid.NewRandom()
			if err != nil {
				return "", nil, err
			}
			submit.RuntimeParams.Volumes = append(submit.RuntimeParams.Volumes, model.Volume{
				Name:      tmpID.String(),
				HostPath:  strings.TrimPrefix(rec.Source, "file://"),
				MountPath: strings.TrimPrefix(rec.Source, "file://"),
				ReadOnly:  true,
			})
		}
	}

	// append infos
	submit.RuntimeParams.WorkflowUUID = s.workflowID
	if regexp.MustCompile(StarlightAppDocPattern).MatchString(process.RunID) {
		arr := strings.Split(process.RunID, "/")
		submit.RuntimeParams.AppName = arr[len(arr)-1]
	}

	// send req
	var job model.Job
	s.verifySubmit(&submit)
	_, err = s.client.PostSpec("/job/submit", submit, &job)
	if err != nil {
		return "", nil, err
	}

	bidChan := make(chan int)
	go s.waitingJob(job.ClusterName, job.ClusterJobID, bidChan)
	return fmt.Sprint(job.ClusterName + "/" + job.ClusterJobID), bidChan, nil
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
		workdir = path.Join(base, s.workflowID, p.Path())
		// 2.3 save the new generated work dir
		s.alloc.Set(p.PathID, SingleJobAllocationModel{
			WorkDir: model.Volume{
				HostPath: workdir,
			},
		})
	} else {
		workdir = allocation.WorkDir.HostPath
	}
	workdir = path.Join(workdir, p.Path())
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
	// get
	baseDir, ok := globalConfig.BaseDir[alloc.Cluster]
	if !ok {
		return "", fmt.Errorf("no matched base dir for cluster %s", alloc.Cluster)
	}
	// replace with username
	if strings.Contains(baseDir, "${USER}") {
		baseDir = strings.ReplaceAll(baseDir, "${USER}", s.username)
	}
	return baseDir, nil
}

func (s StarlightExecutor) waitingJob(cluster, jobIdx string, retChan chan<- int) {
	var (
		intervalIdx   int    = 0
		intervalCount int    = 0
		errorCount    int    = 0
		query         string = fmt.Sprintf("/job/running/%s/%s?history=true", cluster, jobIdx)
	)

	for {
		// 1. query
		var job model.Job
		_, err := s.client.GetSpec(query, &job)
		if err != nil {
			fmt.Println(err)
			errorCount++
			if errorCount > MaxAllowErrorCount {
				// reached timeout, quit as failed
				retChan <- -1
				break
			}
		}
		// 2. end if success or failed
		if job.Status > model.JobStatusRunning {
			if job.Status == model.JobStatusSuccess {
				retChan <- 0
			} else {
				retChan <- job.ExitCode
			}
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

func (s StarlightExecutor) verifySubmit(submit *JobSubmitModel) {
	// 1.检查JobName
	// 1.1.去掉非法字符
	newJobName := regexp.MustCompile(`[^-A-Za-z0-9_.]`).ReplaceAllString(submit.RuntimeParams.JobName, "_")
	// 1.2.最大长度
	if len(newJobName) > StarlightJobNameLimit {
		newJobName = newJobName[:32]
	}
	// 1.3.首尾不能是特殊字符
	for {
		if newJobName == "" {
			break
		}
		if newJobName[0] == '-' || newJobName[0] == '_' || newJobName[0] == '.' {
			newJobName = newJobName[1:]
			continue
		}
		tmpLen := len(newJobName) - 1
		if newJobName[tmpLen] == '-' || newJobName[tmpLen] == '_' || newJobName[tmpLen] == '.' {
			newJobName = newJobName[:tmpLen]
			continue
		}
		break
	}
	// 1.3.最小长度
	if len(newJobName) < StarlightJobNameMin {
		if newJobName == "" {
			newJobName = "wf-job"
		}
		newJobName = newJobName + "-" + fmt.Sprintf("%x", time.Now().Unix())
	}
	// 1.4.替换回去
	submit.RuntimeParams.JobName = newJobName
}
