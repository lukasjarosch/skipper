package main_test

import (
	"io/ioutil"
	"path"
	"testing"

	"github.com/onsi/gomega"
)

func TestInventory(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	goldenFile := "azure_keyvault_inventory.golden.json"
	inventortyFile := path.Join("compiled", "azure_keyvault", "inventory.json")

	goldenData, err := ioutil.ReadFile(path.Join("testdata", goldenFile))
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	inventoryData, err := ioutil.ReadFile(inventortyFile)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	g.Expect(goldenData).To(gomega.MatchJSON(inventoryData))
}
