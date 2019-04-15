package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
)

const prefix = "onewire_"

var (
	upDesc   *prometheus.Desc
	tempDesc *prometheus.Desc
)

func init() {
	upDesc = prometheus.NewDesc(prefix+"up", "Scrape was successful", nil, nil)
	tempDesc = prometheus.NewDesc(prefix+"temp", "Air temperature (in degrees C)", []string{"id", "name"}, nil)
}

type Temp struct {
	ID    string
	Value float64
}

type onewireCollector struct {
}

func getTemperatureFromDevice(device os.FileInfo) Temp {
	reg, err := regexp.Compile("[^0-9]+")
	if err != nil {
		log.Fatal(err)
	}
	for i := 1; i <= 5; i++ {
		content, err := ioutil.ReadFile("/sys/bus/w1/devices/" + device.Name() + "/w1_slave")
		if err != nil {
			log.Infof("Error reading device %s\n", device.Name())
			continue
		}
		lines := strings.Split(string(content), "\n")
		if len(lines) != 3 {
			log.Infof("Unknown format for device %s\n", device.Name())
			continue
		}
		if !strings.Contains(lines[0], "YES") {
			log.Infof("CRC invalid for device %s\n", device.Name())
			continue
		}
		data := strings.SplitAfter(lines[1], "t=")
		if len(data) != 2 {
			log.Infof("Temp value not found for device %s\n", device.Name())
			continue
		}
		strValue := reg.ReplaceAllString(data[1], "")

		tempInt, err := strconv.ParseFloat(strValue, 64)
		if err != nil {
			continue
		}
		if tempInt == 85000 {
			continue
		}
		return Temp{
			ID:    device.Name(),
			Value: tempInt / 1000.0,
		}
	}
	return Temp{}
}

func getTemperatures() ([]Temp, error) {
	devices, err := ioutil.ReadDir("/sys/bus/w1/devices/")
	if err != nil {
		return nil, err
	}
	var wg sync.WaitGroup
	valueChan := make(chan Temp)
	for _, device := range devices {
		if _, err := os.Stat("/sys/bus/w1/devices/" + device.Name() + "/w1_slave"); err != nil {
			continue
		}
		wg.Add(1)
		go func(device os.FileInfo) {
			defer wg.Done()
			valueChan <- getTemperatureFromDevice(device)
		}(device)
	}
	go func() {
		wg.Wait()
		close(valueChan)
	}()
	var values []Temp
	for t := range valueChan {
		if t == (Temp{}) {
			continue
		}
		values = append(values, t)
	}
	return values, nil
}

func (c onewireCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- upDesc
	ch <- tempDesc
}

func (c onewireCollector) Collect(ch chan<- prometheus.Metric) {
	values, err := getTemperatures()
	if err != nil {
		fmt.Fprintln(os.Stderr, "error getting sensor data", err)
		ch <- prometheus.MustNewConstMetric(upDesc, prometheus.GaugeValue, 0)
	} else {
		for _, sensor := range values {
			n := list.Names[sensor.ID]
			if n == "" {
				if *ignoreUnknown == true {
					log.Infof("Ingoring unknown device %s\n", sensor.ID)
					continue
				} else {
					n = sensor.ID
				}
			}
			l := []string{sensor.ID, n}
			ch <- prometheus.MustNewConstMetric(tempDesc, prometheus.GaugeValue, float64(sensor.Value), l...)
		}
	}
}
