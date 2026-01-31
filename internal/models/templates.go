package models

// CardTemplate defines the visual properties for the digital pass
type CardTemplate struct {
	Name            string `json:"name"`
	BackgroundColor string `json:"background_color"`
	ForegroundColor string `json:"foreground_color"`
	LabelColor      string `json:"label_color"`
	Description     string `json:"description"`
}

// GetDefaultTemplates returns the 3 available designs for the brunch card
func GetDefaultTemplates() map[string]CardTemplate {
	return map[string]CardTemplate{
		"minimalist": {
			Name:            "Minimalist White",
			BackgroundColor: "#FFFFFF",
			ForegroundColor: "#000000",
			LabelColor:      "#666666",
			Description:     "A clean and modern look for your brunch rewards.",
		},
		"tropical": {
			Name:            "Tropical Vibes",
			BackgroundColor: "#2E7D32", // Dark Green
			ForegroundColor: "#FFFFFF",
			LabelColor:      "#A5D6A7",
			Description:     "Inspired by summer brunches and fresh fruits.",
		},
		"retro": {
			Name:            "Retro Diner",
			BackgroundColor: "#D84315", // Deep Orange
			ForegroundColor: "#FFFFFF",
			LabelColor:      "#FFCCBC",
			Description:     "Classic aesthetic for coffee and pancake lovers.",
		},
	}
}
