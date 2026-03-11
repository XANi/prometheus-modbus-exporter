package config

import (
	"fmt"
	"github.com/XANi/prometheus-modbus-exporter/modbus_client"
	"github.com/goccy/go-yaml"
	"github.com/simonvetter/modbus"
	"time"
)

type Config struct {
	Bus map[string]modbus_client.Bus
}

func (c *Config) GetDefaultConfig() string {
	defaultCfg := Config{
		Bus: map[string]modbus_client.Bus{
			"serial": {
				Configuration: modbus.ClientConfiguration{
					URL:      "rtu:///dev/ttyUSB0",
					Speed:    9600,
					Parity:   1,
					StopBits: 1,
					Timeout:  time.Second,
				},
				DefaultRegisterType: modbus_client.RegisterInput,
				DefaultRegisterUnit: modbus_client.TypeFloat32,
				Slaves: []modbus_client.Slave{
					{ID: 5,
						Name: "powermeter",
						Metrics: []modbus_client.Metric{
							{
								Name:         "voltage",
								Labels:       map[string]string{"phase": "L1"},
								RegisterType: "INPUT",
								RegisterBase: 0,
								RegisterUnit: modbus_client.TypeFloat32,
								Unit:         "volt",
							},
							{
								Name:         "voltage",
								Labels:       map[string]string{"phase": "L2"},
								RegisterBase: 2,
								RegisterUnit: modbus_client.TypeFloat32,
								Unit:         "volt",
							},
							{
								Name:         "voltage",
								Labels:       map[string]string{"phase": "L3"},
								RegisterBase: 4,
								RegisterUnit: modbus_client.TypeFloat32,
								Unit:         "volt",
							},
							{
								Name:         "current",
								Labels:       map[string]string{"phase": "L1"},
								RegisterBase: 6,
								RegisterUnit: modbus_client.TypeFloat32,
								Scale:        1.1,
								Shift:        -1.0,
								Unit:         "ampere",
							},
						},
					},
				},
			},
		},
	}
	out, err := yaml.Marshal(&defaultCfg)
	if err != nil {
		panic(fmt.Errorf("can't marshal [%T- %+v] into YAML: %s", defaultCfg, defaultCfg, err))
	}
	return string(out)
}
