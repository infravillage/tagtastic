package data

type CodeName struct {
	Name        string   `yaml:"name" json:"name"`
	Aliases     []string `yaml:"aliases" json:"aliases"`
	Description string   `yaml:"description" json:"description"`
}

type Theme struct {
	ID          string     `yaml:"id" json:"id"`
	Name        string     `yaml:"name" json:"name"`
	Description string     `yaml:"description" json:"description"`
	Category    string     `yaml:"category" json:"category"`
	Items       []CodeName `yaml:"items" json:"items"`
}
