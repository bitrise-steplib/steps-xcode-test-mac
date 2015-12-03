#!/bin/bash
set -e

PROJECT_PATH=$BITRISE_PROJECT_PATH

if [ ! -z "${project_path}" ] ; then
	PROJECT_PATH="${project_path}"
fi

if [ ! -z "${workdir}" ] ; then
	echo "==> Switching to working directory: ${workdir}"
	cd "${workdir}"
	if [ $? -ne 0 ] ; then
		echo " [!] Failed to switch to working directory: ${workdir}"
		exit 1
	fi
fi

set -o pipefail && xcodebuild -project "${PROJECT_PATH}" -scheme $BITRISE_SCHEME build test | xcpretty
