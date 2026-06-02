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
ICON_URL="${ORIGIN}${ICON_PATH}"

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
  "statusBarColor": "${THEME_COLOR}",
  "statusBarColorDark": "${THEME_COLOR}",
  "navigationColor": "#000000",
  "navigationColorDark": "#000000",
  "navigationDividerColor": "#000000",
  "navigationDividerColorDark": "#000000",
  "backgroundColor": "${THEME_COLOR}",
  "enableNotifications": false,
  "startUrl": "/app/",
  "webManifestUrl": "${ORIGIN}/app/manifest.webmanifest",
  "fullScopeUrl": "${ORIGIN}/",
  "iconUrl": "${ICON_URL}",
  "maskableIconUrl": "${ICON_URL}",
  "fallbackType": "webview",
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

(
  cd "$WORKDIR"
  npx --yes @bubblewrap/cli@latest update \
    --manifest="twa-manifest.json" \
    --skipVersionUpgrade
  npx --yes @bubblewrap/cli@latest build \
    --manifest="twa-manifest.json" \
    --skipPwaValidation \
    --signingKeyPath="${KEYSTORE_PATH}" \
    --signingKeyAlias="${KEY_ALIAS}"
)

ANDROID_MANIFEST="${WORKDIR}/app/src/main/AndroidManifest.xml"
if [ -f "$ANDROID_MANIFEST" ]; then
  grep -q 'android.support.customtabs.trusted.FALLBACK_STRATEGY' "$ANDROID_MANIFEST" || {
    echo "Generated AndroidManifest.xml does not contain TWA fallback strategy metadata" >&2
    exit 1
  }
  if grep -q 'android:value="webview"' "$ANDROID_MANIFEST"; then
    :
  elif grep -q 'android:value="@string/fallbackType"' "$ANDROID_MANIFEST"; then
    STRINGS_XML="${WORKDIR}/app/src/main/res/values/strings.xml"
    grep -q '<string name="fallbackType">webview</string>' "$STRINGS_XML" || {
      echo "Generated AndroidManifest.xml references @string/fallbackType, but strings.xml is not webview" >&2
      grep -n 'fallbackType' "$STRINGS_XML" >&2 || true
      exit 1
    }
  else
    echo "Generated AndroidManifest.xml fallback strategy is not webview" >&2
    grep -n 'FALLBACK_STRATEGY\\|fallbackType\\|fallback' "$ANDROID_MANIFEST" >&2 || true
    exit 1
  fi
fi

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
