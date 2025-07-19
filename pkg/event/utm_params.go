package event

// UtmParams holds Urchin Tracking Module (UTM) URL parameters.
// See https://en.wikipedia.org/wiki/UTM_parameters.
type UtmParams struct {
	Source   string `json:"source"`
	Medium   string `json:"medium"`
	Campaign string `json:"campaign"`
	Term     string `json:"term"`
	Content  string `json:"content"`
}
