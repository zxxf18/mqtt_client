package main

import (
	"context"
	"fmt"
	"net/http"
	"runtime/debug"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"

	"github.com/zxxf18/mqtt_client/utils"
)

type ServerConfig struct {
	Server Server `yaml:"server" json:"server"`
}

// Server server config
type Server struct {
	Port         string            `yaml:"port" json:"port"`
	ReadTimeout  time.Duration     `yaml:"readTimeout" json:"readTimeout" default:"30s"`
	WriteTimeout time.Duration     `yaml:"writeTimeout" json:"writeTimeout" default:"30s"`
	ShutdownTime time.Duration     `yaml:"shutdownTime" json:"shutdownTime" default:"3s"`
	Certificate  utils.Certificate `yaml:",inline" json:",inline"`
}

type AdminServer struct {
	cfg    ServerConfig
	db     *DB
	router *gin.Engine
	server *http.Server
}

func NewServer(path string, db *DB) (*AdminServer, error) {
	var cfg ServerConfig
	err := utils.LoadYAML(path, &cfg)
	if err != nil {
		return nil, err
	}

	router := gin.New()
	svr := &http.Server{
		Addr:           cfg.Server.Port,
		Handler:        router,
		ReadTimeout:    cfg.Server.ReadTimeout,
		WriteTimeout:   cfg.Server.WriteTimeout,
		MaxHeaderBytes: 1 << 20,
	}
	if cfg.Server.Certificate.Cert != "" &&
		cfg.Server.Certificate.Key != "" &&
		cfg.Server.Certificate.CA != "" {
		t, err := utils.NewTLSConfigServer(utils.Certificate{
			CA:             cfg.Server.Certificate.CA,
			Cert:           cfg.Server.Certificate.Cert,
			Key:            cfg.Server.Certificate.Key,
			ClientAuthType: cfg.Server.Certificate.ClientAuthType,
		})
		if err != nil {
			return nil, err
		}
		svr.TLSConfig = t
	}
	return &AdminServer{
		cfg:    cfg,
		db:     db,
		router: router,
		server: svr,
	}, nil
}

func (a *AdminServer) Run() {
	a.initRoute()
	if a.server.TLSConfig == nil {
		if err := a.server.ListenAndServe(); err != nil {
			fmt.Println("server http stopped", err.Error())
		}
	} else {
		if err := a.server.ListenAndServeTLS("", ""); err != nil {
			fmt.Println("server https stopped", err.Error())
		}
	}
}

// Close close server
func (a *AdminServer) Close() error {
	ctx, _ := context.WithTimeout(context.Background(), a.cfg.Server.ShutdownTime)
	return a.server.Shutdown(ctx)
}

func (a *AdminServer) initRoute() {
	a.router.GET("/health", func(c *gin.Context) {
		c.JSON(200, "success")
	})
	v1 := a.router.Group("v1")
	{
		object := v1.Group("/gitbug")
		object.POST("", wrapper(a.Worker))
	}
}

func (a *AdminServer) Worker(c *gin.Context) (interface{}, error) {
	data := struct {
		Begin time.Time `json:"begin"`
		End   time.Time `json:"end"`
	}{}
	err := c.ShouldBindBodyWith(&data, binding.JSON)
	if err != nil {
		return nil, err
	}
	return a.db.List(data.Begin, data.End)
}

type HandlerFunc func(c *gin.Context) (interface{}, error)

func wrapper(handler HandlerFunc) func(c *gin.Context) {
	return func(cc *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				err, ok := r.(error)
				if !ok {
					err = fmt.Errorf("unknown error: %s", err.Error())
				}
				fmt.Println("handle a panic", string(debug.Stack()))
				cc.JSON(500, err.Error())
			}
		}()
		res, err := handler(cc)
		if err != nil {
			fmt.Println("failed to handler request")
			cc.JSON(500, err.Error())
			return
		}
		fmt.Println("process success", res)
		cc.JSON(http.StatusOK, res)
	}
}
