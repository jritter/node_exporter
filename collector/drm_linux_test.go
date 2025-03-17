// Copyright 2015 The Prometheus Authors
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

//go:build !nodiskstats
// +build !nodiskstats

package collector

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/go-kit/log"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
)

type testDRMCollector struct {
	dsc Collector
}

func (c testDRMCollector) Collect(ch chan<- prometheus.Metric) {
	c.dsc.Update(ch)
}

func (c testDRMCollector) Describe(ch chan<- *prometheus.Desc) {
	prometheus.DescribeByCollect(c, ch)
}

func NewTestDRMCollector(logger log.Logger) (prometheus.Collector, error) {
	dsc, err := NewDrmCollector(logger)
	if err != nil {
		return testDRMCollector{}, err
	}
	return testDRMCollector{
		dsc: dsc,
	}, err
}

func TestDRM(t *testing.T) {
	*sysPath = "fixtures/sys"
	*procPath = "fixtures/proc"
	*udevDataPath = "fixtures/udev/data"

	testcase := `# HELP node_drm_card_info Card information
# TYPE node_drm_card_info gauge
node_drm_card_info{card="card0",driver="amdgpu",memory_vendor="samsung",power_performance_level="manual",unique_id="0123456789abcdef",vendor="amd"} 1
node_drm_card_info{card="card1",driver="i915",memory_vendor="",power_performance_level="",unique_id="",vendor="amd"} 1
# HELP node_drm_card_port_dpms Display Power Management Signaling state of port. Off = 0, On = 1
# TYPE node_drm_card_port_dpms gauge
node_drm_card_port_dpms{card="card1",port="DP-1"} 0
node_drm_card_port_dpms{card="card1",port="DP-5"} 1
# HELP node_drm_card_port_enabled Indicates on whether the port is enabled or disabled. enabled = 1, disabled = 0
# TYPE node_drm_card_port_enabled gauge
node_drm_card_port_enabled{card="card1",port="DP-1"} 0
node_drm_card_port_enabled{card="card1",port="DP-5"} 1
# HELP node_drm_card_port_status Indicates on whether the port is connected to a devices or not. connected = 1, disconnected = 0
# TYPE node_drm_card_port_status gauge
node_drm_card_port_status{card="card1",port="DP-1"} 0
node_drm_card_port_status{card="card1",port="DP-5"} 1
# HELP node_drm_gpu_busy_percent How busy the GPU is as a percentage.
# TYPE node_drm_gpu_busy_percent gauge
node_drm_gpu_busy_percent{card="card0"} 4
node_drm_gpu_busy_percent{card="card1"} 0
# HELP node_drm_memory_gtt_size_bytes The size of the graphics translation table (GTT) block in bytes.
# TYPE node_drm_memory_gtt_size_bytes gauge
node_drm_memory_gtt_size_bytes{card="card0"} 8.573157376e+09
node_drm_memory_gtt_size_bytes{card="card1"} 0
# HELP node_drm_memory_gtt_used_bytes The used amount of the graphics translation table (GTT) block in bytes.
# TYPE node_drm_memory_gtt_used_bytes gauge
node_drm_memory_gtt_used_bytes{card="card0"} 1.44560128e+08
node_drm_memory_gtt_used_bytes{card="card1"} 0
# HELP node_drm_memory_vis_vram_size_bytes The size of visible VRAM in bytes.
# TYPE node_drm_memory_vis_vram_size_bytes gauge
node_drm_memory_vis_vram_size_bytes{card="card0"} 8.573157376e+09
node_drm_memory_vis_vram_size_bytes{card="card1"} 0
# HELP node_drm_memory_vis_vram_used_bytes The used amount of visible VRAM in bytes.
# TYPE node_drm_memory_vis_vram_used_bytes gauge
node_drm_memory_vis_vram_used_bytes{card="card0"} 1.490378752e+09
node_drm_memory_vis_vram_used_bytes{card="card1"} 0
# HELP node_drm_memory_vram_size_bytes The size of VRAM in bytes.
# TYPE node_drm_memory_vram_size_bytes gauge
node_drm_memory_vram_size_bytes{card="card0"} 8.573157376e+09
node_drm_memory_vram_size_bytes{card="card1"} 0
# HELP node_drm_memory_vram_used_bytes The used amount of VRAM in bytes.
# TYPE node_drm_memory_vram_used_bytes gauge
node_drm_memory_vram_used_bytes{card="card0"} 1.490378752e+09
node_drm_memory_vram_used_bytes{card="card1"} 0
`

	logger := log.NewLogfmtLogger(os.Stderr)
	collector, err := NewDrmCollector(logger)
	if err != nil {
		t.Fatal(err)
	}
	c, err := NewTestDRMCollector(logger)
	if err != nil {
		t.Fatal(err)
	}
	reg := prometheus.NewRegistry()
	reg.MustRegister(c)

	sink := make(chan prometheus.Metric)
	go func() {
		err = collector.Update(sink)
		if err != nil {
			panic(fmt.Errorf("failed to update collector: %s", err))
		}
		close(sink)
	}()

	err = testutil.GatherAndCompare(reg, strings.NewReader(testcase))
	if err != nil {
		t.Fatal(err)
	}
}
