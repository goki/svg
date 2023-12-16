// Copyright (c) 2018, The GoKi Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package svg

import (
	"goki.dev/mat32/v2"
)

// Polyline is a SVG multi-line shape
type Polyline struct {
	NodeBase

	// the coordinates to draw -- does a moveto on the first, then lineto for all the rest
	Points []mat32.Vec2 `xml:"points"`
}

func (g *Polyline) SVGName() string { return "polyline" }

func (g *Polyline) CopyFieldsFrom(frm any) {
	fr := frm.(*Polyline)
	g.NodeBase.CopyFieldsFrom(&fr.NodeBase)
	g.Points = make([]mat32.Vec2, len(fr.Points))
	copy(g.Points, fr.Points)
}

func (g *Polyline) SetPos(pos mat32.Vec2) {
	// todo: set offset relative to bbox
}

func (g *Polyline) SetSize(sz mat32.Vec2) {
	// todo: scale bbox
}

func (g *Polyline) LocalBBox() mat32.Box2 {
	bb := mat32.EmptyB2()
	for _, pt := range g.Points {
		bb.ExpandByPoint(pt)
	}
	hlw := 0.5 * g.LocalLineWidth()
	bb.Min.SetSubScalar(hlw)
	bb.Max.SetAddScalar(hlw)
	return bb
}

func (g *Polyline) Render(sv *SVG) {
	sz := len(g.Points)
	if sz < 2 {
		return
	}
	vis, pc := g.PushTransform(sv)
	if !vis {
		return
	}
	pc.Lock()
	pc.DrawPolyline(g.Points)
	pc.FillStrokeClear()
	pc.Unlock()
	g.BBoxes(sv)

	if mrk := sv.MarkerByName(g, "marker-start"); mrk != nil {
		pt := g.Points[0]
		ptn := g.Points[1]
		ang := mat32.Atan2(ptn.Y-pt.Y, ptn.X-pt.X)
		mrk.RenderMarker(sv, pt, ang, g.Paint.StrokeStyle.Width.Dots)
	}
	if mrk := sv.MarkerByName(g, "marker-end"); mrk != nil {
		pt := g.Points[sz-1]
		ptp := g.Points[sz-2]
		ang := mat32.Atan2(pt.Y-ptp.Y, pt.X-ptp.X)
		mrk.RenderMarker(sv, pt, ang, g.Paint.StrokeStyle.Width.Dots)
	}
	if mrk := sv.MarkerByName(g, "marker-mid"); mrk != nil {
		for i := 1; i < sz-1; i++ {
			pt := g.Points[i]
			ptp := g.Points[i-1]
			ptn := g.Points[i+1]
			ang := 0.5 * (mat32.Atan2(pt.Y-ptp.Y, pt.X-ptp.X) + mat32.Atan2(ptn.Y-pt.Y, ptn.X-pt.X))
			mrk.RenderMarker(sv, pt, ang, g.Paint.StrokeStyle.Width.Dots)
		}
	}

	g.RenderChildren(sv)
	pc.PopTransformLock()
}

// ApplyTransform applies the given 2D transform to the geometry of this node
// each node must define this for itself
func (g *Polyline) ApplyTransform(sv *SVG, xf mat32.Mat2) {
	rot := xf.ExtractRot()
	if rot != 0 || !g.Paint.Transform.IsIdentity() {
		g.Paint.Transform = g.Paint.Transform.Mul(xf)
		g.SetProp("transform", g.Paint.Transform.String())
	} else {
		for i, p := range g.Points {
			p = xf.MulVec2AsPt(p)
			g.Points[i] = p
		}
		g.GradientApplyTransform(sv, xf)
	}
}

// ApplyDeltaTransform applies the given 2D delta transforms to the geometry of this node
// relative to given point.  Trans translation and point are in top-level coordinates,
// so must be transformed into local coords first.
// Point is upper left corner of selection box that anchors the translation and scaling,
// and for rotation it is the center point around which to rotate
func (g *Polyline) ApplyDeltaTransform(sv *SVG, trans mat32.Vec2, scale mat32.Vec2, rot float32, pt mat32.Vec2) {
	crot := g.Paint.Transform.ExtractRot()
	if rot != 0 || crot != 0 {
		xf, lpt := g.DeltaTransform(trans, scale, rot, pt, false) // exclude self
		mat := g.Paint.Transform.MulCtr(xf, lpt)
		g.Paint.Transform = mat
		g.SetProp("transform", g.Paint.Transform.String())
	} else {
		xf, lpt := g.DeltaTransform(trans, scale, rot, pt, true) // include self
		for i, p := range g.Points {
			p = xf.MulVec2AsPtCtr(p, lpt)
			g.Points[i] = p
		}
		g.GradientApplyTransformPt(sv, xf, lpt)
	}
}

// WriteGeom writes the geometry of the node to a slice of floating point numbers
// the length and ordering of which is specific to each node type.
// Slice must be passed and will be resized if not the correct length.
func (g *Polyline) WriteGeom(sv *SVG, dat *[]float32) {
	sz := len(g.Points) * 2
	SetFloat32SliceLen(dat, sz+6)
	for i, p := range g.Points {
		(*dat)[i*2] = p.X
		(*dat)[i*2+1] = p.Y
	}
	g.WriteTransform(*dat, sz)
	g.GradientWritePts(sv, dat)
}

// ReadGeom reads the geometry of the node from a slice of floating point numbers
// the length and ordering of which is specific to each node type.
func (g *Polyline) ReadGeom(sv *SVG, dat []float32) {
	sz := len(g.Points) * 2
	for i, p := range g.Points {
		p.X = dat[i*2]
		p.Y = dat[i*2+1]
		g.Points[i] = p
	}
	g.ReadTransform(dat, sz)
	g.GradientReadPts(sv, dat)
}
