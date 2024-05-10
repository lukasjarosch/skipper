package skipper_test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/lukasjarosch/skipper/v1"
)

func TestCompiledInventory_WriteAndLoadCompiledInventoryFile(t *testing.T) {
	// Create a test inventory
	testInventory := skipper.CompiledInventory{
		BuildID: "test-build-id",
	}

	// Define a temporary file path for testing
	testFilePath := "test_inventory.bin"

	// Write the inventory to file
	err := skipper.WriteCompiledInventoryFile(testInventory, testFilePath)
	assert.NoError(t, err)
	defer os.Remove(testFilePath)

	// Load the inventory from file
	loadedInventory, err := skipper.LoadCompiledInventoryFile(testFilePath)
	assert.NoError(t, err)

	// Check if loaded inventory matches the original
	assert.Equal(t, testInventory.BuildID, loadedInventory.BuildID)
}

func TestCompiledInventory_FileIntegrityCheck(t *testing.T) {
	// Create a test inventory
	testInventory := skipper.CompiledInventory{
		BuildID: "test-build-id",
	}

	// Define a temporary file path for testing
	testFilePath := "test_inventory_with_integrity.bin"

	// Write the inventory to file
	err := skipper.WriteCompiledInventoryFile(testInventory, testFilePath)
	assert.NoError(t, err)
	defer os.Remove(testFilePath)

	// Manually modify the file to simulate tampering
	file, err := os.OpenFile(testFilePath, os.O_RDWR, 0644)
	assert.NoError(t, err)
	defer file.Close()

	// Modify the first byte of the file to change the hash
	_, err = file.WriteAt([]byte{0}, 0)
	assert.NoError(t, err)

	// Attempt to load the inventory from the tampered file
	_, err = skipper.LoadCompiledInventoryFile(testFilePath)
	assert.Error(t, err)
}
