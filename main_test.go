package main

import (
	"fmt"
	"testing"

	"bitrise-steplib/steps-xcode-test-mac/mocks"

	"github.com/hashicorp/go-version"
	"github.com/stretchr/testify/require"
)

func TestGivenStep_WhenXCPrettyInstallSucceeds_ThenOutputToolIsXCPretty(t *testing.T) {
	logger := new(mocks.Logger)
	xcpretty := new(mocks.Installer)

	xcpretty.On("Install").Return(&version.Version{}, nil)

	step := NewStep(logger, xcpretty)
	outputTool := step.selectLogFormatter()
	require.Equal(t, xcprettyFormatter, outputTool)
}

func TestGivenStep_WhenXCPrettyInstallFails_ThenOutputToolIsXcodebuild(t *testing.T) {
	logger := new(mocks.Logger)
	xcpretty := new(mocks.Installer)

	xcpretty.On("Install").Return(nil, fmt.Errorf("install failed"))

	step := NewStep(logger, xcpretty)
	outputTool := step.selectLogFormatter()
	require.Equal(t, xcodebuildFormatter, outputTool)
}
