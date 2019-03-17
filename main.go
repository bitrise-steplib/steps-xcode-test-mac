package main

import (
	"fmt"
	"os"
	"strings"

	"errors"

	"github.com/bitrise-io/go-utils/command"
	"github.com/bitrise-io/go-utils/log"
	"github.com/bitrise-io/go-utils/pathutil"
	"github.com/bitrise-io/go-utils/stringutil"
	"github.com/bitrise-tools/go-xcode/utility"
	"github.com/bitrise-tools/go-xcode/xcodebuild"
	"github.com/bitrise-tools/go-xcode/xcpretty"
	"github.com/kballard/go-shellquote"
)

// ConfigsModel ...
type ConfigsModel struct {
	// Project parameters
	ProjectPath string
	Scheme      string
	Destination string

	// Test Run Configs
	OutputTool   string
	IsCleanBuild string

	GenerateCodeCoverageFiles string
	XcodebuildOptions         string
}

func (configs ConfigsModel) print() {
	fmt.Println()
	log.Infof("Project Parameters:")
	log.Printf("- ProjectPath: %s", configs.ProjectPath)
	log.Printf("- Scheme: %s", configs.Scheme)
	log.Printf("- Destination: %s", configs.Destination)

	fmt.Println()
	log.Infof("Test Run Configs:")
	log.Printf("- OutputTool: %s", configs.OutputTool)
	log.Printf("- IsCleanBuild: %s", configs.IsCleanBuild)

	log.Printf("- GenerateCodeCoverageFiles: %s", configs.GenerateCodeCoverageFiles)
	log.Printf("- XcodebuildOptions: %s", configs.XcodebuildOptions)
}

func createConfigsModelFromEnvs() ConfigsModel {
	return ConfigsModel{
		// Project Parameters
		ProjectPath: os.Getenv("project_path"),
		Scheme:      os.Getenv("scheme"),
		Destination: os.Getenv("destination"),

		// Test Run Configs
		OutputTool:   os.Getenv("output_tool"),
		IsCleanBuild: os.Getenv("is_clean_build"),

		GenerateCodeCoverageFiles: os.Getenv("generate_code_coverage_files"),
		XcodebuildOptions: os.Getenv("xcodebuild_options"),
	}
}

func (configs ConfigsModel) validate() error {
	// required
	if err := validateRequiredInput(configs.ProjectPath, "project_path"); err != nil {
		return err
	}
	if exists, err := pathutil.IsDirExists(configs.ProjectPath); err != nil {
		return err
	} else if !exists {
		return errors.New("ProjectPath directory does not exists: %s")
	}

	if err := validateRequiredInput(configs.Scheme, "scheme"); err != nil {
		return err
	}

	if err := validateRequiredInputWithOptions(configs.OutputTool, "output_tool", []string{"xcpretty", "xcodebuild"}); err != nil {
		return err
	}
	if err := validateRequiredInputWithOptions(configs.IsCleanBuild, "is_clean_build", []string{"yes", "no"}); err != nil {
		return err
	}

	return validateRequiredInputWithOptions(configs.GenerateCodeCoverageFiles, "generate_code_coverage_files", []string{"yes", "no"})
}

//--------------------
// Functions
//--------------------

func validateRequiredInput(value, key string) error {
	if value == "" {
		return fmt.Errorf("Missing required input: %s", key)
	}
	return nil
}

func validateRequiredInputWithOptions(value, key string, options []string) error {
	if err := validateRequiredInput(key, value); err != nil {
		return err
	}

	found := false
	for _, option := range options {
		if option == value {
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("Invalid input: (%s) value: (%s), valid options: %s", key, value, strings.Join(options, ", "))
	}

	return nil
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

func main() {
	configs := createConfigsModelFromEnvs()
	configs.print()
	if err := configs.validate(); err != nil {
		failf("Issue with input: %s", err)
	}

	fmt.Println()
	log.Infof("Other Configs:")

	cleanBuild := (configs.IsCleanBuild == "yes")
	generateCodeCoverage := (configs.GenerateCodeCoverageFiles == "yes")

	// Project-or-Workspace flag
	action := ""
	if strings.HasSuffix(configs.ProjectPath, ".xcodeproj") {
		action = "-project"
	} else if strings.HasSuffix(configs.ProjectPath, ".xcworkspace") {
		action = "-workspace"
	} else {
		failf("Invalid project file (%s), extension should be (.xcodeproj/.xcworkspace)", configs.ProjectPath)
	}

	log.Printf("* action: %s", action)

	// Output tools versions
	xcodebuildVersion, err := utility.GetXcodeVersion()
	if err != nil {
		failf("Failed to get the version of xcodebuild! Error: %s", err)
	}

	log.Printf("* xcodebuild_version: %s (%s)", xcodebuildVersion.Version, xcodebuildVersion.BuildVersion)

	// xcpretty version
	if configs.OutputTool == "xcpretty" {
		xcprettyVersion, err := GetXcprettyVersion()
		if err != nil {
			failf("Failed to get the xcpretty version! Error: %s", err)
		} else {
			log.Printf("* xcpretty_version: %s", xcprettyVersion)
		}
	}

	fmt.Println()

	// setup buildActions
	buildAction := []string{}

	if cleanBuild {
		buildAction = append(buildAction, "clean")
	}

	// build before test
	buildAction = append(buildAction, "build")

	// setup CommandModel for test
	testCommandModel := xcodebuild.NewTestCommand(configs.ProjectPath, (action == "-workspace"))
	testCommandModel.SetScheme(configs.Scheme)
	testCommandModel.SetGenerateCodeCoverage(generateCodeCoverage)
	testCommandModel.SetCustomBuildAction(buildAction...)

	if configs.Destination != "" {
		testCommandModel.SetDestination(configs.Destination)
	}

	if configs.XcodebuildOptions != "" {
		options, err := shellquote.Split(configs.XcodebuildOptions)
		if err != nil {
			failf("Failed to shell split XcodebuildOptions (%s), error: %s", configs.XcodebuildOptions)
		}
		testCommandModel.SetCustomOptions(options)
	}

	if configs.OutputTool == "xcpretty" {
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
