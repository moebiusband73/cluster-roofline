package gnuplot

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"text/template"
)

type Range struct {
	From string
	To   string
}

type Dataset struct {
	Datafile string
	Using    string
	Title    string
	Style    string
}

type Plot struct {
	Filename string
	Title    string
	Xlabel   string
	Ylabel   string
	Logscale string
	Xrange   Range
	Yrange   Range
	Style    []string
	Sets     []Dataset
}

func (p *Plot) AddData(d *Dataset, x []float64, y []float64) {

	f, err := os.Create(d.Datafile)
	if err != nil {
		log.Println("Add data:", err)
	}
	defer f.Close()

	if len(x) != len(y) {
		log.Println("x, y unequal length")
		os.Exit(1)
	}

	fmt.Fprintf(f, "# %s\n", d.Title)

	for i := 0; i < len(x); i++ {
		fmt.Fprintf(f, "%f %f\n", x[i], y[i])
	}

	if d.Using == "" {
		d.Using = "2"
	}
	if d.Style == "" {
		d.Style = "lines"
	}

	p.Sets = append(p.Sets, *d)
}

func (p *Plot) Create() {
	const gpTemplate = `
set terminal png size 1400,768 enhanced font ,16
set output '{{.Filename}}'
set title  '{{.Title}}'
set xlabel '{{.Xlabel}}'
set ylabel '{{.Ylabel}}'
set xrange [{{.Xrange.From}}:{{.Xrange.To}}]
set yrange [{{.Yrange.From}}:{{.Yrange.To}}]
set key left top
{{if .Logscale}}set logscale {{.Logscale}}
{{end}}
{{range $i, $e := .Style}} set style {{$e}}
{{end}}
plot {{range .Sets}} '{{.Datafile}}' t "{{.Title}}" w {{.Style}},
{{- end}}
`
	f, err := os.Create("gp.plot")
	if err != nil {
		log.Println("Write macro file:", err)
	}
	defer f.Close()

	t := template.Must(template.New("gpMacros").Parse(gpTemplate))
	err = t.Execute(f, p)
	if err != nil {
		log.Println("executing template:", err)
	}

	cmd := exec.Command("gnuplot", "gp.plot")
	log.Printf("Running gnuplot and waiting for it to finish...")
	err = cmd.Run()
	if err != nil {
		log.Printf("Command finished with error: %v", err)
	}
}
