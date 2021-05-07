package main

import (
	"context"
	"lib"
	"thin-peak/httpservice"
)

type config struct {
	Configurator    string
	Listen          string
	MgoAddr         string
	MgoColl         string
	ClickhouseAddr  string
	ClickhouseTable string
}

func (c *config) GetListenAddress() string {
	return c.Listen
}
func (c *config) GetConfiguratorAddress() string {
	return c.Configurator
}
func (c *config) CreateHandler(ctx context.Context, connectors map[httpservice.ServiceName]*httpservice.InnerService) (httpservice.HttpService, error) {

	return NewSendMessage(c.MgoAddr, c.MgoColl, c.ClickhouseAddr)
}

func main() {
	httpservice.InitNewService(lib.ServiceNameAuthentication, false, 5, &config{}, lib.ServiceNameCookieGen)
}
