#!/bin/bash
set -e

#
# Validate parameters
echo "Configs:"
echo "* workdir: $workdir"
echo "* project_path: $project_path"
echo "* scheme: $scheme"
echo "* is_clean_build: $is_clean_build"
echo "* generate_code_coverage_files: $generate_code_coverage_files"

echo

# Detect Xcode major version
xcode_major_version=""
major_version_regex="Xcode ([0-9]).[0-9]"
out=$(xcodebuild -version)
if [[ "${out}" =~ ${major_version_regex} ]] ; then
	xcode_major_version="${BASH_REMATCH[1]}"
fi

IFS=$'\n'
xcodebuild_version_split=($out)
unset IFS

xcodebuild_version="${xcodebuild_version_split[0]} (${xcodebuild_version_split[1]})"
echo "* xcodebuild_version: $xcodebuild_version"

# Detect xcpretty version
xcpretty_version=""
if [[ "${output_tool}" == "xcpretty" ]] ; then
	xcpretty_version=$(xcpretty --version)
	exit_code=$?
	if [[ $exit_code != 0 || -z "$xcpretty_version" ]] ; then
		echo "xcpretty is not installed
		For xcpretty installation see: 'https://github.com/supermarin/xcpretty',
		or use 'xcodebuild' as 'output_tool'.
		"
	fi

	echo "* xcpretty_version: $xcpretty_version"
fi

echo

# Run test
BUILD_COMMAND=""
GENERATE_CODE_COVERAGE_FILES=
XCPROJECT_OR_WORKSPACE=""

if [ "${is_clean_build}" == "yes" ] ; then
	BUILD_COMMAND="clean"
fi

if [ "${generate_code_coverage_files}" == "yes" ] ; then
	GENERATE_CODE_COVERAGE_FILES="GCC_INSTRUMENT_PROGRAM_FLOW_ARCS=YES GCC_GENERATE_TEST_COVERAGE_FILES=YES"
fi

if [[ "${project_path}" == *".xcworkspace"* ]]
then
  XCPROJECT_OR_WORKSPACE="-workspace ${project_path}"
fi

if [[ "${project_path}" == *".xcodeproj"* ]]
then
  XCPROJECT_OR_WORKSPACE="-project ${project_path}"
fi

if [ ! -z "${workdir}" ] ; then
	echo "==> Switching to working directory: ${workdir}"
	cd "${workdir}"
	if [ $? -ne 0 ] ; then
		echo " [!] Failed to switch to working directory: ${workdir}"
		exit 1
	fi
fi

set -x

if [ $xcpretty_version != "" ] ; then
	set -o pipefail && xcodebuild ${XCPROJECT_OR_WORKSPACE} -scheme "${scheme}" ${BUILD_COMMAND} build test ${GENERATE_CODE_COVERAGE_FILES} | xcpretty
else
	xcodebuild ${XCPROJECT_OR_WORKSPACE} -scheme "${scheme}" ${BUILD_COMMAND} build test ${GENERATE_CODE_COVERAGE_FILES}
fi

ret=$?
set +x

if [ $ret -eq 0 ] ;
then BITRISE_XCODE_TEST_RESULT="succeeded"
else BITRISE_XCODE_TEST_RESULT="failed"
fi

envman add --key BITRISE_XCODE_TEST_RESULT --value "${BITRISE_XCODE_TEST_RESULT}"
