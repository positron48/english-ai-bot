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
ENGLISH_ICON="${ROOT_DIR}/webapp/public/icons/english-512.png"
SPANISH_ICON="${ROOT_DIR}/webapp/public/icons/spanish-512.png"
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

mkdir -p "${ANDROID_DIR}/app/src/english/res/mipmap-nodpi" "${ANDROID_DIR}/app/src/spanish/res/mipmap-nodpi"
cp "${ENGLISH_ICON}" "${ANDROID_DIR}/app/src/english/res/mipmap-nodpi/ic_launcher.png"
cp "${SPANISH_ICON}" "${ANDROID_DIR}/app/src/spanish/res/mipmap-nodpi/ic_launcher.png"

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
  assembleEnglishRelease \
  assembleSpanishRelease

cp "${ANDROID_DIR}/app/build/outputs/apk/english/release/app-english-release.apk" "${ROOT_DIR}/dist/qantrix-english-${VERSION_NAME}.apk"
cp "${ANDROID_DIR}/app/build/outputs/apk/spanish/release/app-spanish-release.apk" "${ROOT_DIR}/dist/qantrix-spanish-${VERSION_NAME}.apk"
