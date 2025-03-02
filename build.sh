#!/usr/bin/env bash

Version=$1

if [ -z $Version ]; then
    echo "Version Identifier is required"
    exit 1
fi

# check if the input is only numbers with a dot
if ! [[ $Version =~ ^[0-9]+(\.[0-9]+)*$ ]]; then
	echo "Version Identifier should be in the format of x.y.z"
	exit 1
fi


package_name="KubeNodeUsage"

platforms=("linux/amd64" "linux/arm64" "darwin/amd64" "darwin/arm64"  "windows/amd64" "windows/arm64" "windows/386" "windows/arm" "linux/arm")

# Create a temporary file to store the SHA256 checksums
checksum_file=$(mktemp)

for platform in "${platforms[@]}"
do
    platform_split=(${platform//\// })
    GOOS=${platform_split[0]}
    GOARCH=${platform_split[1]}
    output_name=$package_name'-'$GOOS'-'$GOARCH'-v'$Version
    if [ $GOOS = "windows" ]; then
        output_name+='.exe'
    fi

    env GOOS=$GOOS GOARCH=$GOARCH go build -o releases/$output_name

    # Compress the binary with zip
	echo "Creating zip file for $GOOS/$GOARCH"
    zip -j releases/$output_name.zip releases/$output_name

	echo "Calculating sha256sum for $GOOS/$GOARCH"
    # Calculate the SHA256 checksum and store it in the temporary file
    sha256sum=$(shasum -a 256 releases/$output_name.zip | awk '{ print $1 }')
    echo "$GOOS/$GOARCH $sha256sum" >> $checksum_file

    # Remove the binary
	echo "Removing binary for $GOOS/$GOARCH", $output_name
    rm releases/$output_name

    if [ $? -ne 0 ]; then
        echo 'An error has occurred! Aborting the script execution...'
        exit 1
    fi
done

# Find the version declaration on the main.go file and change it to the new version
# var semver = "v3.0.2" 
echo "Updating the version in main.go"
sed -i '' "s/var semver = \"v[0-9]*\.[0-9]*\.[0-9]*\"/var semver = \"v$Version\"/g" main.go

# Create the Homebrew formula
cat <<EOF > kubenodeusage.rb
class Kubenodeusage < Formula
    desc "KubeNodeUsage is a command line utility to get the usage of the nodes and pods in a Kubernetes cluster graphically."
    homepage "https://github.com/AKSarav/KubeNodeUsage"
    version "$Version"
    license "MIT"

EOF

while read -r line; do
    platform=$(echo $line | awk '{ print $1 }')
    sha256=$(echo $line | awk '{ print $2 }')
    platform_split=(${platform//\// })
    GOOS=${platform_split[0]}
    GOARCH=${platform_split[1]}

    if [ "$GOOS" == "linux" ] && [ "$GOARCH" == "amd64" ]; then
        cat <<EOF >> kubenodeusage.rb
    if OS.linux? && Hardware::CPU.intel?
      url "https://github.com/AKSarav/KubeNodeUsage/releases/download/v$Version/KubeNodeUsage-linux-amd64-v$Version.zip"
      sha256 "$sha256"
    elsif OS.linux? && Hardware::CPU.arm?
EOF
    elif [ "$GOOS" == "linux" ] && [ "$GOARCH" == "arm64" ]; then
        cat <<EOF >> kubenodeusage.rb
      url "https://github.com/AKSarav/KubeNodeUsage/releases/download/v$Version/KubeNodeUsage-linux-arm64-v$Version.zip"
      sha256 "$sha256"
    elsif OS.mac? && Hardware::CPU.intel?
EOF
    elif [ "$GOOS" == "darwin" ] && [ "$GOARCH" == "amd64" ]; then
        cat <<EOF >> kubenodeusage.rb
      url "https://github.com/AKSarav/KubeNodeUsage/releases/download/v$Version/KubeNodeUsage-darwin-amd64-v$Version.zip"
      sha256 "$sha256"
    elsif OS.mac? && Hardware::CPU.arm?
EOF
    elif [ "$GOOS" == "darwin" ] && [ "$GOARCH" == "arm64" ]; then
        cat <<EOF >> kubenodeusage.rb
      url "https://github.com/AKSarav/KubeNodeUsage/releases/download/v$Version/KubeNodeUsage-darwin-arm64-v$Version.zip"
      sha256 "$sha256"
    elsif OS.windows? && Hardware::CPU.intel?
EOF
    elif [ "$GOOS" == "windows" ] && [ "$GOARCH" == "amd64" ]; then
        cat <<EOF >> kubenodeusage.rb
      url "https://github.com/AKSarav/KubeNodeUsage/releases/download/v$Version/KubeNodeUsage-windows-amd64-v$Version.exe.zip"
      sha256 "$sha256"
    elsif OS.windows? && Hardware::CPU.arm?
EOF
    elif [ "$GOOS" == "windows" ] && [ "$GOARCH" == "arm64" ]; then
        cat <<EOF >> kubenodeusage.rb
      url "https://github.com/AKSarav/KubeNodeUsage/releases/download/v$Version/KubeNodeUsage-windows-arm64-v$Version.exe.zip"
      sha256 "$sha256"
    end
EOF
    fi
done < $checksum_file

cat <<EOF >> kubenodeusage.rb

    def install
    if OS.mac? && Hardware::CPU.intel?
      bin.install "KubeNodeUsage-darwin-amd64-v#{version}" => "KubeNodeUsage"
    elsif OS.mac? && Hardware::CPU.arm?
      bin.install "KubeNodeUsage-darwin-arm64-v#{version}" => "KubeNodeUsage"
    elsif OS.linux? && Hardware::CPU.intel?
      bin.install "KubeNodeUsage-linux-amd64-v#{version}" => "KubeNodeUsage"
    elsif OS.linux? && Hardware::CPU.arm?
      bin.install "KubeNodeUsage-linux-arm64-v#{version}" => "KubeNodeUsage"
    end
  end

  test do
    system "#{bin}/KubeNodeUsage", "--version"
  end
end
EOF

# Clean up the temporary file
rm $checksum_file

echo "Homebrew formula created: kubenodeusage.rb"