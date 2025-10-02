package geonames

type status struct {
	Value   int    `json:"value"`
	Message string `json:"message"`
}

type response struct {
	Time        string  `json:"time"`
	TimeZoneID  string  `json:"timezoneId"`
	GMTOffset   float64 `json:"gmtOffset"`
	DSTOffset   float64 `json:"dstOffset"`
	RawOffset   float64 `json:"rawOffset"`
	Country     string  `json:"countryName"`
	CountryCode string  `json:"countryCode"`
	Sunrise     string  `json:"sunrise"`
	Sunset      string  `json:"sunset"`
	Status      *status `json:"status,omitempty"`
}
