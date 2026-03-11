
## Configuration

### supported register units

* `f32`
* `f64`
* `u32`
* `u64`

### supported register types

* `INPUT` - usually contains sensor data. Default
* `HOLDING` - usually contains device settings




all units are converted to f64 coz of prometheus

```
#prometheus write protocol URL
prometheus_url: http://127.0.0.1:8480/insert/1:0/prometheus/api/v1/write
bus:
  serial: # name of the bus, will be added to label
    configuration: # this is direct github.com/simonvetter/modbus config
      url: rtu:///dev/ttyUSB0
      speed: 9600
      databits: 0
      parity: 1 # 0 no - 1 odd - 2 even
      stopbits: 1
      timeout: 1s
      tlsclientcert: null
      tlsrootcas: null
      logger: null
    # default values for metrics, if those are defined they can be skipped in metric      
    default_register_type: INPUT 
    default_register_unit: f32

    slaves:
    - id: 5
      name: powermeter
      metrics:
      - name: voltage
        register_base: 0
        unit: volt
        labels:
          phase: L1
      - name: voltage
        register_base: 2
        unit: volt
        labels:
          phase: L2
      - name: voltage
        register_base: 4
        unit: volt
        labels:
          phase: L3
      - name: current
        register_base: 6
        register_type: INPUT
        register_unit: float32
        unit: ampere
        scale: 1.1
        shift: -1.0
        labels:
          phase: L1

```