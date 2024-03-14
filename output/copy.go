package output

import (
	"fmt"
	"os"

	"github.com/davecgh/go-spew/spew"

	"github.com/lukasjarosch/skipper"
)

const CopyOutputType = skipper.OutputType("copy")

type CopyConfig struct {
	SourcePaths []string `yaml:"sourcePaths"`
	TargetPath  string   `yaml:"targetPath"`
	Overwrite   bool     `yaml:"overwrite"`
}

type Copy struct {
	config CopyConfig
}

func NewCopyOutput() skipper.OutputConstructorFunc {
	return func() skipper.Output {
		return &Copy{}
	}
}

// Configure checks that all configured source-paths exist and are readable.
func (copy *Copy) Configure() error {
	// source-paths must be readable
	for _, sourcePath := range copy.config.SourcePaths {
		_, err := os.Open(sourcePath)
		if err != nil {
			return fmt.Errorf("cannot open source-path: %w", err)
		}
	}

	return nil
}

func (copy *Copy) Run() error {
	spew.Println("RUN")
	return nil
}

// ConfigPointer returns a pointer to [Config] which is going to be used by skipper
// to automatically load the config into said pointer, nice.
func (copy *Copy) ConfigPointer() interface{} {
	return &copy.config
}
