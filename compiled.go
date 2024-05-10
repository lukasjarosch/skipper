package skipper

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"fmt"
	"os"
	"time"

	"github.com/google/uuid"
)

var ErrCompiledInventoryIntegityCheckFailed = fmt.Errorf("CompiledInventory failed to pass the integrity check and cannot be loaded")

// CompiledInventory is the result of calling [Inventory.Compile].
// It is essentially a read-only copy of the source Inventory
// but with all expressions executed and other dependencies (e.g. target handling, secrets, ...) taken care of.
// If one then makes further modifications to the source inventory, a new CompiledInventory artifact must be created.
// The CompiledInventory can then be passed on to make good use of the heap of data :)
// TODO: should this have metadata like buildTimestamp and who build it?
// TODO: This should have capabilities() func like (can-write, can-reveal) to tell the consumer what it can do with it (and to allow more flexibility)
type CompiledInventory struct {
	// BuildID is an UUIDv4
	BuildID string
	// BuildTimestamp is the timestamp of creation of the instance, in UTF.
	BuildTimestamp time.Time
	// capabilities determine what can be done with this inventory when loaded by a consumer
	capability InventoryCapability
	scopes     map[Scope]*Registry
}

func (c CompiledInventory) HasCapability(cap InventoryCapability) bool {
	return c.capability&cap != 0
}

// NewCompiledInventory returns a new CompiledInventory
// with read capability, buildId and buildTimestamp.
func NewCompiledInventory() CompiledInventory {
	return CompiledInventory{
		BuildID:        uuid.New().String(),
		BuildTimestamp: time.Now().UTC(),
		capability:     InventoryRead,
	}
}

// InventoryCapability is a binary value.
// Capabilities define what a consumer can do with an Inventory.
type InventoryCapability uint8

const (
	// Read (default capability) allows a consumer to read all values within the inventory.
	InventoryRead InventoryCapability = 1 << iota
	// Write allows the user to modify the inventory
	InventoryWrite
	// Reveal allows, given the config is provided, that the consumer can reveal secrets
	InventoryReveal
)

func (c *InventoryCapability) AddCapability(cap InventoryCapability) {
	*c |= cap
}

func (c *InventoryCapability) RemoveCapability(cap InventoryCapability) {
	*c &= ^cap
}

func init() {
	gob.Register(CompiledInventory{})
}

// WriteCompiledInventoryFile writes the provided CompiledInventory to the specified file path.
// It encodes the inventory using gob encoding and appends a SHA256 hash of the encoded data to ensure file integrity.
// If the file already exists, it will be overwritten.
// Parameters:
//   - inventory: The CompiledInventory to write to the file.
//   - filePath: The file path where the inventory will be written.
//
// Returns:
//   - An error if any occurred during file writing or encoding.
func WriteCompiledInventoryFile(inventory CompiledInventory, filePath string) error {
	var buf bytes.Buffer
	encoder := gob.NewEncoder(&buf)

	err := encoder.Encode(inventory)
	if err != nil {
		return err
	}

	// calculate SHA256 hash of the encoded data and append it
	hash := sha256.Sum256(buf.Bytes())
	data := append(hash[:], buf.Bytes()...)

	err = os.WriteFile(filePath, data, 0644)
	if err != nil {
		return err
	}

	return nil
}

// LoadCompiledInventoryFile loads a CompiledInventory from the specified file path.
// It verifies the integrity of the file by comparing the SHA256 hash of the data with the hash appended to the file.
// Parameters:
//   - filePath: The file path from which to load the inventory.
//
// Returns:
//   - The Loaded CompiledInventory.
//   - An error if any occurred during file reading, decoding, or if the integrity check failed.
func LoadCompiledInventoryFile(filePath string) (CompiledInventory, error) {
	var zero CompiledInventory

	data, err := os.ReadFile(filePath)
	if err != nil {
		return zero, err
	}

	// extract hash from the beginning of the data
	hash := data[:sha256.Size]
	fileData := data[sha256.Size:]

	// calculate hash of the actual data and compare
	actualHash := sha256.Sum256(fileData)
	if !bytes.Equal(hash, actualHash[:]) {
		return zero, ErrCompiledInventoryIntegityCheckFailed
	}

	// decode the remaining fileData
	buf := bytes.NewBuffer(fileData)
	decoder := gob.NewDecoder(buf)

	inventory := CompiledInventory{}
	err = decoder.Decode(&inventory)
	if err != nil {
		return zero, err
	}

	return inventory, nil
}
