#!/bin/bash

rm -r ./build/macos/Muscrat.app > /dev/null 2>&1 || true
mkdir -p ./build/macos/Muscrat.app/Contents/{MacOS,Resources}

cp ./build/bin/mrat ./build/macos/Muscrat.app/Contents/MacOS/mrat

################################################################################
# create icon.icns
pushd ./data/icons

ICONDIR=myicons.iconset
mkdir -p $ICONDIR

function create_icon {
    SIZE=${1}x${1}
    magick muscrat.svg -background none \
           -transparent white \
           -gravity center \
           -resize ${SIZE} \
           -extent ${SIZE} \
           ${ICONDIR}/icon_${SIZE}.png
}

create_icon 1024

SIZES=(16 32 64 128 256 512)
for SIZE in ${SIZES[@]}; do
    create_icon ${SIZE}
done

# make @2x icons
for SIZE in ${SIZES[@]}; do
    DOUBLE=$((SIZE * 2))
    cp ${ICONDIR}/icon_${DOUBLE}x${DOUBLE}.png ${ICONDIR}/icon_${SIZE}x${SIZE}@2x.png
done

popd
iconutil -c icns -o ./build/macos/Muscrat.app/Contents/Resources/icon.icns "./data/icons/$ICONDIR"
################################################################################

################################################################################
# Copy samples to app bundle

cp -r ./data/samples ./build/macos/Muscrat.app/Contents/Resources/samples

################################################################################

# create Info.plist
cat > ./build/macos/Muscrat.app/Contents/Info.plist <<EOF
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
	<key>CFBundleExecutable</key>
	<string>mrat</string>
	<key>CFBundleIconFile</key>
	<string>icon.icns</string>
	<key>CFBundleIdentifier</key>
	<string>com.muscratsynth.mrat</string>
	<key>NSHighResolutionCapable</key>
	<true/>
	<key>LSUIElement</key>
	<true/>
</dict>
</plist>
EOF
