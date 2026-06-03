# Android APK signing для embedded APK

Документ описывает, где лежит локальный release keystore для сборки English/Spanish embedded WebView APK и какие значения нужно прописать в GitHub Actions.

## Что уже сгенерировано

Локальные файлы подписи лежат в ignored-директории:

```text
english-ai-bot/secrets/android/
```

Содержимое:

- `qantrix-release.p12` — release keystore в формате PKCS#12;
- `qantrix-release.cert.pem` — публичный сертификат, из него берётся SHA-256 fingerprint для `assetlinks.json`;
- `qantrix-release.key.pem` — приватный ключ;
- `github-actions-secrets.env` — готовые значения для GitHub Actions secrets;
- `fingerprint.txt` — SHA-256 fingerprint сертификата.

Директория `secrets/` добавлена в `.gitignore`; эти файлы нельзя коммитить.

## Что прописать в GitHub

Открыть репозиторий `english-ai-bot`:

```text
Settings -> Secrets and variables -> Actions -> New repository secret
```

Добавить 4 repository secrets из файла:

```text
english-ai-bot/secrets/android/github-actions-secrets.env
```

Имена secrets должны быть строго такими:

```text
ANDROID_KEYSTORE_BASE64
ANDROID_KEYSTORE_PASSWORD
ANDROID_KEY_ALIAS
ANDROID_KEY_PASSWORD
```

`ANDROID_KEY_ALIAS` сейчас:

```text
qantrix-release
```

`ANDROID_KEYSTORE_PASSWORD` и `ANDROID_KEY_PASSWORD` для этого keystore одинаковые. Это нормально для PKCS#12 и упрощает Gradle signing.

## Fingerprint для серверов

Текущий SHA-256 fingerprint:

```text
9C:12:1A:B3:43:EC:BF:15:2C:63:F9:F2:E4:96:F1:07:A4:50:63:BD:E7:88:0E:19:33:D7:09:A5:21:B4:DC:A3
```

Он уже прописан в GitOps ConfigMap:

- `devops-time-host/apps/english/base/configmap.yaml` -> `WEBAPP_ANDROID_CERT_FINGERPRINTS`;
- `devops-time-host/apps/spanish/base/configmap.yaml` -> `WEBAPP_ANDROID_CERT_FINGERPRINTS`.

Этот fingerprint остаётся полезным для web/PWA и совместимости с App Links, но embedded APK больше не зависит от verified TWA: frontend shell лежит внутри APK.

## Как работает CI

APK job запускается только на tag push.

Workflow:

1. GitHub Actions читает `ANDROID_KEYSTORE_BASE64`.
2. Декодирует его в `release.keystore`.
3. `actions/setup-java` ставит JDK 17, `android-actions/setup-android` ставит Android SDK.
4. Workflow собирает `webapp/dist`.
5. `scripts/build-embedded-apks.sh` копирует `webapp/dist` в Android assets.
6. Gradle собирает два signed embedded APK:
   - `qantrix-english-<tag>.apk` для `ru.qantrix.english`, frontend origin `https://qantrix.ru/app/`;
   - `qantrix-spanish-<tag>.apk` для `ru.qantrix.spanish`, frontend origin `https://es.qantrix.ru/app/`.
7. APK и `checksums.txt` загружаются в GitHub Release.

В APK лежит весь frontend shell: `index.html`, Vite JS/CSS chunks, icons, manifest и `asset-manifest.json`. WebView подменяет локальными assets только `/app/*`; `/api/*`, `/auth/*` и другие backend requests уходят в production host.

Если `ANDROID_KEYSTORE_BASE64 is required`, значит GitHub secret не добавлен или добавлен не в тот репозиторий.

Если CI падает на `SDK location not found` или `failed to find target android-35`, проверить шаг `Install Android SDK packages` в `.github/workflows/ci.yml`.

Если CI падает на Gradle dependency resolution, проверить доступность `google()`/`mavenCentral()` и версию Android Gradle Plugin в `android-embedded/build.gradle`.

## Проверка локальных значений

Посмотреть fingerprint:

```bash
cat english-ai-bot/secrets/android/fingerprint.txt
```

Проверить, что keystore base64 не пустой:

```bash
grep '^ANDROID_KEYSTORE_BASE64=' english-ai-bot/secrets/android/github-actions-secrets.env | wc -c
```

Проверить, что файл не попадёт в git:

```bash
git -C english-ai-bot status --ignored --short secrets/android
```

Проверить Digital Asset Links на prod, если нужно валидировать web/PWA совместимость:

```bash
curl -sS https://qantrix.ru/.well-known/assetlinks.json | jq .
curl -sS https://es.qantrix.ru/.well-known/assetlinks.json | jq .
```

В ответе должны быть package names `ru.qantrix.english` / `ru.qantrix.spanish` и тот же SHA-256 fingerprint, которым подписан APK.

## Проверка prod перед rerun APK job

Перед rerun failed APK job проверить, что English/Spanish prod уже отдают PWA manifest и иконки:

```bash
curl -fsSI https://qantrix.ru/app/manifest.webmanifest
curl -fsSI https://qantrix.ru/app/icons/english-512.png
curl -fsSI https://es.qantrix.ru/app/manifest.webmanifest
curl -fsSI https://es.qantrix.ru/app/icons/spanish-512.png
```

Все четыре команды должны вернуть `2xx`. Если там `404`, значит Flux/k3s ещё не выкатил новый image или запрос попадает в старый pod.

## Важное правило

Не терять `qantrix-release.p12` и пароли из `github-actions-secrets.env`. Если keystore потерять, следующие APK с теми же package names нельзя будет установить как обновление поверх уже установленного приложения.
