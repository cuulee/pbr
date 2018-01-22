package surface

import (
	"github.com/hunterloftis/pbr/geom"
	"github.com/hunterloftis/pbr/surface/material"
)

// Triangle describes a triangle
// TODO: store per-vertex Normal data so .obj file curved surfaces can be read in and rendered smoothly / without edges
type Triangle struct {
	Points  [3]geom.Vector3
	Normals [3]geom.Direction
	Mat     *material.Material
	edge1   geom.Vector3
	edge2   geom.Vector3
	box     *Box
}

// NewTriangle creates a new triangle
func NewTriangle(a, b, c geom.Vector3, m *material.Material) *Triangle {
	edge1 := b.Minus(a)
	edge2 := c.Minus(a)
	n := edge1.Cross(edge2).Unit()
	t := &Triangle{
		Points:  [3]geom.Vector3{a, b, c},
		Normals: [3]geom.Direction{n, n, n},
		Mat:     m,
		edge1:   edge1,
		edge2:   edge2,
	}
	min := t.Points[0].Min(t.Points[1]).Min(t.Points[2])
	max := t.Points[0].Max(t.Points[1]).Max(t.Points[2])
	t.box = NewBox(min, max)
	return t
}

func (t *Triangle) Box() *Box {
	return t.box
}

func (t *Triangle) Material() *material.Material {
	return t.Mat
}

// Intersect determines whether or not, and where, a Ray intersects this Triangle
// https://en.wikipedia.org/wiki/M%C3%B6ller%E2%80%93Trumbore_intersection_algorithm
func (t *Triangle) Intersect(ray *geom.Ray3) Hit {
	if ok, _, _ := t.box.Check(ray); !ok {
		return Miss
	}
	h := ray.Dir.Cross(geom.Direction(t.edge2))
	a := t.edge1.Dot(geom.Vector3(h))
	if a > -bias && a < bias {
		return Miss
	}
	f := 1.0 / a
	s := ray.Origin.Minus(t.Points[0])
	u := f * s.Dot(geom.Vector3(h))
	if u < 0 || u > 1 {
		return Miss
	}
	q := s.Cross(t.edge1)
	v := f * geom.Vector3(ray.Dir).Dot(q)
	if v < 0 || (u+v) > 1 {
		return Miss
	}
	dist := f * t.edge2.Dot(q)
	if dist < bias {
		return Miss
	}
	return NewHit(t, dist)
}

func (t *Triangle) Center() geom.Vector3 {
	c := geom.Vector3{}
	for _, p := range t.Points {
		c = c.Plus(p)
	}
	return c.Scaled(1.0 / 3.0)
}

// At returns the material at a point on the Triangle
func (t *Triangle) At(v geom.Vector3) (norm geom.Direction, mat *material.Material) {
	return t.Normal(v), t.Mat
}

// SetNormals sets values for each vertex normal
func (t *Triangle) SetNormals(a, b, c *geom.Direction) {
	if a != nil {
		t.Normals[0] = *a
	}
	if b != nil {
		t.Normals[1] = *b
	}
	if c != nil {
		t.Normals[2] = *c
	}
}

// Normal computes the smoothed normal
func (t *Triangle) Normal(p geom.Vector3) geom.Direction {
	u, v, w := t.Bary(p)
	n0 := t.Normals[0].Scaled(u)
	n1 := t.Normals[1].Scaled(v)
	n2 := t.Normals[2].Scaled(w)
	return n0.Plus(n1).Plus(n2).Unit()
}

// Bary returns the Barycentric coords of Vector p on Triangle t
// TODO: using this in several places; integrate
// https://codeplea.com/triangular-interpolation
func (t *Triangle) Bary(p geom.Vector3) (u, v, w float64) {
	v0 := t.Points[1].Minus(t.Points[0])
	v1 := t.Points[2].Minus(t.Points[0])
	v2 := p.Minus(t.Points[0])
	d00 := v0.Dot(v0)
	d01 := v0.Dot(v1)
	d11 := v1.Dot(v1)
	d20 := v2.Dot(v0)
	d21 := v2.Dot(v1)
	d := d00*d11 - d01*d01
	v = (d11*d20 - d01*d21) / d
	w = (d00*d21 - d01*d20) / d
	u = 1 - v - w
	return
}
