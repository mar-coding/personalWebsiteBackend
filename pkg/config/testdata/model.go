package config

type Config struct {
	Address   string    `yaml:"address" json:"address"`
	ExtraData ExtraData `yaml:"extra_data" json:"extra_data"`
}
type ExtraData struct {
	Email string `yaml:"email" json:"email"`
}
