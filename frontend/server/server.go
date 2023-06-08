package server

import (
	"context"
	"encoding/json"
	"fmt"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_zap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
	"github.com/lijiang2014/cwl.go"
	"github.com/lijiang2014/cwl.go/frontend/proto"
	"github.com/lijiang2014/cwl.go/frontend/status"
	"github.com/lijiang2014/cwl.go/runner"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"net"
	"os"
	"path"
)

var (
	globalCwlServer cwlServer
	GlobalFlags     runner.EngineFlags
)

type cwlServer struct {
	status   status.StepStatusArray
	values   cwl.Values
	doc, job []byte
	fragID   string
	finish   bool

	engine *runner.Engine
}

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

	// Prevent nil
	if d == nil {
		return nil, fmt.Errorf("no cwl file specified")
	}

	// Read cwl file
	var docName string
	docName, c.fragID = splitPackedFile(d.Name)
	c.doc, err = openFileAsJSON(docName)
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
	c.job, err = openFileAsJSON(j.Name)
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
	c.engine.Flags = GlobalFlags

	// Run in go routine
	go func() {
		c.values, _ = c.engine.Run()
		c.finish = true
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
	c.engine.SendSignal(runner.SignalPause)
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
	c.engine.SendSignal(runner.SignalResume)
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
	c.engine.SendSignal(runner.SignalAbort)
	return &proto.Result{Success: true}, nil
}

func (c *cwlServer) Export(ctx context.Context, needed *proto.NotNeeded) (*proto.Status, error) {
	return status.ToGrpcStatus(&c.status), nil
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
	if s.Job == nil || s.Job.Name == "" {
		return nil, fmt.Errorf("no input file specified")
	}

	// Read input file
	c.job, err = openFileAsJSON(s.Job.Name)
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

	err = c.status.FromGrpcStatus(s)
	if err != nil {
		return nil, nil
	}
	c.status.GenerateTree()
	c.engine.ImportedStatus = &c.status

	// Run in go routine
	go func() {
		c.values, _ = c.engine.Run()
		c.finish = true
	}()

	return &proto.Result{Success: true}, nil
}

// StartServe will start a grpc server at target port
// Notice: it will block until service stop
// TODO: when will service stop?
func StartServe(port int, pem string, key string) error {
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
	err = server.Serve(listener)
	if err != nil {
		return err
	}
	return nil
}
