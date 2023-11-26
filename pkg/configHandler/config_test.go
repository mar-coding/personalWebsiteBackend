package configHandler

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

type ExtraData struct {
	Email   string `yaml:"email" json:"email"`
	Counter int    `yaml:"counter" json:"counter"`
}

type ExtraData1 struct {
	Foo string `yaml:"foo" json:"foo"`
	Bar int    `yaml:"bar" json:"bar"`
}

func Test_UnmarshalYAML(t *testing.T) {
	c, err := New[ExtraData]("./testdata/config.yml")
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, "0.0.0.0", c.Address)
	assert.Equal(t, "test@test.com", c.ExtraData.Email)
}

func Test_UnmarshalJSON(t *testing.T) {
	c, err := New[ExtraData1]("./testdata/config.json")
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, "bar", c.ExtraData.Foo)
	assert.Equal(t, 1234, c.ExtraData.Bar)
}
