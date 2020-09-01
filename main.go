package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/moebiusband73/cluster-roofline/gnuplot"
)

type nodestat struct {
	flops float64
	memBw float64
}

func createRoof(peakMemBw float64, peakFlopsAny float64) map[string][]float64 {
	yCut := 0.01 * peakMemBw
	knee := (peakFlopsAny - yCut) / peakMemBw
	roof := make(map[string][]float64)

	roof["x"] = []float64{0.01, knee, 1000}
	roof["y"] = []float64{yCut, peakFlopsAny, peakFlopsAny}

	return roof
}

func getCluster(cluster string) ([]float64, []float64) {
	m := make(map[string]*nodestat)

	readFile, err := os.Open("./state/" + cluster + ".txt")

	if err != nil {
		log.Fatalf("failed to open file: %s", err)
	}

	fileScanner := bufio.NewScanner(readFile)
	fileScanner.Split(bufio.ScanLines)

	for fileScanner.Scan() {
		var flopsAny, memBw float64

		fields := strings.Split(fileScanner.Text(), " ")
		if len(fields) < 3 {
			continue
		}
		if fn, err := strconv.ParseFloat(fields[1], 64); err == nil {
			flopsAny = fn
		} else {
			flopsAny = 0.0
		}
		if fn, err := strconv.ParseFloat(fields[2], 64); err == nil {
			memBw = fn
		} else {
			memBw = 0.0
		}
		m[fields[0]] = &nodestat{flopsAny, memBw}
	}

	readFile.Close()

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

	return xval, yval
}

func main() {
	log.SetPrefix("roof: ")
	log.SetFlags(log.Ldate | log.Lmicroseconds | log.Llongfile)

	last := fmt.Sprintf("last updated: %s", time.Now().Format("Mon Jan 2 15:04 2006"))

	p := gnuplot.Plot{Filename: "roofline.png",
		Title:    last,
		Xlabel:   "Intensity [flops/byte]",
		Ylabel:   "Performance [MFlops/s]",
		Logscale: "xy",
		Xrange:   gnuplot.Range{From: "0.009", To: "1000"},
		Yrange:   gnuplot.Range{From: "0", To: "1600"}}

	p.Style = append(p.Style, "circle radius graph 0.008")

	roof := createRoof(100.0, 1536.0)
	p.AddData(&gnuplot.Dataset{Datafile: "siroof.dat", Title: "Meggie - simd", Style: "lines lc \"red\" lw 3"}, roof["x"], roof["y"])
	roof = createRoof(100.0, 44.0)
	p.AddData(&gnuplot.Dataset{Datafile: "scroof.dat", Title: "Meggie - scalar", Style: "lines lc \"blue\" lw 3"}, roof["x"], roof["y"])

	xval, yval := getCluster("emmy")
	p.AddData(&gnuplot.Dataset{Datafile: "nodes-emmy.dat", Title: "Emmy nodes", Style: "circles fs solid 1.0  border -1  fc \"royalblue\""}, xval, yval)
	xval, yval = getCluster("woody")
	p.AddData(&gnuplot.Dataset{Datafile: "nodes-woody.dat", Title: "Woody nodes", Style: "circles fs solid 1.0  border -1  fc \"goldenrod\""}, xval, yval)
	xval, yval = getCluster("meggie")
	p.AddData(&gnuplot.Dataset{Datafile: "nodes-meggie.dat", Title: "Meggie nodes", Style: "circles fs solid 1.0  border -1  fc \"purple\""}, xval, yval)
	p.Create()
}
