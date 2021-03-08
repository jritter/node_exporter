// Copyright 2019 The Prometheus Authors
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// +build !nothermalzone

package collector

import (
	"fmt"

	"github.com/go-kit/kit/log"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/procfs/sysfs"
)

type drmCollector struct {
	fs          sysfs.FS
	cardEnable  *prometheus.Desc
	cardInfo    *prometheus.Desc
	portDpms    *prometheus.Desc
	portEnabled *prometheus.Desc
	portStatus  *prometheus.Desc
	logger      log.Logger
}

func init() {
	registerCollector("drm", defaultEnabled, NewDrmCollector)
}

// NewThermalZoneCollector returns a new Collector exposing kernel/system statistics.
func NewDrmCollector(logger log.Logger) (Collector, error) {
	fs, err := sysfs.NewFS(*sysPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open sysfs: %w", err)
	}

	return &drmCollector{
		fs: fs,
		cardEnable: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "drm_card", "enable"),
			"Indicates on whether the card is enabled (1) or disabled (0)",
			[]string{"card"}, nil,
		),
		cardInfo: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "drm_card", "info"),
			"Information regarding the card",
			[]string{"card", "driver"}, nil,
		),
		portDpms: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "drm_card_port", "dpms"),
			"Display Power Management Signaling state of Port. Off = 0, On = 1",
			[]string{"port"}, nil,
		),
		portEnabled: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "drm_card_port", "enabled"),
			"Indicates on whether the port is enabled (1) or disabled (0)",
			[]string{"port"}, nil,
		),
		portStatus: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "drm_card_port", "status"),
			"Indicates on whether the port is connected to a devices or not. connected = 1, disconnected = 0",
			[]string{"port"}, nil,
		),
		logger: logger,
	}, nil
}

func (c *drmCollector) Update(ch chan<- prometheus.Metric) error {

	drmCards, err := c.fs.ClassDrmCard()
	if err != nil {
		return err
	}

	for _, stats := range drmCards {
		ch <- prometheus.MustNewConstMetric(
			c.cardEnable,
			prometheus.GaugeValue,
			float64(stats.Enable),
			stats.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.cardInfo,
			prometheus.GaugeValue,
			1,
			stats.Name,
			stats.Driver,
		)

	}

	drmCardPorts, err := c.fs.ClassDrmCardPort()
	if err != nil {
		return err
	}

	for _, stats := range drmCardPorts {
		ch <- prometheus.MustNewConstMetric(
			c.portDpms,
			prometheus.GaugeValue,
			float64(stats.Dpms),
			stats.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.portEnabled,
			prometheus.GaugeValue,
			float64(stats.Enabled),
			stats.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.portStatus,
			prometheus.GaugeValue,
			float64(stats.Status),
			stats.Name,
		)
	}

	return nil
}
