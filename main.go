package main

import (
	"encoding/json"
	"fmt"
	"time"

	_ "github.com/influxdata/influxdb1-client"
	client "github.com/influxdata/influxdb1-client/v2"
	"github.com/moebiusband73/cluster-roofline/gnuplot"
)

type nodestat struct {
	flops float64
	memBw float64
}

func main() {
	c, err := client.NewHTTPClient(client.HTTPConfig{
		Addr: "http://localhost:8086",
	})
	if err != nil {
		fmt.Println("Error creating InfluxDB Client: ", err.Error())
	}
	defer c.Close()

	qf := client.NewQuery("SELECT flops_any FROM data GROUP BY \"host\" ORDER BY time DESC LIMIT 1", "ClusterCockpit", "s")
	m := make(map[string]*nodestat)

	if response, err := c.Query(qf); err == nil && response.Error() == nil {
		for _, row := range response.Results[0].Series {
			v := row.Values[0][1].(json.Number)
			f, _ := v.Float64()
			m[row.Tags["host"]] = &nodestat{f, 0.0}
		}
	}

	qm := client.NewQuery("SELECT mem_bw FROM data GROUP BY \"host\" ORDER BY time DESC LIMIT 1", "ClusterCockpit", "s")
	if response, err := c.Query(qm); err == nil && response.Error() == nil {
		for _, row := range response.Results[0].Series {
			v := row.Values[0][1].(json.Number)
			f, _ := v.Float64()
			m[row.Tags["host"]].memBw = f
		}
	}

	xval := make([]float64, len(m))
	yval := make([]float64, len(m))
	i := 0

	for _, s := range m {
		ns := *s
		if ns.memBw == 0.0 {
			ns.memBw = 0.0001
		}
		xval[i] = ns.flops / ns.memBw
		yval[i] = ns.flops * 0.001
		i++
	}

	yCut := 0.01 * 80.0
	scalarKnee := (44.0 - yCut) / 80.0
	simdKnee := (704.0 - yCut) / 80.0

	xRoofSimd := []float64{0.01, simdKnee, 1000}
	yRoofSimd := []float64{yCut, 704.0, 704.0}
	xRoofScalar := []float64{0.01, scalarKnee, 1000}
	yRoofScalar := []float64{yCut, 44.0, 44.0}
	last := fmt.Sprintf("last updated: %s", time.Now().Format("Mon Jan 2 15:04 2006"))

	p := gnuplot.Plot{Filename: "roofline.png",
		Title:    last,
		Xlabel:   "Intensity [flops/byte]",
		Ylabel:   "Performance [MFlops/s]",
		Logscale: "x, y",
		Xrange:   gnuplot.Range{From: "0", To: "1000"},
		Yrange:   gnuplot.Range{From: "0", To: "1000"}}

	p.AddData(&gnuplot.Dataset{Datafile: "siroof.dat", Title: "simd", Style: "lines"}, xRoofSimd, yRoofSimd)
	p.AddData(&gnuplot.Dataset{Datafile: "scroof.dat", Title: "scalar", Style: "lines"}, xRoofScalar, yRoofScalar)
	p.AddData(&gnuplot.Dataset{Datafile: "nodes.dat", Title: "", Style: "points ls 1"}, xval, yval)
	p.Create()
}
