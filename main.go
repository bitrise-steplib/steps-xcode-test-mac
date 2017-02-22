package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/bitrise-io/go-utils/log"
	cmd "github.com/bitrise-io/steps-xcode-test/command"
	"github.com/bitrise-io/steps-xcode-test/xcodeutil"
	"github.com/bitrise-tools/go-xcode/xcodebuild"
	"github.com/bitrise-tools/go-xcode/xcpretty"
)

// ConfigsModel ...
type ConfigsModel struct {
	// Project parameters
	ProjectPath string
	Scheme      string

	// Test Run Configs
	OutputTool   string
	IsCleanBuild string

	GenerateCodeCoverageFiles string
}

func (configs ConfigsModel) print() {
	fmt.Println()
	log.Infof("Project Parameters:")
	log.Printf("- ProjectPath: %s", configs.ProjectPath)
	log.Printf("- Scheme: %s", configs.Scheme)

	fmt.Println()
	log.Infof("Test Run Configs:")
	log.Printf("- OutputTool: %s", configs.OutputTool)
	log.Printf("- IsCleanBuild: %s", configs.IsCleanBuild)

	log.Printf("- GenerateCodeCoverageFiles: %s", configs.GenerateCodeCoverageFiles)
}

func createConfigsModelFromEnvs() ConfigsModel {
	return ConfigsModel{
		// Project Parameters
		ProjectPath: os.Getenv("project_path"),
		Scheme:      os.Getenv("scheme"),

		// Test Run Configs
		OutputTool:   os.Getenv("output_tool"),
		IsCleanBuild: os.Getenv("is_clean_build"),

		GenerateCodeCoverageFiles: os.Getenv("generate_code_coverage_files"),
	}
}

func (configs ConfigsModel) validate() error {
	// required
	if err := validateRequiredInput(configs.ProjectPath, "project_path"); err != nil {
		return err
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

	if err := validateRequiredInputWithOptions(configs.GenerateCodeCoverageFiles, "generate_code_coverage_files", []string{"yes", "no"}); err != nil {
		return err
	}

	return nil
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

func exportTestResult(status string) {
	if err := cmd.ExportEnvironmentWithEnvman("BITRISE_XCODE_TEST_RESULT", status); err != nil {
		log.Warnf("Failed to export: BITRISE_XCODE_TEST_RESULT, error: %s", err)
	}
}

func runTest(output string, err error) {
	if err != nil {
		log.Errorf("Test failed, error: %s", err)
		exportTestResult("failed")
		os.Exit(1)
	}
}

//--------------------
// Main
//--------------------

func main() {
	configs := createConfigsModelFromEnvs()
	configs.print()
	if err := configs.validate(); err != nil {
		log.Errorf("Issue with input: %s", err)
		os.Exit(1)
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
		log.Errorf("Invalid project file (%s), extension should be (.xcodeproj/.xcworkspace)", configs.ProjectPath)
		exportTestResult("failed")
		os.Exit(1)
	}

	log.Printf("* action: %s", action)

	// Output tools versions
	xcodebuildVersion, err := xcodeutil.GetXcodeVersion()
	if err != nil {
		log.Errorf("Failed to get the version of xcodebuild! Error: %s", err)
		exportTestResult("failed")
		os.Exit(1)
	}

	log.Printf("* xcodebuild_version: %s (%s)", xcodebuildVersion.Version, xcodebuildVersion.BuildVersion)

	// xcpretty version
	if configs.OutputTool == "xcpretty" {
		xcprettyVersion, err := cmd.GetXcprettyVersion()
		if err != nil {
			log.Warnf("Failed to get the xcpretty version! Error: %s", err)
		} else {
			log.Printf("* xcpretty_version: %s", xcprettyVersion)
		}
	}

	fmt.Println()

	//setup buildActios
	buildAction := []string{}

	if cleanBuild {
		buildAction = append(buildAction, "clean")
	}

	//build before test
	buildAction = append(buildAction, "build")

	//setup CommandModel for test
	testCommandModel := xcodebuild.NewTestCommand(configs.ProjectPath, (action == "-workspace"))
	testCommandModel.SetScheme(configs.Scheme)
	testCommandModel.SetGenerateCodeCoverage(generateCodeCoverage)
	testCommandModel.SetCustomBuildAction(buildAction...)

	if configs.OutputTool == "xcpretty" {
		runTest(xcpretty.New(testCommandModel).Run())
	} else {
		runTest("", testCommandModel.Run())
	}
	exportTestResult("succeeded")
}
