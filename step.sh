#!/bin/bash
set -e

if [ ! -z "${workdir}" ] ; then
	echo "==> Switching to working directory: ${workdir}"
	cd "${workdir}"
	if [ $? -ne 0 ] ; then
		echo " [!] Failed to switch to working directory: ${workdir}"
		exit 1
	fi
fi

set -o pipefail && xcodebuild -project $BITRISE_PROJECT_PATH -scheme $BITRISE_SCHEME build test | xcpretty
