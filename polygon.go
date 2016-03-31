package polygen

import (
	"fmt"
	"github.com/llgcode/draw2d/draw2dimg"
	"image"
	"image/color"
	"log"
	"math/rand"
)

const (
	MutationColor            = iota
	MutationPoint            = iota
	MutationAddOrDeletePoint = iota
)

const (
	MutationChance        = 0.15
	PopulationCount       = 10
	PolygonsPerIndividual = 100
	MaxPolygonPoints      = 6
	MinPolygonPoints      = 3
)

var (
	Mutations = []int{MutationColor, MutationPoint, MutationAddOrDeletePoint}
)

type Candidate struct {
	w, h     int
	Polygons []*Polygon
	Fitness  int64
	img      *image.RGBA
}

type Point struct {
	X, Y int
}

type Polygon struct {
	Points []*Point
	color.Color
}

func RandomCandidate(w, h int) *Candidate {
	result := &Candidate{w: w, h: h, Polygons: make([]*Polygon, PolygonsPerIndividual)}
	for j := 0; j < len(result.Polygons); j++ {
		result.Polygons[j] = RandomPolygon(w, h)
	}

	result.RenderImage()

	return result
}

func RandomPolygon(maxW, maxH int) *Polygon {
	result := &Polygon{}
	result.Color = color.RGBA{uint8(rand.Intn(0xff)), uint8(rand.Intn(0xff)), uint8(rand.Intn(0xff)), uint8(rand.Intn(0xff))}

	numPoints := RandomInt(MinPolygonPoints, MaxPolygonPoints)

	for i := 0; i < numPoints; i++ {
		result.AddPoint(RandomPoint(maxW, maxH))
	}

	return result
}

func RandomPoint(maxW, maxH int) *Point {
	return &Point{rand.Intn(maxW), rand.Intn(maxH)}
}

func (m1 *Candidate) Mate(m2 *Candidate) *Candidate {
	w, h := m1.w, m1.h
	crossover := rand.Intn(len(m1.Polygons))
	polygons := make([]*Polygon, len(m1.Polygons))

	for i := 0; i < len(polygons); i++ {
		var p Polygon

		if i <= crossover {
			p = *m1.Polygons[i] // NB copy the polygon, not the pointer
		} else {
			p = *m2.Polygons[i]
		}

		if shouldMutate() {
			p.Mutate(w, h)
		}

		polygons[i] = &p
	}

	result := &Candidate{w: m1.w, h: m1.h, Polygons: polygons}
	result.RenderImage()
	return result
}

func (p *Polygon) AddPoint(point *Point) {
	p.Points = append(p.Points, point)
}

func (p *Polygon) Mutate(maxW, maxH int) {
	switch randomMutation() {
	case MutationColor:
		orig := p.Color
		p.Color = MutateColor(p.Color)
		log.Printf("MutationColor: %v -> %v", orig, p.Color)

	case MutationPoint:
		i := rand.Intn(len(p.Points))
		orig := *p.Points[i]
		mutated := MutatePoint(orig, maxW, maxH)
		p.Points[i] = &mutated
		log.Printf("MutationPoint: %v -> %v", orig, mutated)

	case MutationAddOrDeletePoint:
		origPointCount := len(p.Points)

		if len(p.Points) == MinPolygonPoints {
			// can't delete
			p.AddPoint(RandomPoint(maxW, maxH))
		} else if len(p.Points) == MaxPolygonPoints {
			// can't add
			p.DeleteRandomPoint()
		} else {
			// we can do either add or delete
			switch rand.Intn(2) {
			case 0:
				p.AddPoint(RandomPoint(maxW, maxH))
			case 1:
				p.DeleteRandomPoint()
			}
		}

		newPointCount := len(p.Points)

		log.Printf("MutationAddOrDeletePoint: %d -> %d points", origPointCount, newPointCount)
	}
}

func (p *Polygon) DeleteRandomPoint() {
	i := rand.Intn(len(p.Points))
	p.Points = append(p.Points[:i], p.Points[i+1:]...)
}

// NB: operates on copy of p
func MutatePoint(p Point, maxW, maxH int) Point {
	result := p

	i := rand.Intn(2)
	switch i {
	case 0:
		result.X = rand.Intn(maxW)
	case 1:
		result.Y = rand.Intn(maxH)
	}

	return result
}

func MutateColor(c color.Color) color.Color {
	// get the non-premultiplied rgba values
	nrgba := color.NRGBAModel.Convert(c).(color.NRGBA)

	// randomly select one of the r/g/b/a values to mutate
	i := rand.Intn(4)
	val := uint8(rand.Intn(256))

	switch i {
	case 0:
		nrgba.R = val
	case 1:
		nrgba.G = val
	case 2:
		nrgba.B = val
	case 3:
		nrgba.A = val
	}

	return color.RGBAModel.Convert(nrgba)
}

func randomMutation() int {
	return Mutations[rand.Int()%len(Mutations)]
}

func (cd *Candidate) RenderImage() {
	cd.img = image.NewRGBA(image.Rect(0, 0, cd.w, cd.h))
	gc := draw2dimg.NewGraphicContext(cd.img)

	gc.SetLineWidth(1)

	for _, polygon := range cd.Polygons {
		gc.SetStrokeColor(polygon.Color)
		gc.SetFillColor(polygon.Color)

		firstPoint := polygon.Points[0]
		gc.MoveTo(float64(firstPoint.X), float64(firstPoint.Y))

		for _, point := range polygon.Points[1:] {
			gc.LineTo(float64(point.X), float64(point.Y))
		}

		gc.Close()
		gc.FillStroke()
	}
}

func (cd *Candidate) DrawAndSave(destFile string) {
	cd.RenderImage()
	draw2dimg.SaveToPngFile(destFile, cd.img)
}

func (cd *Candidate) String() string {
	return fmt.Sprintf("fitness: %d", cd.Fitness)
}

func shouldMutate() bool {
	return rand.Float32() < MutationChance
}

type ByFitness []*Candidate

func (cds ByFitness) Len() int           { return len(cds) }
func (cds ByFitness) Swap(i, j int)      { cds[i], cds[j] = cds[j], cds[i] }
func (cds ByFitness) Less(i, j int) bool { return cds[i].Fitness < cds[j].Fitness }
