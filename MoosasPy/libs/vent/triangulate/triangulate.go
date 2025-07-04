package main

import (
	"fmt"
	"io"
	"math"
	"os"
	"sort"
	"strings"
	"strconv"
)

func main() {
	input,_:=os.ReadFile("triangulate.input")
	coordinates:=strings.Split(string(input),",")
	points:=make([]Point,len(coordinates)/2)
	for i:=0;i<len(points);i++ {
		x,_:=strconv.ParseFloat(coordinates[i*2+0],64)
		y,_:=strconv.ParseFloat(coordinates[i*2+1],64)
		points[i]=Point{x,y}
	}
	triangulation, _ := Triangulate(points)
	var output string
	for _,v:= range triangulation.Triangles {
		output+=strconv.FormatInt(int64(v),10)
		output+=","
	}
	os.Remove("triangulate.output")
	file,_:=os.Create("triangulate.output")
	io.WriteString(file,output[:len(output)-1])
}

func cross2D(p, a, b Point) float64 {
	return (a.X-p.X)*(b.Y-p.Y) - (a.Y-p.Y)*(b.X-p.X)
}

// ConvexHull returns the convex hull of the provided points.
func ConvexHull(points []Point) []Point {
	// copy points
	pointsCopy := make([]Point, len(points))
	copy(pointsCopy, points)
	points = pointsCopy

	// sort points
	sort.Slice(points, func(i, j int) bool {
		a := points[i]
		b := points[j]
		if a.X != b.X {
			return a.X < b.X
		}
		return a.Y < b.Y
	})

	// filter nearly-duplicate points
	distinctPoints := points[:0]
	for i, p := range points {
		if i > 0 && p.squaredDistance(points[i-1]) < eps {
			continue
		}
		distinctPoints = append(distinctPoints, p)
	}
	points = distinctPoints

	// find upper and lower portions
	var U, L []Point
	for _, p := range points {
		for len(U) > 1 && cross2D(U[len(U)-2], U[len(U)-1], p) > 0 {
			U = U[:len(U)-1]
		}
		for len(L) > 1 && cross2D(L[len(L)-2], L[len(L)-1], p) < 0 {
			L = L[:len(L)-1]
		}
		U = append(U, p)
		L = append(L, p)
	}

	// reverse upper portion
	for i, j := 0, len(U)-1; i < j; i, j = i+1, j-1 {
		U[i], U[j] = U[j], U[i]
	}

	// construct complete hull
	if len(U) > 0 {
		U = U[:len(U)-1]
	}
	if len(L) > 0 {
		L = L[:len(L)-1]
	}
	return append(L, U...)
}

type node struct {
	i    int
	t    int
	prev *node
	next *node
}

func newNode(nodes []node, i int, prev *node) *node {
	n := &nodes[i]
	n.i = i
	if prev == nil {
		n.prev = n
		n.next = n
	} else {
		n.next = prev.next
		n.prev = prev
		prev.next.prev = n
		prev.next = n
	}
	return n
}

func (n *node) remove() *node {
	n.prev.next = n.next
	n.next.prev = n.prev
	n.i = -1
	return n.prev
}

type Point struct {
	X, Y float64
}

func (a Point) squaredDistance(b Point) float64 {
	dx := a.X - b.X
	dy := a.Y - b.Y
	return dx*dx + dy*dy
}

func (a Point) distance(b Point) float64 {
	return math.Hypot(a.X-b.X, a.Y-b.Y)
}

func (a Point) sub(b Point) Point {
	return Point{a.X - b.X, a.Y - b.Y}
}

type Triangulation struct {
	Points     []Point
	ConvexHull []Point
	Triangles  []int
	Halfedges  []int
}

// Triangulate returns a Delaunay triangulation of the provided points.
func Triangulate(points []Point) (*Triangulation, error) {
	t := newTriangulator(points)
	err := t.triangulate()
	return &Triangulation{points, t.convexHull(), t.triangles, t.halfedges}, err
}

func (t *Triangulation) area() float64 {
	var result float64
	points := t.Points
	ts := t.Triangles
	for i := 0; i < len(ts); i += 3 {
		p0 := points[ts[i+0]]
		p1 := points[ts[i+1]]
		p2 := points[ts[i+2]]
		result += area(p0, p1, p2)
	}
	return result / 2
}

// Validate performs several sanity checks on the Triangulation to check for
// potential errors. Returns nil if no issues were found. You normally
// shouldn't need to call this function but it can be useful for debugging.
func (t *Triangulation) Validate() error {
	// verify halfedges
	for i1, i2 := range t.Halfedges {
		if i1 != -1 && t.Halfedges[i1] != i2 {
			return fmt.Errorf("invalid halfedge connection")
		}
		if i2 != -1 && t.Halfedges[i2] != i1 {
			return fmt.Errorf("invalid halfedge connection")
		}
	}

	// verify convex hull area vs sum of triangle areas
	hull1 := t.ConvexHull
	hull2 := ConvexHull(t.Points)
	area1 := polygonArea(hull1)
	area2 := polygonArea(hull2)
	area3 := t.area()
	if math.Abs(area1-area2) > 1e-9 || math.Abs(area1-area3) > 1e-9 {
		return fmt.Errorf("hull areas disagree: %f, %f, %f", area1, area2, area3)
	}

	// verify convex hull perimeter
	perimeter1 := polygonPerimeter(hull1)
	perimeter2 := polygonPerimeter(hull2)
	if math.Abs(perimeter1-perimeter2) > 1e-9 {
		return fmt.Errorf("hull perimeters disagree: %f, %f", perimeter1, perimeter2)
	}

	return nil
}

type triangulator struct {
	points           []Point
	squaredDistances []float64
	ids              []int
	center           Point
	triangles        []int
	halfedges        []int
	trianglesLen     int
	hull             *node
	hash             []*node
}

func newTriangulator(points []Point) *triangulator {
	return &triangulator{points: points}
}

// sorting a triangulator sorts the `ids` such that the referenced points
// are in order by their distance to `center`
func (a *triangulator) Len() int {
	return len(a.points)
}

func (a *triangulator) Swap(i, j int) {
	a.ids[i], a.ids[j] = a.ids[j], a.ids[i]
}

func (a *triangulator) Less(i, j int) bool {
	d1 := a.squaredDistances[a.ids[i]]
	d2 := a.squaredDistances[a.ids[j]]
	if d1 != d2 {
		return d1 < d2
	}
	p1 := a.points[a.ids[i]]
	p2 := a.points[a.ids[j]]
	if p1.X != p2.X {
		return p1.X < p2.X
	}
	return p1.Y < p2.Y
}

func (tri *triangulator) triangulate() error {
	points := tri.points

	n := len(points)
	if n == 0 {
		return nil
	}

	tri.ids = make([]int, n)

	// compute bounds
	x0 := points[0].X
	y0 := points[0].Y
	x1 := points[0].X
	y1 := points[0].Y
	for i, p := range points {
		if p.X < x0 {
			x0 = p.X
		}
		if p.X > x1 {
			x1 = p.X
		}
		if p.Y < y0 {
			y0 = p.Y
		}
		if p.Y > y1 {
			y1 = p.Y
		}
		tri.ids[i] = i
	}

	var i0, i1, i2 int

	// pick a seed point close to midpoint
	m := Point{(x0 + x1) / 2, (y0 + y1) / 2}
	minDist := infinity
	for i, p := range points {
		d := p.squaredDistance(m)
		if d < minDist {
			i0 = i
			minDist = d
		}
	}

	// find point closest to seed point
	minDist = infinity
	for i, p := range points {
		if i == i0 {
			continue
		}
		d := p.squaredDistance(points[i0])
		if d > 0 && d < minDist {
			i1 = i
			minDist = d
		}
	}

	// find the third point which forms the smallest circumcircle
	minRadius := infinity
	for i, p := range points {
		if i == i0 || i == i1 {
			continue
		}
		r := circumradius(points[i0], points[i1], p)
		if r < minRadius {
			i2 = i
			minRadius = r
		}
	}
	if minRadius == infinity {
		return fmt.Errorf("No Delaunay triangulation exists for this input.")
	}

	// swap the order of the seed points for counter-clockwise orientation
	if area(points[i0], points[i1], points[i2]) < 0 {
		i1, i2 = i2, i1
	}

	tri.center = circumcenter(points[i0], points[i1], points[i2])

	// sort the points by distance from the seed triangle circumcenter
	tri.squaredDistances = make([]float64, n)
	for i, p := range points {
		tri.squaredDistances[i] = p.squaredDistance(tri.center)
	}
	sort.Sort(tri)

	// initialize a hash table for storing edges of the advancing convex hull
	hashSize := int(math.Ceil(math.Sqrt(float64(n))))
	tri.hash = make([]*node, hashSize)

	// initialize a circular doubly-linked list that will hold an advancing convex hull
	nodes := make([]node, n)

	e := newNode(nodes, i0, nil)
	e.t = 0
	tri.hashEdge(e)

	e = newNode(nodes, i1, e)
	e.t = 1
	tri.hashEdge(e)

	e = newNode(nodes, i2, e)
	e.t = 2
	tri.hashEdge(e)

	tri.hull = e

	maxTriangles := 2*n - 5
	tri.triangles = make([]int, maxTriangles*3)
	tri.halfedges = make([]int, maxTriangles*3)

	tri.addTriangle(i0, i1, i2, -1, -1, -1)

	pp := Point{infinity, infinity}
	for k := 0; k < n; k++ {
		i := tri.ids[k]
		p := points[i]

		// skip nearly-duplicate points
		if p.squaredDistance(pp) < eps {
			continue
		}
		pp = p

		// skip seed triangle points
		if i == i0 || i == i1 || i == i2 {
			continue
		}

		// find a visible edge on the convex hull using edge hash
		var start *node
		key := tri.hashKey(p)
		for j := 0; j < len(tri.hash); j++ {
			start = tri.hash[key]
			if start != nil && start.i >= 0 {
				break
			}
			key++
			if key >= len(tri.hash) {
				key = 0
			}
		}
		start = start.prev

		e := start
		for area(p, points[e.i], points[e.next.i]) >= 0 {
			e = e.next
			if e == start {
				e = nil
				break
			}
		}
		if e == nil {
			// likely a near-duplicate point; skip it
			continue
		}
		walkBack := e == start

		// add the first triangle from the point
		t := tri.addTriangle(e.i, i, e.next.i, -1, -1, e.t)
		e.t = t // keep track of boundary triangles on the hull
		e = newNode(nodes, i, e)

		// recursively flip triangles from the point until they satisfy the Delaunay condition
		e.t = tri.legalize(t + 2)

		// walk forward through the hull, adding more triangles and flipping recursively
		q := e.next
		for area(p, points[q.i], points[q.next.i]) < 0 {
			t = tri.addTriangle(q.i, i, q.next.i, q.prev.t, -1, q.t)
			q.prev.t = tri.legalize(t + 2)
			tri.hull = q.remove()
			q = q.next
		}

		if walkBack {
			// walk backward from the other side, adding more triangles and flipping
			q := e.prev
			for area(p, points[q.prev.i], points[q.i]) < 0 {
				t = tri.addTriangle(q.prev.i, i, q.i, -1, q.t, q.prev.t)
				tri.legalize(t + 2)
				q.prev.t = t
				tri.hull = q.remove()
				q = q.prev
			}
		}

		// save the two new edges in the hash table
		tri.hashEdge(e)
		tri.hashEdge(e.prev)
	}

	tri.triangles = tri.triangles[:tri.trianglesLen]
	tri.halfedges = tri.halfedges[:tri.trianglesLen]

	return nil
}

func (t *triangulator) hashKey(point Point) int {
	d := point.sub(t.center)
	return int(pseudoAngle(d.X, d.Y) * float64(len(t.hash)))
}

func (t *triangulator) hashEdge(e *node) {
	t.hash[t.hashKey(t.points[e.i])] = e
}

func (t *triangulator) addTriangle(i0, i1, i2, a, b, c int) int {
	i := t.trianglesLen
	t.triangles[i] = i0
	t.triangles[i+1] = i1
	t.triangles[i+2] = i2
	t.link(i, a)
	t.link(i+1, b)
	t.link(i+2, c)
	t.trianglesLen += 3
	return i
}

func (t *triangulator) link(a, b int) {
	t.halfedges[a] = b
	if b >= 0 {
		t.halfedges[b] = a
	}
}

func (t *triangulator) legalize(a int) int {
	// if the pair of triangles doesn't satisfy the Delaunay condition
	// (p1 is inside the circumcircle of [p0, pl, pr]), flip them,
	// then do the same check/flip recursively for the new pair of triangles
	//
	//           pl                    pl
	//          /||\                  /  \
	//       al/ || \bl            al/    \a
	//        /  ||  \              /      \
	//       /  a||b  \    flip    /___ar___\
	//     p0\   ||   /p1   =>   p0\---bl---/p1
	//        \  ||  /              \      /
	//       ar\ || /br             b\    /br
	//          \||/                  \  /
	//           pr                    pr

	b := t.halfedges[a]

	a0 := a - a%3
	b0 := b - b%3

	al := a0 + (a+1)%3
	ar := a0 + (a+2)%3
	bl := b0 + (b+2)%3

	if b < 0 {
		return ar
	}

	p0 := t.triangles[ar]
	pr := t.triangles[a]
	pl := t.triangles[al]
	p1 := t.triangles[bl]

	illegal := inCircle(t.points[p0], t.points[pr], t.points[pl], t.points[p1])

	if illegal {
		t.triangles[a] = p1
		t.triangles[b] = p0

		// edge swapped on the other side of the hull (rare)
		// fix the halfedge reference
		if t.halfedges[bl] == -1 {
			e := t.hull
			for {
				if e.t == bl {
					e.t = a
					break
				}
				e = e.next
				if e == t.hull {
					break
				}
			}
		}

		t.link(a, t.halfedges[bl])
		t.link(b, t.halfedges[ar])
		t.link(ar, bl)

		br := b0 + (b+1)%3

		t.legalize(a)
		return t.legalize(br)
	}

	return ar
}

func (t *triangulator) convexHull() []Point {
	var result []Point
	e := t.hull
	for e != nil {
		result = append(result, t.points[e.i])
		e = e.prev
		if e == t.hull {
			break
		}
	}
	return result
}

var eps = math.Nextafter(1, 2) - 1

var infinity = math.Inf(1)

func pseudoAngle(dx, dy float64) float64 {
	p := dx / (math.Abs(dx) + math.Abs(dy))
	if dy > 0 {
		p = (3 - p) / 4
	} else {
		p = (1 + p) / 4
	}
	return math.Max(0, math.Min(1-eps, p))
}

func area(a, b, c Point) float64 {
	return (b.Y-a.Y)*(c.X-b.X) - (b.X-a.X)*(c.Y-b.Y)
}

func inCircle(a, b, c, p Point) bool {
	dx := a.X - p.X
	dy := a.Y - p.Y
	ex := b.X - p.X
	ey := b.Y - p.Y
	fx := c.X - p.X
	fy := c.Y - p.Y

	ap := dx*dx + dy*dy
	bp := ex*ex + ey*ey
	cp := fx*fx + fy*fy

	return dx*(ey*cp-bp*fy)-dy*(ex*cp-bp*fx)+ap*(ex*fy-ey*fx) < 0
}

func circumradius(a, b, c Point) float64 {
	dx := b.X - a.X
	dy := b.Y - a.Y
	ex := c.X - a.X
	ey := c.Y - a.Y

	bl := dx*dx + dy*dy
	cl := ex*ex + ey*ey
	d := dx*ey - dy*ex

	x := (ey*bl - dy*cl) * 0.5 / d
	y := (dx*cl - ex*bl) * 0.5 / d

	r := x*x + y*y

	if bl == 0 || cl == 0 || d == 0 || r == 0 {
		return infinity
	}

	return r
}

func circumcenter(a, b, c Point) Point {
	dx := b.X - a.X
	dy := b.Y - a.Y
	ex := c.X - a.X
	ey := c.Y - a.Y

	bl := dx*dx + dy*dy
	cl := ex*ex + ey*ey
	d := dx*ey - dy*ex

	x := a.X + (ey*bl-dy*cl)*0.5/d
	y := a.Y + (dx*cl-ex*bl)*0.5/d

	return Point{x, y}
}

func polygonArea(points []Point) float64 {
	var result float64
	for i, p := range points {
		q := points[(i+1)%len(points)]
		result += (p.X - q.X) * (p.Y + q.Y)
	}
	return result / 2
}

func polygonPerimeter(points []Point) float64 {
	if len(points) == 0 {
		return 0
	}
	var result float64
	q := points[len(points)-1]
	for _, p := range points {
		result += p.distance(q)
		q = p
	}
	return result
}
