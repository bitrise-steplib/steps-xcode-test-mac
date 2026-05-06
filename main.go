package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	xcprettyinstaller "bitrise-steplib/steps-xcode-test-mac/xcpretty"

	"github.com/bitrise-io/bitrise-build-cache-cli/v2/pkg/reactnative/wrap"
	"github.com/bitrise-io/go-steputils/stepconf"
	"github.com/bitrise-io/go-utils/command"
	"github.com/bitrise-io/go-utils/log"
	"github.com/bitrise-io/go-utils/stringutil"
	logV2 "github.com/bitrise-io/go-utils/v2/log"
	"github.com/bitrise-io/go-xcode/utility"
	xcprettyV2 "github.com/bitrise-io/go-xcode/v2/xcpretty"
	"github.com/bitrise-io/go-xcode/xcodebuild"
	"github.com/bitrise-io/go-xcode/xcpretty"
	"github.com/kballard/go-shellquote"
)

// runXcodebuildWithRNWrap routes an xcodebuild invocation through
// `bitrise-build-cache react-native run -- ...` when React Native build cache
// is active on the host. Returns (combinedOutput, didRun, err). When didRun
// is false, callers should fall through to the existing v1 path.
//
// When useXcpretty is true and the wrap engages, xcpretty is preserved by
// piping the wrapped command's stdout into a separately-spawned xcpretty
// process, while teeing the raw output into the returned string for the
// existing log-parsing path.
func runXcodebuildWithRNWrap(originalCmd *exec.Cmd, useXcpretty bool) (string, bool, error) {
	det := wrap.Detect(context.Background(), wrap.DetectParams{Logger: logV2.NewLogger()})
	if !det.ReactNativeEnabled {
		return "", false, nil
	}

	args := originalCmd.Args
	if len(args) == 0 {
		return "", false, nil
	}
	name, wrapped := wrap.Wrap(det, args[0], args[1:])
	display := append([]string{name}, wrapped...)
	log.Infof("$ %s\n", strings.Join(display, " "))

	var combined bytes.Buffer
	xcCmd := exec.Command(name, wrapped...) //nolint:gosec
	xcCmd.Dir = originalCmd.Dir

	if !useXcpretty {
		xcCmd.Stdout = io.MultiWriter(os.Stdout, &combined)
		xcCmd.Stderr = io.MultiWriter(os.Stderr, &combined)

		return combined.String(), true, xcCmd.Run()
	}

	// xcpretty pipeline: wrapped xcodebuild stdout → xcpretty stdin, while we
	// tee the raw output into `combined` for downstream log handling.
	xcprettyCmd := exec.Command("xcpretty") //nolint:gosec
	pr, pw := io.Pipe()
	xcCmd.Stdout = io.MultiWriter(pw, &combined)
	xcCmd.Stderr = io.MultiWriter(os.Stderr, &combined)
	xcprettyCmd.Stdin = pr
	xcprettyCmd.Stdout = os.Stdout
	xcprettyCmd.Stderr = os.Stderr

	if err := xcprettyCmd.Start(); err != nil {
		_ = pw.Close()

		return "", true, fmt.Errorf("start xcpretty: %w", err)
	}

	runErr := xcCmd.Run()
	_ = pw.Close()
	waitErr := xcprettyCmd.Wait()

	if runErr != nil {
		return combined.String(), true, runErr
	}
	if waitErr != nil {
		return combined.String(), true, waitErr
	}

	return combined.String(), true, nil
}

const (
	xcprettyFormatter   = "xcpretty"
	xcodebuildFormatter = "xcodebuild"
)

// configs ...
type configs struct {
	// Project parameters
	ProjectPath string `env:"project_path"`
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

// Step ...
type Step struct {
	logger   logV2.Logger
	xcpretty xcprettyinstaller.Installer
}

// NewStep ...
func NewStep(logger logV2.Logger, xcpretty xcprettyinstaller.Installer) Step {
	return Step{logger: logger, xcpretty: xcpretty}
}

func (s Step) selectLogFormatter(outputTool string) string {
	if outputTool == xcprettyFormatter {
		ver, err := s.xcpretty.Install()
		if err != nil {
			log.Warnf("Failed to ensure xcpretty log formatter: %s", err)
			log.Printf("Switching to xcodebuild for output tool")
			return xcodebuildFormatter
		} else {
			log.Printf("- xcpretty version: %s", ver.String())
			fmt.Println()
		}
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

	// Output tools versions
	xcodebuildVersion, err := utility.GetXcodeVersion()
	if err != nil {
		failf("Failed to get the version of xcodebuild! Error: %s", err)
	}

	log.Printf("* xcodebuild_version: %s (%s)", xcodebuildVersion.Version, xcodebuildVersion.BuildVersion)

	// xcpretty
	cfgs.OutputTool = s.selectLogFormatter(cfgs.OutputTool)

	fmt.Println()

	// setup buildActions
	buildAction := []string{}

	if cfgs.IsCleanBuild {
		buildAction = append(buildAction, "clean")
	}

	// build before test
	buildAction = append(buildAction, "build")

	// setup CommandModel for test
	testCommandModel := xcodebuild.NewTestCommand(cfgs.ProjectPath)
	testCommandModel.SetScheme(cfgs.Scheme)
	testCommandModel.SetGenerateCodeCoverage(cfgs.GenerateCodeCoverageFiles)
	testCommandModel.SetCustomBuildAction(buildAction...)

	// `Package.swift` project files do not have an xcodebuild parameter. Instead, xcodebuild needs to be run from the
	// folder where this file is located, and then it will automatically detect it.
	workDir := filepath.Dir(cfgs.ProjectPath)
	testCommandModel.SetDir(workDir)

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

	if rawXcodebuildOutput, didRun, err := runXcodebuildWithRNWrap(testCommandModel.Command().GetCmd(), cfgs.OutputTool == xcprettyFormatter); didRun {
		if err != nil {
			log.Errorf("\nLast lines of the Xcode's build log:")
			fmt.Println(stringutil.LastNLines(rawXcodebuildOutput, 10))
			failf("Test failed, error: %s", err)
		}
	} else if cfgs.OutputTool == xcprettyFormatter {
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
	logger := logV2.NewLogger()
	xcpretty := xcprettyinstaller.NewInstaller(logger, xcprettyV2.NewXcpretty(logger))

	step := NewStep(logger, xcpretty)
	step.run()
}
