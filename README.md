# Xcode Test for Mac

[![Step changelog](https://shields.io/github/v/release/bitrise-steplib/steps-xcode-test-mac?include_prereleases&label=changelog&color=blueviolet)](https://github.com/bitrise-steplib/steps-xcode-test-mac/releases)

Runs Xcode's `test` action for macOS app projects.

<details>
<summary>Description</summary>

This Step runs your pre-defined tests, and the **Deploy to Bitrise.io** Step deploys your test results to Bitrise.
You don't have to upload code signing files for this.
However, if you set a team for your project locally, in Xcode, then Xcode will ask for that team’s Developer certificate before running the test.

### Configuring the Step
This Step has a default configuration that does not need to be modified, which means that if pre-defined tests are written correctly, they will work.
Here is a rundown of the inputs should you wish to modify them.
1. Add the path of your project in the **Project (or Workspace) path** input.
2. Add the scheme name in the **Scheme name** input. Please note the scheme has to be marked as shared in Xcode.
3. Add the device or simulator on which the app will run in the **Destination** input, for example, `platform=OS X,arch=x86_64`.
4. Set the **Should a clean Xcode build run before testing?** input to `yes` to run a clean build without cache.
5. Select `yes` in **Generate code coverage files?** input if you wish to get code coverage analysis of your tests.
6. If you wish to use xcpretty formatter for your xcodebuild as an output tool, select `xcpretty` in the **Output tool** input.
If this input is set to `xcodebuild`, the raw xcodebuild output gets printed.
7. Add extra options to the end of the `xcodebuild` call in the **Additional options for xcodebuild call** input.
Use multiple options separated by a space character, for example, `-xcconfig PATH -verboseAdditional`.
9. Set the **Disable indexing during the build** input to `yes` to speed up your build.

### Troubleshooting

If your app does not have test targets defined, the primary workflow will be the only automatically created workflow and it will NOT include the **Xcode Test for Mac** Step.

### Useful links
- [Getting started with MacOS apps](https://devcenter.bitrise.io/getting-started/getting-started-with-macos-apps/)
- [About code signing](https://devcenter.bitrise.io/code-signing/code-signing-index/)

### Related Steps
- [Xcode Archive for Mac](https://www.bitrise.io/integrations/steps/xcode-archive-mac)
- [Deploy to iTunes Connect - Application Loader ](https://www.bitrise.io/integrations/steps/deploy-to-itunesconnect-application-loader)
</details>

## 🧩 Get started

Add this step directly to your workflow in the [Bitrise Workflow Editor](https://docs.bitrise.io/en/bitrise-ci/workflows-and-pipelines/steps/adding-steps-to-a-workflow.html).

You can also run this step directly with [Bitrise CLI](https://github.com/bitrise-io/bitrise).

## ⚙️ Configuration

<details>
<summary>Inputs</summary>

| Key | Description | Flags | Default |
| --- | --- | --- | --- |
| `project_path` | A `.xcodeproj`, `.xcworkspace` or `Package.swift` path, relative to the Working directory (if specified).  | required | `$BITRISE_PROJECT_PATH` |
| `scheme` | The Scheme to use. **IMPORTANT**: The Scheme have to be marked as __shared__ in Xcode!  | required | `$BITRISE_SCHEME` |
| `destination` | The Destination to use.  Read more in [Xcodebuild Destination Cheatsheet](http://www.mokacoding.com/blog/xcodebuild-destination-options/).  Example value: `platform=OS X,arch=x86_64`  |  |  |
| `is_clean_build` | Run a clean Xcode build (no incremental cache) before testing when set to `yes`. | required | `yes` |
| `generate_code_coverage_files` | Generate code coverage files alongside the test run when set to `yes`. | required | `no` |
| `output_tool` | If output_tool is set to xcpretty, the xcodebuild output will be prettified by xcpretty. If output_tool is set to xcodebuild, the raw xcodebuild output will be printed. | required | `xcpretty` |
| `xcodebuild_options` | Options added to the end of the xcodebuild call.  You can use multiple options, separated by a space character. Example: `-xcconfig PATH -verbose` |  | `CODE_SIGNING_ALLOWED='NO'` |
| `disable_index_while_building` | Could make the build faster by adding `COMPILER_INDEX_STORE_ENABLE=NO` flag to the `xcodebuild` command which will disable the indexing during the build.  Indexing is needed for  * Autocomplete * Ability to quickly jump to definition * Get class and method help by alt clicking.  Which are not needed in CI environment.  **Note:** In Xcode you can turn off the `Index-WhileBuilding` feature  by disabling the `Enable Index-WhileBuilding Functionality` in the `Build Settings`.<br/> In CI environment you can disable it by adding `COMPILER_INDEX_STORE_ENABLE=NO` flag to the `xcodebuild` command. |  | `yes` |
| `workdir` | This input is __deprecated__, please __use change-workdir step__ instead. Working directory of the step. You can leave it empty to don't change it.  |  | `$BITRISE_SOURCE_DIR` |
</details>

<details>
<summary>Outputs</summary>

| Environment Variable | Description |
| --- | --- |
| `BITRISE_XCODE_TEST_RESULT` | Result of the `xcodebuild test` run. Either `succeeded` or `failed`. |
</details>

## 🙋 Contributing

We welcome [pull requests](https://github.com/bitrise-steplib/steps-xcode-test-mac/pulls) and [issues](https://github.com/bitrise-steplib/steps-xcode-test-mac/issues) against this repository.

For pull requests, work on your changes in a forked repository and use the Bitrise CLI to [run step tests locally](https://docs.bitrise.io/en/bitrise-ci/bitrise-cli/running-your-first-local-build-with-the-cli.html).

Learn more about developing steps:

- [Create your own step](https://docs.bitrise.io/en/bitrise-ci/workflows-and-pipelines/developing-your-own-bitrise-step/developing-a-new-step.html)
