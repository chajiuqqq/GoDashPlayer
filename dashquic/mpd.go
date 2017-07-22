// Package mpd implements parsing and generating of MPEG-DASH Media Presentation Description (MPD) files.
package main

import (
	"encoding/xml"
)

// http://mpeg.chiariglione.org/standards/mpeg-dash
// https://www.brendanlong.com/the-structure-of-an-mpeg-dash-mpd.html
// http://standards.iso.org/ittf/PubliclyAvailableStandards/MPEG-DASH_schema_files/DASH-MPD.xsd

// Decode parses MPD XML.
func (m *MPD) Decode(b []byte) error {
	return xml.Unmarshal(b, m)
}

// MPD represents root XML element.
type MPD struct {
	XMLNS                     string `xml:"xmlns,attr"`
	MinBufferTime             string `xml:"minBufferTime,attr"` //PT1.500000S
	Type                      string `xml:"type,attr"`
	MediaPresentationDuration string `xml:"mediaPresentationDuration,attr"` //PT0H9M56.46S
	Profiles                  string `xml:"profiles,attr"`
	Period                    Period `xml:"Period,omitempty"`
}

// Period represents XSD's PeriodType.
type Period struct {
	Duration       string          `xml:"duration,attr"`
	AdaptationSets []AdaptationSet `xml:"AdaptationSet,omitempty"`
}

// AdaptationSet represents XSD's AdaptationSetType.
type AdaptationSet struct {
	MimeType         string           `xml:"mimeType,attr"`
	SegmentAlignment bool             `xml:"segmentAlignment,attr"`
	Group            int              `xml:"group,attr"`
	MaxWidth         int              `xml:"maxWidth,attr"`
	MaxHeight        int              `xml:"maxHeight,attr"`
	MaxFrameRate     float64          `xml:"maxFrameRate,attr"`
	Par              string           `xml:"par,attr"`
	Representations  []Representation `xml:"Representation,omitempty"`
}

// Representation represents XSD's RepresentationType.
type Representation struct {
	ID              string          `xml:"id,attr"`
	MimeType        string          `xml:"mimeType,attr"`
	Codecs          string          `xml:"codecs,attr"`
	Width           int             `xml:"width,attr"`
	Height          int             `xml:"height,attr"`
	FrameRate       float64         `xml:"frameRate,attr"`
	Sar             string          `xml:"sar,attr"`
	StartWithSAP    int             `xml:"startWithSAP,attr"`
	Bandwidth       float64         `xml:"bandwidth,attr"`
	SegmentTemplate SegmentTemplate `xml:"SegmentTemplate,omitempty"`
	SegmentSizes    []SegmentSize   `xml:"SegmentSize,omitempty"`
}

// SegmentTemplate represents XSD's SegmentTemplateType.
type SegmentTemplate struct {
	Timescale      float64 `xml:"timescale,attr"`
	Media          string  `xml:"media,attr"`
	Duration       float64 `xml:"duration,attr"`
	StartNumber    int     `xml:"startNumber,attr"`
	Initialization string  `xml:"initialization,attr"`
}

// SegmentSize represents XSD's SegmentSize
type SegmentSize struct {
	ID    string  `xml:"id,attr"`
	Size  float64 `xml:"size,attr"`
	Scale string  `xml:"scale,attr"`
}

func (*SegmentSize) getSegmentScaleAsFloat(scale string) float64 {
	return segmentScaleMap[scale]
}

var segmentScaleMap = map[string]float64{
	"bits":  1,
	"Kbits": 1024,
	"Mbits": 1024 * 1024,
	"bytes": 8,
	"KB":    1024 * 8,
	"MB":    1024 * 1024 * 8,
}
