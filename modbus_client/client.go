package modbus_client

import (
	"fmt"
	"go.uber.org/zap"
	"slices"
	"time"
)
import "github.com/simonvetter/modbus"

const RegisterInput = "INPUT"
const RegisterHolding = "HOLDING"

const TypeFloat32 = "f32"
const TypeFloat64 = "f64"
const TypeUInt32 = "u32"
const TypeUInt64 = "u64"

var units = []string{TypeFloat32, TypeFloat64, TypeUInt32, TypeUInt64}
var registers = []string{RegisterInput, RegisterHolding}

type Client struct {
	mbclient *modbus.ModbusClient
	cfg      *Config
}

type Config struct {
	Bus    Bus
	Logger *zap.SugaredLogger
}

type Bus struct {
	Configuration       modbus.ClientConfiguration `yaml:"configuration"`
	DefaultRegisterType string                     `yaml:"default_register_type"`
	DefaultRegisterUnit string                     `yaml:"default_register_unit"`
	Slaves              []Slave
}

type Slave struct {
	ID      uint8
	Name    string
	Metrics []Metric
}

type Metric struct {
	Name         string            `yaml:"name"`
	RegisterBase uint16            `yaml:"register_base,"`
	RegisterType string            `yaml:"register_type,omitempty"`
	RegisterUnit string            `yaml:"register_unit,omitempty"`
	Unit         string            `yaml:"unit,omitempty"`
	Scale        float64           `yaml:"scale,omitempty"`
	Shift        float64           `yaml:"shift,omitempty"`
	Labels       map[string]string `yaml:"labels,omitempty"`
}

func New(cfg Config) (*Client, error) {
	cl := Client{
		cfg: &cfg,
	}
	if cfg.Bus.Configuration.Timeout < time.Millisecond {
		cfg.Bus.Configuration.Timeout = time.Second
	}
	client, err := modbus.NewClient(&cfg.Bus.Configuration)
	if err != nil {
		return nil, err
	}
	for bi, v := range cfg.Bus.Slaves {
		for mi, metric := range v.Metrics {
			if metric.RegisterType == "" {
				cfg.Bus.Slaves[bi].Metrics[mi].RegisterType = cfg.Bus.DefaultRegisterType
				metric.RegisterType = cfg.Bus.DefaultRegisterType
			}
			if metric.RegisterUnit == "" {
				cfg.Bus.Slaves[bi].Metrics[mi].RegisterUnit = cfg.Bus.DefaultRegisterUnit
				metric.RegisterUnit = cfg.Bus.DefaultRegisterUnit
			}
			if metric.Scale == 0 {
				cfg.Bus.Slaves[bi].Metrics[mi].Scale = 1.0
			}
			if !slices.Contains(units, metric.RegisterUnit) {
				return nil, fmt.Errorf("metric %s register_unit[%s] type should be in %+v", metric.Name, metric.RegisterUnit, units)
			}
			if !slices.Contains(registers, metric.RegisterType) {
				return nil, fmt.Errorf("metric %s register_type[%s] type should be in %+v", metric.Name, metric.RegisterType, registers)
			}
		}
	}
	cl.mbclient = client
	err = client.Open()
	if err == nil {
		go func() {
			for {
				fmt.Printf("%s\n", cl.Run())
				time.Sleep(time.Second * 1)
			}
		}()
	}
	return &cl, err
}

func (c *Client) Run() error {
	for _, slave := range c.cfg.Bus.Slaves {
		c.mbclient.SetUnitId(slave.ID)
		for _, metric := range slave.Metrics {
			regType := modbus.INPUT_REGISTER
			if metric.RegisterType == RegisterHolding {
				regType = modbus.HOLDING_REGISTER
			}
			var value float64
			var err error
			switch metric.RegisterUnit {
			case TypeFloat32:
				v, e := c.mbclient.ReadFloat32(metric.RegisterBase, regType)
				value = float64(v)
				err = e
			case TypeFloat64:
				v, e := c.mbclient.ReadFloat64(metric.RegisterBase, regType)
				value = float64(v)
				err = e
			case TypeUInt32:
				v, e := c.mbclient.ReadFloat32(metric.RegisterBase, regType)
				value = float64(v)
				err = e
			case TypeUInt64:
				v, e := c.mbclient.ReadFloat32(metric.RegisterBase, regType)
				value = float64(v)
				err = e
			default:
				panic(fmt.Sprintf("[%s:%s]register unit[%s] type is not supported", slave.Name, metric.Name, metric.RegisterUnit))

			}
			value *= metric.Scale
			value += metric.Shift
			if err == nil {
				fmt.Printf("%s:%s %f\n", slave.Name, metric.Name, value)
			} else {
				fmt.Println(err)
			}
		}
	}
	return nil
}

func (c *Client) Close() error {
	return c.mbclient.Close()
}
