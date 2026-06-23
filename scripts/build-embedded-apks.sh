#!/usr/bin/env bash
set -euo pipefail

if [ "$#" -ne 3 ]; then
  echo "Usage: $0 <version-name> <keystore-path> <key-alias>" >&2
  exit 2
fi

VERSION_NAME="$1"
KEYSTORE_PATH="$2"
KEY_ALIAS="$3"

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
ANDROID_DIR="${ROOT_DIR}/android-embedded"
ASSETS_DIR="${ANDROID_DIR}/app/src/main/assets/public/app"
DIST_DIR="${ROOT_DIR}/webapp/dist"
GRADLE_VERSION="${GRADLE_VERSION:-8.10.2}"
GRADLE_DIR="${ROOT_DIR}/dist/gradle-${GRADLE_VERSION}"
GRADLE_ZIP="${ROOT_DIR}/dist/gradle-${GRADLE_VERSION}-bin.zip"

if [ ! -d "${DIST_DIR}" ]; then
  echo "${DIST_DIR} does not exist; run webapp build first" >&2
  exit 1
fi

mkdir -p "${ROOT_DIR}/dist"
rm -rf "${ASSETS_DIR}"
mkdir -p "${ASSETS_DIR}"
cp -R "${DIST_DIR}/." "${ASSETS_DIR}/"

# Launcher icon (mipmap/ic_launcher) is committed under src/main/res — no injection needed.

if [ ! -x "${GRADLE_DIR}/bin/gradle" ]; then
  curl -fsSL "https://services.gradle.org/distributions/gradle-${GRADLE_VERSION}-bin.zip" -o "${GRADLE_ZIP}"
  rm -rf "${GRADLE_DIR}"
  unzip -q "${GRADLE_ZIP}" -d "${ROOT_DIR}/dist"
fi

export ANDROID_VERSION_NAME="${VERSION_NAME}"
export ANDROID_VERSION_CODE="${GITHUB_RUN_NUMBER:-1}"
KEYSTORE_PASSWORD="${BUBBLEWRAP_KEYSTORE_PASSWORD:?BUBBLEWRAP_KEYSTORE_PASSWORD is required}"
KEY_PASSWORD="${BUBBLEWRAP_KEY_PASSWORD:?BUBBLEWRAP_KEY_PASSWORD is required}"

"${GRADLE_DIR}/bin/gradle" \
  --no-daemon \
  --project-dir "${ANDROID_DIR}" \
  -Pandroid.injected.signing.store.file="${KEYSTORE_PATH}" \
  -Pandroid.injected.signing.store.password="${KEYSTORE_PASSWORD}" \
  -Pandroid.injected.signing.key.alias="${KEY_ALIAS}" \
  -Pandroid.injected.signing.key.password="${KEY_PASSWORD}" \
  assembleRelease

cp "${ANDROID_DIR}/app/build/outputs/apk/release/app-release.apk" "${ROOT_DIR}/dist/qantrix-linglow-${VERSION_NAME}.apk"
