package server

import (
	"context"
	"encoding/json"
	"fmt"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_zap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
	"github.com/lijiang2014/cwl.go"
	"github.com/lijiang2014/cwl.go/frontend/proto"
	"github.com/lijiang2014/cwl.go/intergration/client"
	"github.com/lijiang2014/cwl.go/intergration/sfs"
	"github.com/lijiang2014/cwl.go/intergration/slex"
	"github.com/lijiang2014/cwl.go/message"
	"github.com/lijiang2014/cwl.go/runner"
	"github.com/zetako/scontrol"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"log"
	"net"
	"os"
	"path"
)

var (
	globalCwlServer cwlServer
	GlobalFlags     runner.EngineFlags
)

type cwlServer struct {
	status     *message.StepStatusArray
	values     cwl.Values
	doc, job   []byte
	fragID     string
	finish     bool
	workflowID string

	token        string
	clientConfig client.Config
	importer     runner.Importer
	engine       *runner.Engine
}

// Load is grpc method Load. It will generate a new engine and load a doc using Importer.
//
// ⚠️Notice: the old engine and all running task will be immediately discard!
func (c *cwlServer) Load(ctx context.Context, d *proto.Doc) (result *proto.Result, err error) {
	defer func() {
		// if there's error, this function help generate result
		if err != nil && result == nil {
			result = &proto.Result{
				Success: false,
				Info:    getStringPointer(err.Error()),
			}
		}
	}()

	// Clear all previous infos
	c.status = &message.StepStatusArray{}
	c.values = nil
	c.finish = false

	// Prevent nil
	if d == nil {
		return nil, fmt.Errorf("no cwl file specified")
	}

	// New Importer
	c.token = d.Token
	tmpClient, err := c.generateStarlightClient()
	if err != nil {
		return nil, err
	}
	c.importer, err = sfs.New(context.TODO(), d.Token, tmpClient, "", false)
	if err != nil {
		return nil, err
	}

	// Read cwl file
	var docName string
	docName, c.fragID = splitPackedFile(d.Name)
	doc, err := c.importer.Load(docName)
	if err != nil {
		return nil, err
	}
	c.doc, err = cwl.Y2J(doc)
	if err != nil {
		return nil, err
	}
	return &proto.Result{Success: true}, nil
}

func (c *cwlServer) Start(ctx context.Context, j *proto.Job) (result *proto.Result, err error) {
	defer func() {
		// if there's error, this function help generate result
		if err != nil && result == nil {
			result = &proto.Result{
				Success: false,
				Info:    getStringPointer(err.Error()),
			}
		}
	}()
	// Prevent nil
	if j == nil {
		return nil, fmt.Errorf("no cwl file specified")
	}

	// Read input file
	c.job, err = cwl.Y2J([]byte(j.Values))
	if err != nil {
		return nil, err
	}

	// Create engine
	pwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	c.engine, err = runner.NewEngine(runner.EngineConfig{
		DocumentID:   c.fragID,
		RunID:        "cwl.go",
		Process:      c.doc,
		Params:       c.job,
		DocImportDir: pwd,
		RootHost:     pwd,
		InputsDir:    path.Join(pwd, "inputs"),
		WorkDir:      path.Join(pwd, "run"),
		Importer:     c.importer,
		NewFSMethod: func(workdir string) (runner.Filesystem, error) {
			tmpClient, err := c.generateStarlightClient()
			if err != nil {
				return nil, err
			}
			fs, err := sfs.New(context.TODO(), c.token, tmpClient, workdir, true)
			if err != nil {
				return nil, err
			}
			return fs, nil
		},
	})
	if err != nil {
		return nil, err
	}
	// generate executor
	tmpClient, err := c.generateStarlightClient()
	if err != nil {
		return nil, err
	}
	exec, err := slex.New(context.TODO(), j.Id, tmpClient, j.Username, FromGrpcAllocation(j.Allocations))
	if err != nil {
		return nil, err
	}
	c.engine.SetDefaultExecutor(exec)
	c.engine.MessageReceiver = serverMsgReceiver{}
	c.engine.Flags = GlobalFlags

	// Run in go routine
	go func() {
		c.status.Append(&message.StepStatus{
			ID:     message.PathID{"root"},
			Status: message.StatusStart,
		})
		c.values, _ = c.engine.Run()
		c.finish = true
		c.status.Append(&message.StepStatus{
			ID:     message.PathID{"root"},
			Status: message.StatusFinish,
		})
	}()

	return &proto.Result{Success: true}, nil
}

func (c *cwlServer) Out(ctx context.Context, needed *proto.NotNeeded) (*proto.Output, error) {
	if c.finish {
		tmp, _ := json.Marshal(c.values)
		return &proto.Output{
			Result: &proto.Result{Success: true},
			Values: string(tmp),
		}, nil
	} else {
		return &proto.Output{
			Result: &proto.Result{
				Success: false,
				Info:    getStringPointer("Workflow not finished"),
			},
		}, nil
	}
}

func (c *cwlServer) Pause(ctx context.Context, needed *proto.NotNeeded) (*proto.Result, error) {
	if c.engine == nil {
		return &proto.Result{
			Success: false,
			Info:    getStringPointer("Workflow not started"),
		}, nil
	}
	if c.finish {
		return &proto.Result{
			Success: false,
			Info:    getStringPointer("Workflow already finished"),
		}, nil
	}
	c.engine.SendSignal(scontrol.StatusPause)
	return &proto.Result{Success: true}, nil
}

func (c *cwlServer) Resume(ctx context.Context, needed *proto.NotNeeded) (*proto.Result, error) {
	if c.engine == nil {
		return &proto.Result{
			Success: false,
			Info:    getStringPointer("Workflow not started"),
		}, nil
	}
	if c.finish {
		return &proto.Result{
			Success: false,
			Info:    getStringPointer("Workflow already finished"),
		}, nil
	}
	c.engine.SendSignal(scontrol.StatusRunning)
	return &proto.Result{Success: true}, nil
}

func (c *cwlServer) Abort(ctx context.Context, needed *proto.NotNeeded) (*proto.Result, error) {
	if c.engine == nil {
		return &proto.Result{
			Success: false,
			Info:    getStringPointer("Workflow not started"),
		}, nil
	}
	if c.finish {
		return &proto.Result{
			Success: false,
			Info:    getStringPointer("Workflow already finished"),
		}, nil
	}
	c.engine.SendSignal(scontrol.StatusStop)
	return &proto.Result{Success: true}, nil
}

func (c *cwlServer) Export(ctx context.Context, needed *proto.NotNeeded) (*proto.Status, error) {
	ret := ToGrpcStatus(c.status)
	ret.Finish = c.finish
	return ret, nil
}

func (c *cwlServer) Import(ctx context.Context, s *proto.Status) (result *proto.Result, err error) {
	defer func() {
		// if there's error, this function help generate result
		if err != nil && result == nil {
			result = &proto.Result{
				Success: false,
				Info:    getStringPointer(err.Error()),
			}
		}
	}()
	// Prevent nil
	if s.Job == nil || s.Job.Values == "" {
		return nil, fmt.Errorf("no input specified")
	}

	// Read input file
	c.job, err = cwl.Y2J([]byte(s.Job.Values))
	if err != nil {
		return nil, err
	}

	// Create engine
	pwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	c.engine, err = runner.NewEngine(runner.EngineConfig{
		DocumentID:   c.fragID,
		RunID:        "cwl.go",
		Process:      c.doc,
		Params:       c.job,
		DocImportDir: pwd,
		RootHost:     pwd,
		InputsDir:    path.Join(pwd, "inputs"),
		WorkDir:      path.Join(pwd, "run"),
	})
	if err != nil {
		return nil, err
	}
	c.engine.SetDefaultExecutor(&runner.LocalExecutor{})
	c.engine.MessageReceiver = serverMsgReceiver{}

	c.status, err = FromGrpcStatus(s)
	if err != nil {
		return nil, nil
	}
	c.status.GenerateTree()
	c.engine.ImportedStatus = c.status

	// Run in go routine
	go func() {
		c.values, _ = c.engine.Run()
		c.finish = true
	}()

	return &proto.Result{Success: true}, nil
}

func (c *cwlServer) Watch(server proto.Cwl_WatchServer) error {
	return fmt.Errorf("not implemented")
}

// StartServe will start a grpc server at target port
// Notice: it will block until service stop
// TODO: when will service stop?
func StartServe(port int, pem string, key string, conf client.Config) error {
	// set client config
	globalCwlServer.clientConfig = conf

	// Options
	var opts []grpc.ServerOption

	// Set logger
	// Need a better logger
	logger, err := zap.NewDevelopment()
	if err != nil {
		return err
	}
	grpc_zap.ReplaceGrpcLoggerV2(logger)
	opts = append(opts,
		grpc.UnaryInterceptor(
			grpc_middleware.ChainUnaryServer(
				grpc_zap.UnaryServerInterceptor(logger),
			),
		),
	)

	// Set credential
	if pem != "" && key != "" {
		cred, err := credentials.NewServerTLSFromFile(pem, key)
		if err != nil {
			return err
		}
		opts = append(opts, grpc.Creds(cred))
	}

	// Use logger to init server
	server := grpc.NewServer(opts...)
	proto.RegisterCwlServer(server, &globalCwlServer)

	// Start Listening
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return err
	}
	log.Printf("grpc server listen at port %d", port)
	err = server.Serve(listener)
	if err != nil {
		return err
	}
	return nil
}
