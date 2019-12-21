package v1

import (
	"github.com/uswitch/ontology/pkg/types"
	"github.com/uswitch/ontology/pkg/types/entity"
)

type Computer entity.Entity

func init() { types.RegisterType(Computer{}, "/entity/v1/computer", "/entity") }

type NetworkInterface entity.Entity

func init() { types.RegisterType(NetworkInterface{}, "/entity/v1/network_interface", "/entity") }

type LoadBalancer entity.Entity

func init() { types.RegisterType(LoadBalancer{}, "/entity/v1/load_balancer", "/entity") }

type IPV4Address entity.Entity

func init() { types.RegisterType(IPV4Address{}, "/entity/v1/ip_v4_address", "/entity") }
