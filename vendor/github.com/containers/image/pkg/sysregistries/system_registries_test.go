package sysregistries

import (
	"github.com/containers/image/types"
	"github.com/stretchr/testify/assert"
	"testing"
)

var testConfig = []byte("")

func init() {
	readConf = func(_ *types.SystemContext) ([]byte, error) {
		return testConfig, nil
	}
}

func TestGetRegistriesWithBlankData(t *testing.T) {
	testConfig = []byte("")
	registriesConfig, _ := GetRegistries(nil)
	assert.Nil(t, registriesConfig)
}

func TestGetRegistriesWithData(t *testing.T) {
	answer := []string{"one.com"}
	testConfig = []byte(`[registries.search]
registries= ['one.com']
`)
	registriesConfig, err := GetRegistries(nil)
	assert.Nil(t, err)
	assert.Equal(t, registriesConfig, answer)
}

func TestGetRegistriesWithBadData(t *testing.T) {
	testConfig = []byte(`registries:
    - one.com
    ,`)
	_, err := GetRegistries(nil)
	assert.Error(t, err)
}

func TestGetInsecureRegistriesWithBlankData(t *testing.T) {
	answer := []string(nil)
	testConfig = []byte("")
	insecureRegistriesConfig, err := GetInsecureRegistries(nil)
	assert.Nil(t, err)
	assert.Equal(t, insecureRegistriesConfig, answer)
}

func TestGetInsecureRegistriesWithData(t *testing.T) {
	answer := []string{"two.com", "three.com"}
	testConfig = []byte(`[registries.search]
registries = ['one.com']
[registries.insecure]
registries = ['two.com', 'three.com']
`)
	insecureRegistriesConfig, err := GetInsecureRegistries(nil)
	if err != nil {
		t.Fail()
	}
	assert.Equal(t, insecureRegistriesConfig, answer)
}
