package output

import (
	cp "github.com/otiai10/copy"

	"github.com/lukasjarosch/skipper"
)

const CopyOutputType = skipper.OutputType("copy")

type CopyConfig struct {
	SourcePath  string   `yaml:"sourcePath"`
	TargetPaths []string `yaml:"targetPaths"`
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

func (copy *Copy) Run() error {
	opts := cp.Options{}

	for _, target := range copy.config.TargetPaths {
		err := cp.Copy(copy.config.SourcePath, target, opts)
		if err != nil {
			return err
		}
	}

	return nil
}

func (copy *Copy) Copy(source, target string) error {
	// There are multiple cases which need to be handled.
	// 1. source is a FILE, target does NOT EXIST
	//   -

	return nil
}

func (copy *Copy) Configure() error {
	return nil
}

// ConfigPointer returns a pointer to [Config] which is going to be used by skipper
// to automatically load the config into said pointer, nice.
func (copy *Copy) ConfigPointer() interface{} {
	return &copy.config
}
