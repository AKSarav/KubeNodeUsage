#!/usr/bin/env bash

Version=$1

if [ -z $Version ]; then
	echo "Version Identifier is required"
	exit 1
fi

package_name="KubeNodeUsage"

platforms=("windows/amd64" "windows/386" "darwin/amd64" "darwin/arm64" "windows/arm64" "windows/arm" "linux/arm64" "linux/amd64" "linux/arm")

for platform in "${platforms[@]}"
do
	platform_split=(${platform//\// })
	GOOS=${platform_split[0]}
	GOARCH=${platform_split[1]}
	output_name=$package_name'-'$GOOS'-'$GOARCH-$Version
	if [ $GOOS = "windows" ]; then
		output_name+='.exe'
	fi

	env GOOS=$GOOS GOARCH=$GOARCH go build -o releases/$output_name

	# Compress the binary with zip
	zip -j releases/$output_name.zip releases/$output_name

	# Remove the binary
	rm releases/$output_name

	if [ $? -ne 0 ]; then
   		echo 'An error has occurred! Aborting the script execution...'
		exit 1
	fi
done
