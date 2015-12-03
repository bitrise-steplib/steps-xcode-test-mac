#!/bin/bash
set -e

PROJECT_PATH=$BITRISE_PROJECT_PATH
SCHEME=$BITRISE_SCHEME
CLEAN_BUILD="clean"

if [ ! -z "${project_path}" ] ; then
	PROJECT_PATH="${project_path}"
fi

if [ ! -z "${scheme}" ] ; then
	SCHEME="${scheme}"
fi

if [ "${is_clean_build}" == "no" ] ; then
	CLEAN_BUILD=""
fi

if [ ! -z "${workdir}" ] ; then
	echo "==> Switching to working directory: ${workdir}"
	cd "${workdir}"
	if [ $? -ne 0 ] ; then
		echo " [!] Failed to switch to working directory: ${workdir}"
		exit 1
	fi
fi

set -o pipefail && xcodebuild -project "${PROJECT_PATH}" -scheme "${SCHEME}" "${CLEAN_BUILD}" build test | xcpretty
