package handlers

import (
	"fmt"

	"github.com/uber/h3-go/v3"
)

type HotspotsIndex struct {
	Name    string `json:"name"`
	Index   string `json:"index"`
	Address string `json:"id"`
}

func LatLongToH3(lat float64, long float64, resolution int) string {
	geo := h3.GeoCoord{
		Latitude:  lat,
		Longitude: long,
	}

	indexValue := fmt.Sprintf("%x", h3.FromGeo(geo, resolution))

	return indexValue
}

func H3ToLatLong(index string) (float64, float64) {

	if index != "" {
		geoString := h3.FromString(index)
		geo := h3.ToGeo(geoString)
		return geo.Latitude, geo.Longitude
	}

	return 0, 0
}
