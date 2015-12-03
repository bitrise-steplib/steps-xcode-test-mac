#!/bin/bash
set -e

PROJECT_PATH=$BITRISE_PROJECT_PATH
SCHEME=$BITRISE_SCHEME
BUILD_COMMAND="clean build"
GENERATE_CODE_COVERAGE_FILES="no"

export BITRISE_XCODE_TEST_RESULT

if [ ! -z "${project_path}" ] ; then
	PROJECT_PATH="${project_path}"
fi

if [ ! -z "${scheme}" ] ; then
	SCHEME="${scheme}"
fi

if [ "${is_clean_build}" == "no" ] ; then
	BUILD_COMMAND="build"
fi

if [ "${generate_code_coverage_files}" == "yes" ] ; then
	GENERATE_CODE_COVERAGE_FILES="GCC_INSTRUMENT_PROGRAM_FLOW_ARCS=YES GCC_GENERATE_TEST_COVERAGE_FILES=YES"
fi

if [ ! -z "${workdir}" ] ; then
	echo "==> Switching to working directory: ${workdir}"
	cd "${workdir}"
	if [ $? -ne 0 ] ; then
		echo " [!] Failed to switch to working directory: ${workdir}"
		exit 1
	fi
fi

if set -o pipefail && xcodebuild -project "${PROJECT_PATH}" -scheme "${SCHEME}" "${BUILD_COMMAND}" test "${GENERATE_CODE_COVERAGE_FILES}" | xcpretty
then BITRISE_XCODE_TEST_RESULT="succeeded"
else BITRISE_XCODE_TEST_RESULT="failed"
fi
