package xcpretty

import (
	"fmt"

	"github.com/bitrise-io/go-utils/v2/log"
	"github.com/bitrise-io/go-xcode/v2/xcpretty"
	"github.com/hashicorp/go-version"
)

// Installer ...
type Installer interface {
	Install() (*version.Version, error)
}

type installer struct {
	logger   log.Logger
	xcpretty xcpretty.Xcpretty
}

// NewInstaller ...
func NewInstaller(logger log.Logger, xcpretty xcpretty.Xcpretty) Installer {
	return &installer{
		logger:   logger,
		xcpretty: xcpretty,
	}
}

// Install installs and gets xcpretty version
func (i installer) Install() (*version.Version, error) {
	i.logger.Println()
	i.logger.Infof("Checking if output tool (xcpretty) is installed")

	installed, err := i.xcpretty.IsInstalled()
	if err != nil {
		return nil, err
	} else if !installed {
		i.logger.Warnf(`xcpretty is not installed`)
		i.logger.Println()
		i.logger.Printf("Installing xcpretty")

		cmdModelSlice, err := i.xcpretty.Install()
		if err != nil {
			return nil, fmt.Errorf("failed to create xcpretty install commands: %w", err)
		}

		for _, cmd := range cmdModelSlice {
			if err := cmd.Run(); err != nil {
				return nil, fmt.Errorf("failed to run xcpretty install command (%s): %w", cmd.PrintableCommandArgs(), err)
			}
		}
	}

	xcprettyVersion, err := i.xcpretty.Version()
	if err != nil {
		return nil, fmt.Errorf("failed to get xcpretty version: %w", err)
	}
	return xcprettyVersion, nil
}
