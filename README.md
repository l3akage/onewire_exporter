# onewire_exporter

Prometheus exporter for 1-wire temperature sensors connected to a Raspberry PI

# Usage

```
./onewire_exporter
```

# Options

Name     | Default | Description
---------|-------------|----
--version || Print version information
--listen-address | :9330 | Address on which to expose metrics.
--path | /metrics | Path under which to expose metrics.
--ignoreUnknown |true | Ignores sensors without a name
--names | names.yaml | File mapping IDs to names"


# Configuration

The names.yaml files contains the sensor id and name to use
```
names:
  28-0416506c85ff: temp_sensor_01
```

# Setup
For gpio 4 append this to /boot/config.txt and reboot

```
dtoverlay=w1-gpio-pullup,gpiopin=4,extpullup=on
```

You should now see your sensors with IDs in /sys/bus/w1/devices

# Example output
```
# HELP onewire_temp Air temperature (in degrees C)
# TYPE onewire_temp gauge
onewire_temp{id="28-0416506c85ff",name="temp_sensor_01"} 28.187
```

# Circuit

I used DS18b20 1-wire sensors. For a long distance (50m worked fine) i used 5v power for the sensors and 3.3v with a 4.7k resistor as pull-up.

![Circuit](https://m0u.de/assets/onewire_circuit.png)
