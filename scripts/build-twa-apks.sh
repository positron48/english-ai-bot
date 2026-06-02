#!/usr/bin/env bash
set -euo pipefail

if [ "$#" -ne 8 ]; then
  echo "Usage: $0 <app> <origin> <package-id> <name> <theme-color> <icon-path> <keystore-path> <key-alias>" >&2
  exit 2
fi

APP="$1"
ORIGIN="$2"
PACKAGE_ID="$3"
NAME="$4"
THEME_COLOR="$5"
ICON_PATH="$6"
KEYSTORE_PATH="$7"
KEY_ALIAS="$8"

WORKDIR="dist/twa-${APP}"
mkdir -p "$WORKDIR"

HOST="${ORIGIN#https://}"
HOST="${HOST#http://}"
HOST="${HOST%%/*}"
VERSION_NAME="${GITHUB_REF_NAME:-0.0.0}"
VERSION_CODE="${GITHUB_RUN_NUMBER:-1}"

cat > "${WORKDIR}/twa-manifest.json" <<JSON
{
  "packageId": "${PACKAGE_ID}",
  "host": "${HOST}",
  "name": "${NAME}",
  "launcherName": "${NAME}",
  "display": "standalone",
  "orientation": "portrait",
  "themeColor": "${THEME_COLOR}",
  "themeColorDark": "${THEME_COLOR}",
  "navigationColor": "#000000",
  "navigationColorDark": "#000000",
  "navigationDividerColor": "#000000",
  "navigationDividerColorDark": "#000000",
  "backgroundColor": "#ffffff",
  "enableNotifications": false,
  "startUrl": "/app/",
  "webManifestUrl": "${ORIGIN}/app/manifest.webmanifest",
  "fullScopeUrl": "${ORIGIN}/",
  "iconUrl": "${ORIGIN}${ICON_PATH}",
  "maskableIconUrl": "${ORIGIN}${ICON_PATH}",
  "fallbackType": "customtabs",
  "appVersion": "${VERSION_NAME}",
  "appVersionCode": ${VERSION_CODE},
  "splashScreenFadeOutDuration": 300,
  "signingKey": {
    "path": "${KEYSTORE_PATH}",
    "alias": "${KEY_ALIAS}"
  }
}
JSON

# Bubblewrap tries to create ~/.bubblewrap/config.json interactively on the
# first run. CI must be fully non-interactive, so write the config directly.
if [ -z "${JAVA_HOME:-}" ]; then
  echo "JAVA_HOME is required; run actions/setup-java before this script" >&2
  exit 1
fi
if [ -z "${ANDROID_HOME:-}" ]; then
  echo "ANDROID_HOME is required; run android-actions/setup-android before this script" >&2
  exit 1
fi
mkdir -p "${HOME}/.bubblewrap"
cat > "${HOME}/.bubblewrap/config.json" <<JSON
{"jdkPath":"${JAVA_HOME}","androidSdkPath":"${ANDROID_HOME}"}
JSON

npx --yes @bubblewrap/cli@latest build \
  --manifest="${WORKDIR}/twa-manifest.json" \
  --skipPwaValidation \
  --signingKeyPath="${KEYSTORE_PATH}" \
  --signingKeyAlias="${KEY_ALIAS}"

APK="${WORKDIR}/app-release-signed.apk"
if [ ! -f "$APK" ]; then
  APK="${WORKDIR}/app/build/outputs/apk/release/app-release-signed.apk"
fi
if [ ! -f "$APK" ]; then
  echo "Bubblewrap did not produce app-release-signed.apk under ${WORKDIR}" >&2
  find "$WORKDIR" -name '*.apk' -print >&2 || true
  exit 1
fi
cp "$APK" "dist/qantrix-${APP}-${VERSION_NAME}.apk"
