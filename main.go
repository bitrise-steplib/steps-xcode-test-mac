package main

import (
	"fmt"
	"os"
	"strings"

	xcprettyi "bitrise-steplib/steps-xcode-test-mac/xcpretty"

	"github.com/bitrise-io/go-steputils/stepconf"
	"github.com/bitrise-io/go-utils/command"
	"github.com/bitrise-io/go-utils/log"
	"github.com/bitrise-io/go-utils/stringutil"
	logv2 "github.com/bitrise-io/go-utils/v2/log"
	"github.com/bitrise-io/go-xcode/utility"
	xcpretty2 "github.com/bitrise-io/go-xcode/v2/xcpretty"
	"github.com/bitrise-io/go-xcode/xcodebuild"
	"github.com/bitrise-io/go-xcode/xcpretty"
	"github.com/kballard/go-shellquote"
)

// configs ...
type configs struct {
	// Project parameters
	ProjectPath string `env:"project_path,dir"`
	Scheme      string `env:"scheme,required"`
	Destination string `env:"destination"`

	// Test Run Configs
	OutputTool   string `env:"output_tool,opt[xcpretty,xcodebuild]"`
	IsCleanBuild bool   `env:"is_clean_build,opt[yes,no]"`

	GenerateCodeCoverageFiles bool   `env:"generate_code_coverage_files,opt[yes,no]"`
	XcodebuildOptions         string `env:"xcodebuild_options"`
	DisableIndexWhileBuilding bool   `env:"disable_index_while_building,opt[yes,no]"`
}

// ExportEnvironmentWithEnvman ...
func ExportEnvironmentWithEnvman(keyStr, valueStr string) error {
	return command.New("envman", "add", "--key", keyStr).SetStdin(strings.NewReader(valueStr)).Run()
}

// GetXcprettyVersion ...
func GetXcprettyVersion() (string, error) {
	cmd := command.New("xcpretty", "-version")
	return cmd.RunAndReturnTrimmedCombinedOutput()
}

func exportTestResult(status string) {
	if err := ExportEnvironmentWithEnvman("BITRISE_XCODE_TEST_RESULT", status); err != nil {
		log.Warnf("Failed to export: BITRISE_XCODE_TEST_RESULT, error: %s", err)
	}
}

func failf(format string, v ...interface{}) {
	exportTestResult("failed")
	log.Errorf(format, v...)
	os.Exit(1)
}

//--------------------
// Main
//--------------------

type Step struct {
	logger   logv2.Logger
	xcpretty xcprettyi.Installer
}

func NewStep(logger logv2.Logger, xcpretty xcprettyi.Installer) Step {
	return Step{logger: logger, xcpretty: xcpretty}
}

func (s Step) ensureXCPretty() string {
	outputTool := "xcpretty"

	ver, err := s.xcpretty.Install()
	if err != nil {
		log.Warnf("Failed to ensure xcpretty log formatter: %s", err)
		log.Printf("Switching to xcodebuild for output tool")
		outputTool = "xcodebuild"
	} else {
		log.Printf("- xcpretty version: %s", ver.String())
		fmt.Println()
	}

	return outputTool
}

func (s Step) run() {
	var cfgs configs
	if err := stepconf.Parse(&cfgs); err != nil {
		failf("Issue with input: %s", err)
	}

	stepconf.Print(cfgs)
	fmt.Println()

	// Project-or-Workspace flag
	action := ""
	if strings.HasSuffix(cfgs.ProjectPath, ".xcodeproj") {
		action = "-project"
	} else if strings.HasSuffix(cfgs.ProjectPath, ".xcworkspace") {
		action = "-workspace"
	} else {
		failf("Invalid project file (%s), extension should be (.xcodeproj/.xcworkspace)", cfgs.ProjectPath)
	}

	log.Printf("* action: %s", action)

	// Output tools versions
	xcodebuildVersion, err := utility.GetXcodeVersion()
	if err != nil {
		failf("Failed to get the version of xcodebuild! Error: %s", err)
	}

	log.Printf("* xcodebuild_version: %s (%s)", xcodebuildVersion.Version, xcodebuildVersion.BuildVersion)

	// xcpretty
	cfgs.OutputTool = s.ensureXCPretty()

	fmt.Println()

	// setup buildActions
	buildAction := []string{}

	if cfgs.IsCleanBuild {
		buildAction = append(buildAction, "clean")
	}

	// build before test
	buildAction = append(buildAction, "build")

	// setup CommandModel for test
	testCommandModel := xcodebuild.NewTestCommand(cfgs.ProjectPath, (action == "-workspace"))
	testCommandModel.SetScheme(cfgs.Scheme)
	testCommandModel.SetGenerateCodeCoverage(cfgs.GenerateCodeCoverageFiles)
	testCommandModel.SetCustomBuildAction(buildAction...)

	testCommandModel.SetDisableIndexWhileBuilding(cfgs.DisableIndexWhileBuilding)

	if cfgs.Destination != "" {
		testCommandModel.SetDestination(cfgs.Destination)
	}

	if cfgs.XcodebuildOptions != "" {
		options, err := shellquote.Split(cfgs.XcodebuildOptions)
		if err != nil {
			failf("Failed to shell split XcodebuildOptions (%s), error: %s", cfgs.XcodebuildOptions)
		}
		testCommandModel.SetCustomOptions(options)
	}

	if cfgs.OutputTool == "xcpretty" {
		xcprettyCmd := xcpretty.New(testCommandModel)

		log.Infof("$ %s\n", xcprettyCmd.PrintableCmd())

		if rawXcodebuildOutput, err := xcprettyCmd.Run(); err != nil {
			log.Errorf("\nLast lines of the Xcode's build log:")
			fmt.Println(stringutil.LastNLines(rawXcodebuildOutput, 10))
			failf("Test failed, error: %s", err)
		}
	} else {
		log.Infof("$ %s\n", testCommandModel.PrintableCmd())

		if err := testCommandModel.Run(); err != nil {
			failf("Test failed, error: %s", err)
		}
	}
	exportTestResult("succeeded")
}

func main() {
	logger := logv2.NewLogger()
	xcpretty := xcprettyi.NewInstaller(logger, xcpretty2.NewXcpretty(logger))

	step := NewStep(logger, xcpretty)
	step.run()
}
