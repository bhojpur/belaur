package cmd

// Copyright (c) 2018 Bhojpur Consulting Private Limited, India. All rights reserved.

// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:

// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.

// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

import (
	"crypto/rand"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	plcsvc "github.com/bhojpur/policy/pkg/engine"
	"github.com/golang-jwt/jwt"
	"github.com/hashicorp/go-hclog"
	"github.com/labstack/echo/v4"

	"github.com/bhojpur/belaur/pkg/utils/flag"

	"github.com/bhojpur/belaur"
	"github.com/bhojpur/belaur/pkg/handlers"
	"github.com/bhojpur/belaur/pkg/plugin"
	"github.com/bhojpur/belaur/pkg/providers/pipelines"
	rbacProvider "github.com/bhojpur/belaur/pkg/providers/rbac"
	userProvider "github.com/bhojpur/belaur/pkg/providers/user"
	"github.com/bhojpur/belaur/pkg/providers/workers"
	"github.com/bhojpur/belaur/pkg/security"
	"github.com/bhojpur/belaur/pkg/security/rbac"
	"github.com/bhojpur/belaur/pkg/services"
	"github.com/bhojpur/belaur/pkg/store"
	stamping "github.com/bhojpur/belaur/pkg/version"
	"github.com/bhojpur/belaur/pkg/workers/agent"
	"github.com/bhojpur/belaur/pkg/workers/pipeline"
	"github.com/bhojpur/belaur/pkg/workers/scheduler/belaurscheduler"
	"github.com/bhojpur/belaur/pkg/workers/server"
)

var (
	echoInstance *echo.Echo
)

const (
	dataFolder      = "data"
	pipelinesFolder = "pipelines"
	workspaceFolder = "workspace"
)

var fs *flag.FlagSet

func init() {
	// set configuration file name for run-time arguments

	// set a prefix for environment properties so they are distinct to Bhojpur Belaur
	fs = flag.NewFlagSetWithEnvPrefix(os.Args[0], "BHOJPUR", 0)

	// set the configuration filename
	fs.String("config", ".belaur_config", "this describes the name of the config file to use")

	// command line arguments
	fs.StringVar(&belaur.Cfg.ListenPort, "port", "8080", "Listen port for Bhojpur Belaur")
	fs.StringVar(&belaur.Cfg.HomePath, "home-path", "", "Path to the Bhojpur Belaur HOME folder where all data will be stored")
	fs.StringVar(&belaur.Cfg.Hostname, "hostname", "https://localhost", "The host's name under which Bhojpur Belaur is deployed at e.g.: https://app.bhojpur.net")
	fs.StringVar(&belaur.Cfg.VaultPath, "vault-path", "", "Path to the Bhojpur Belaur vault folder. By default, will be stored inside the home folder")
	fs.IntVar(&belaur.Cfg.Worker, "concurrent-worker", 2, "Number of concurrent worker the Bhojpur Belaur instance will use to execute pipelines in parallel")
	fs.StringVar(&belaur.Cfg.JwtPrivateKeyPath, "jwt-private-key-path", "", "A RSA private key used to sign JWT tokens used for Web UI authentication")
	fs.StringVar(&belaur.Cfg.CAPath, "ca-path", "", "Path where the generated CA certificate files will be saved")
	fs.BoolVar(&belaur.Cfg.DevMode, "dev", false, "If true, Bhojpur Belaur will be started in development mode. Don't use this in production!")
	fs.BoolVar(&belaur.Cfg.VersionSwitch, "version", false, "If true, will print the version and immediately exit")
	fs.BoolVar(&belaur.Cfg.Poll, "pipeline-poll", false, "If true, Bhojpur Belaur will periodically poll pipeline repositories, watch for changes and rebuild them accordingly")
	fs.IntVar(&belaur.Cfg.PVal, "pipeline-poll-interval", 1, "The interval in minutes in which to poll source repositories for changes")
	fs.StringVar(&belaur.Cfg.ModeRaw, "mode", "server", "The mode which Bhojpur Belaur should be started in. Possible options are server and worker")
	fs.StringVar(&belaur.Cfg.WorkerName, "worker-name", "", "The name of the worker which will be displayed at the primary instance. Only used in worker mode or for docker runs")
	fs.StringVar(&belaur.Cfg.WorkerHostURL, "worker-host-url", "http://127.0.0.1:8080", "The host url of a Bhojpur Belaur primary instance to connect to. Only used in worker mode or for docker runs")
	fs.StringVar(&belaur.Cfg.WorkerGRPCHostURL, "worker-grpc-host-url", "127.0.0.1:8989", "The host url of a Bhojpur Belaur primary instance gRPC interface used for worker connection. Only used in worker mode or for docker runs")
	fs.StringVar(&belaur.Cfg.WorkerSecret, "worker-secret", "", "The secret which is used to register a worker at a Bhojpur Belaur primary instance. Only used in worker mode")
	fs.StringVar(&belaur.Cfg.WorkerServerPort, "worker-server-port", "8989", "Listen port for Bhojpur Belaur primary worker gRPC communication. Only used in server mode")
	fs.StringVar(&belaur.Cfg.WorkerTags, "worker-tags", "", "Comma separated list of custom tags for this worker. Only used in worker mode")
	fs.BoolVar(&belaur.Cfg.PreventPrimaryWork, "prevent-primary-work", false, "If true, prevents the scheduler to schedule work on this Bhojpur Belaur primary instance. Only used in server mode")
	fs.BoolVar(&belaur.Cfg.AutoDockerMode, "auto-docker-mode", false, "If true, by default runs all pipelines in a docker container")
	fs.StringVar(&belaur.Cfg.DockerHostURL, "docker-host-url", "unix:///var/run/docker.sock", "Docker daemon host url which is used to build and run pipelines in a docker container")
	fs.StringVar(&belaur.Cfg.DockerRunImage, "docker-run-image", "bhojpur/belaur:latest", "Docker image repository name with tag which will be used for running pipelines in a docker container")
	fs.StringVar(&belaur.Cfg.DockerWorkerHostURL, "docker-worker-host-url", "http://127.0.0.1:8080", "The host url of the primary/worker API endpoint used for docker worker communication")
	fs.StringVar(&belaur.Cfg.DockerWorkerGRPCHostURL, "docker-worker-grpc-host-url", "127.0.0.1:8989", "The host url of the primary/worker gRPC endpoint used for docker worker communication")
	fs.BoolVar(&belaur.Cfg.RBACEnabled, "rbac-enabled", false, "Force RBAC to be enabled. Takes priority over value saved within the database")
	fs.BoolVar(&belaur.Cfg.RBACDebug, "rbac-debug", false, "Enable RBAC debug logging.")

	// Default values
	belaur.Cfg.Bolt.Mode = 0600
}

// Start initiates all components of Bhojpur Belaur and starts the server/agent.
func Start() (err error) {
	// Parse command line flags
	if err := fs.Parse(os.Args[1:]); err != nil {
		if err.Error() == "flag: help requested" {
			return nil
		}
	}

	// Check version switch
	if belaur.Cfg.VersionSwitch {
		fmt.Printf("Bhojpur Belaur %s\n", stamping.FullVersion())
		return
	}

	// Initialize shared logger
	belaur.Cfg.Logger = hclog.New(&hclog.LoggerOptions{
		Level:  hclog.Trace,
		Output: hclog.DefaultOutput,
		Name:   "Belaur",
	})

	// Determine the mode in which Bhojpur Belaur should be started
	switch belaur.Cfg.ModeRaw {
	case "server":
		belaur.Cfg.Mode = belaur.ModeServer
	case "worker":
		belaur.Cfg.Mode = belaur.ModeWorker
	default:
		belaur.Cfg.Logger.Error("unsupported mode used", "mode", belaur.Cfg.Mode)
		return errors.New("unsupported mode used")
	}

	// Find path for Bhojpur Belaur home folder if not given by parameter
	if belaur.Cfg.HomePath == "" {
		// Find executable path
		execPath, err := findExecutablePath()
		if err != nil {
			belaur.Cfg.Logger.Error("cannot find executeable path", "error", err.Error())
			return err
		}
		belaur.Cfg.HomePath = execPath
		belaur.Cfg.Logger.Debug("executeable path found", "path", execPath)
	}

	// Set data path, workspace path and pipeline path relative to home folder and create it
	// if not exist.
	belaur.Cfg.DataPath = filepath.Join(belaur.Cfg.HomePath, dataFolder)
	err = os.MkdirAll(belaur.Cfg.DataPath, 0700)
	if err != nil {
		belaur.Cfg.Logger.Error("cannot create folder", "error", err.Error(), "path", belaur.Cfg.DataPath)
		return
	}
	belaur.Cfg.PipelinePath = filepath.Join(belaur.Cfg.HomePath, pipelinesFolder)
	err = os.MkdirAll(belaur.Cfg.PipelinePath, 0700)
	if err != nil {
		belaur.Cfg.Logger.Error("cannot create folder", "error", err.Error(), "path", belaur.Cfg.PipelinePath)
		return
	}
	belaur.Cfg.WorkspacePath = filepath.Join(belaur.Cfg.HomePath, workspaceFolder)
	err = os.MkdirAll(belaur.Cfg.WorkspacePath, 0700)
	if err != nil {
		belaur.Cfg.Logger.Error("cannot create data folder", "error", err.Error(), "path", belaur.Cfg.WorkspacePath)
		return
	}

	// Check CA path
	if belaur.Cfg.CAPath == "" {
		// Set default to data folder
		belaur.Cfg.CAPath = belaur.Cfg.DataPath
	}

	// Initialize the certificate manager service
	ca, err := security.InitCA()
	if err != nil {
		belaur.Cfg.Logger.Error("cannot create CA", "error", err.Error())
		return
	}

	// Initialize store
	store, err := services.StorageService()
	if err != nil {
		return
	}

	// Initialize MemDB
	db, err := services.MemDBService(store)
	if err != nil {
		belaur.Cfg.Logger.Error("cannot initialize memdb service", "error", err.Error())
		return err
	}
	if err = db.SyncStore(); err != nil {
		return err
	}

	var jwtKey interface{}
	// Check JWT key is set
	if belaur.Cfg.JwtPrivateKeyPath == "" {
		belaur.Cfg.Logger.Warn("using auto-generated key to sign jwt tokens, do not use in production")
		jwtKey = make([]byte, 64)
		_, err = rand.Read(jwtKey.([]byte))
		if err != nil {
			belaur.Cfg.Logger.Error("error auto-generating jwt key", "error", err.Error())
			return
		}
	} else {
		keyData, err := ioutil.ReadFile(belaur.Cfg.JwtPrivateKeyPath)
		if err != nil {
			belaur.Cfg.Logger.Error("could not read jwt key file", "error", err.Error())
			return err
		}
		jwtKey, err = jwt.ParseRSAPrivateKeyFromPEM(keyData)
		if err != nil {
			belaur.Cfg.Logger.Error("could not parse jwt key file", "error", err.Error())
			return err
		}
	}
	belaur.Cfg.JWTKey = jwtKey

	// Initialize echo instance
	echoInstance = echo.New()

	// Initiating Vault
	if belaur.Cfg.VaultPath == "" {
		// Set default to data folder
		belaur.Cfg.VaultPath = belaur.Cfg.DataPath
	}
	v, err := services.DefaultVaultService()
	if err != nil {
		belaur.Cfg.Logger.Error("error initiating vault")
		return err
	}
	if err = v.LoadSecrets(); err != nil {
		belaur.Cfg.Logger.Error("error loading secrets from vault")
		return err
	}

	// Generate global worker secret if it does not exist
	_, err = v.Get(belaur.WorkerRegisterKey)
	if err != nil {
		// Secret hasn't been generated yet
		belaur.Cfg.Logger.Info("global worker registration secret has not been generated yet. Will generate it now...")
		secret := []byte(security.GenerateRandomUUIDV5())

		// Store secret in vault
		v.Add(belaur.WorkerRegisterKey, secret)
		if err := v.SaveSecrets(); err != nil {
			belaur.Cfg.Logger.Error("failed to store secret into vault", "error", err.Error())
			return err
		}
	}

	schedulerService, err := belaurscheduler.NewScheduler(belaurscheduler.Dependencies{
		Store: store,
		DB:    db,
		CA:    ca,
		PS:    &plugin.GoPlugin{},
		Vault: v,
	})
	if err != nil {
		belaur.Cfg.Logger.Error("cannot initialize scheduler", "error", err.Error())
		return err
	}
	// initiate the ticker
	schedulerService.Init()
	pipelineService := pipeline.NewBelaurPipelineService(pipeline.Dependencies{
		Scheduler: schedulerService,
	})

	rbacService, err := initRBACService(store)
	if err != nil {
		belaur.Cfg.Logger.Error("error initializing rbac service", "error", err.Error())
		return err
	}

	pipelineProvider := pipelines.NewPipelineProvider(pipelines.Dependencies{
		Scheduler:       schedulerService,
		PipelineService: pipelineService,
		SettingsStore:   store,
	})
	rbacPrv := rbacProvider.NewProvider(rbacService)
	userPrv := userProvider.NewProvider(store, rbacService)
	// initialize the worker provider
	workerProvider := workers.NewWorkerProvider(workers.Dependencies{
		Scheduler:   schedulerService,
		Certificate: ca,
	})
	// Initialize handlers
	handlerService := handlers.NewBelaurHandler(handlers.Dependencies{
		Scheduler:        schedulerService,
		PipelineService:  pipelineService,
		PipelineProvider: pipelineProvider,
		WorkerProvider:   workerProvider,
		RBACProvider:     rbacPrv,
		UserProvider:     userPrv,
		Certificate:      ca,
		RBACService:      rbacService,
		Store:            store,
	})

	err = handlerService.InitHandlers(echoInstance)
	if err != nil {
		belaur.Cfg.Logger.Error("cannot initialize handlers", "error", err.Error())
		return err
	}

	// Allocate SIG channel
	exitChan := make(chan os.Signal, 1)

	// Register the signal channel
	signal.Notify(exitChan, syscall.SIGINT, syscall.SIGTERM)

	// Start worker gRPC server.
	// We need this in both modes (server and worker) for docker worker to run.
	workerServer := server.InitWorkerServer(server.Dependencies{
		Certificate: ca,
	})
	go func() {
		if err := workerServer.Start(); err != nil {
			belaur.Cfg.Logger.Error("failed to start gRPC worker server", "error", err)
			exitChan <- syscall.SIGTERM
		}
	}()

	cleanUpFunc := func() {}
	switch belaur.Cfg.Mode {
	case belaur.ModeServer:
		// Start ticker. Periodic job to check for new plugins.
		pipelineService.InitTicker()

		// Start API server
		go func() {
			err := echoInstance.Start(":" + belaur.Cfg.ListenPort)
			if err != nil {
				belaur.Cfg.Logger.Error("failed to start echo listener", "error", err)
				exitChan <- syscall.SIGTERM
			}
		}()
	case belaur.ModeWorker:
		// Start API server
		go func() {
			err := echoInstance.Start(":" + belaur.Cfg.ListenPort)
			if err != nil {
				belaur.Cfg.Logger.Error("failed to start echo listener", "error", err)
				exitChan <- syscall.SIGTERM
			}
		}()

		// Start agent
		ag := agent.InitAgent(exitChan, schedulerService, pipelineService, store, belaur.Cfg.HomePath)
		go func() {
			cleanUpFunc, err = ag.StartAgent()
			if err != nil {
				belaur.Cfg.Logger.Error("failed to start agent", "error", err)
				exitChan <- syscall.SIGTERM
			}
		}()
	}

	// Wait for exit signal
	<-exitChan
	belaur.Cfg.Logger.Info("exit signal received. Exiting...")

	// Run clean up func
	cleanUpFunc()
	return
}

// findExecutablePath returns the absolute path for the current
// process.
func findExecutablePath() (string, error) {
	ex, err := os.Executable()
	if err != nil {
		return "", err
	}
	return filepath.Dir(ex), nil
}

func initRBACService(store store.BelaurStore) (rbac.Service, error) {
	belaur.Cfg.Logger.Info(fmt.Sprintf("rbac enabled: %v", belaur.Cfg.RBACEnabled))

	if !belaur.Cfg.RBACEnabled {
		settings, err := store.SettingsGet()
		if err != nil {
			return nil, fmt.Errorf("failed to get store settings: %w", err)
		}

		if !settings.RBACEnabled {
			belaur.Cfg.Logger.Info("rbac disabled")
			return rbac.NewNoOpService(), nil
		}
	}

	model, err := rbac.LoadModel()
	if err != nil {
		return nil, fmt.Errorf("error loading model: %w", err)
	}

	apiLookup, err := rbac.LoadAPILookup()
	if err != nil {
		return nil, fmt.Errorf("error loading rbac api lookup: %w", err)
	}

	enforcer, err := plcsvc.NewEnforcer(model, store.BhojpurStore())
	if err != nil {
		return nil, fmt.Errorf("error instantiating policy enforcer: %w", err)
	}

	enforcer.EnableLog(belaur.Cfg.RBACDebug)

	return rbac.NewEnforcerSvc(enforcer, apiLookup), nil
}
