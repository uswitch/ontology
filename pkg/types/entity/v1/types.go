package v1

import (
	"github.com/uswitch/ontology/pkg/types"
	"github.com/uswitch/ontology/pkg/types/entity"
)

type Computer struct{ entity.Entity }

func init() { types.RegisterType(Computer{}, "/entity/v1/computer", "/entity") }

type Classification struct{ entity.Entity }

func init() { types.RegisterType(Classification{}, "/entity/v1/classification", "/entity") }

type Service struct{ entity.Entity }

func init() { types.RegisterType(Service{}, "/entity/v1/service", "/entity") }

type Team struct{ entity.Entity }

func init() { types.RegisterType(Team{}, "/entity/v1/team", "/entity") }

type NetworkInterface struct{ entity.Entity }

func init() { types.RegisterType(NetworkInterface{}, "/entity/v1/network_interface", "/entity") }

type LoadBalancer struct{ entity.Entity }

func init() { types.RegisterType(LoadBalancer{}, "/entity/v1/load_balancer", "/entity") }

type IPV4Address struct{ entity.Entity }

func init() { types.RegisterType(IPV4Address{}, "/entity/v1/ip_v4_address", "/entity") }
