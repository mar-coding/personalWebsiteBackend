package configHandler

import (
	"fmt"
	"log"
)

func ExampleNew() {
	type ExtraData struct {
		Email   string `yaml:"email" json:"email"`
		Counter int    `yaml:"counter" json:"counter"`
	}

	cfg, err := New[ExtraData]("./config.yml")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(cfg)
}
