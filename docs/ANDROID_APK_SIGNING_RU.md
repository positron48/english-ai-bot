# Android APK signing для TWA/PWA

Документ описывает, где лежит локальный release keystore для сборки English/Spanish APK и какие значения нужно прописать в GitHub Actions.

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

`ANDROID_KEYSTORE_PASSWORD` и `ANDROID_KEY_PASSWORD` для этого keystore одинаковые. Это нормально для PKCS#12 и упрощает Gradle/Bubblewrap signing.

## Fingerprint для серверов

Текущий SHA-256 fingerprint:

```text
9C:12:1A:B3:43:EC:BF:15:2C:63:F9:F2:E4:96:F1:07:A4:50:63:BD:E7:88:0E:19:33:D7:09:A5:21:B4:DC:A3
```

Он уже прописан в GitOps ConfigMap:

- `devops-time-host/apps/english/base/configmap.yaml` -> `WEBAPP_ANDROID_CERT_FINGERPRINTS`;
- `devops-time-host/apps/spanish/base/configmap.yaml` -> `WEBAPP_ANDROID_CERT_FINGERPRINTS`.

Этот fingerprint нужен, чтобы backend отдавал корректный `/.well-known/assetlinks.json`, а Android мог verified-связать APK с доменом.

## Как работает CI

APK job запускается только на tag push.

Workflow:

1. GitHub Actions читает `ANDROID_KEYSTORE_BASE64`.
2. Декодирует его в `release.keystore`.
3. `actions/setup-java` ставит JDK 17, `android-actions/setup-android` ставит Android SDK.
4. `scripts/build-twa-apks.sh` заранее пишет `${HOME}/.bubblewrap/config.json` с `JAVA_HOME` и `ANDROID_HOME`, чтобы Bubblewrap не задавал интерактивный вопрос `Do you want Bubblewrap to install the JDK?`.
5. Для `iconUrl`/`maskableIconUrl` в TWA manifest CI использует raw GitHub URL вида `https://raw.githubusercontent.com/<repo>/<sha>/webapp/public/icons/<app>-512.png`. Это убирает зависимость APK build от того, успел ли новый web image выкатиться на `qantrix.ru` / `es.qantrix.ru`.
6. `scripts/build-twa-apks.sh` перед `build` выполняет `bubblewrap update --skipVersionUpgrade` внутри `dist/twa-<app>/`, чтобы создать Android project и `manifest-checksum.txt` без интерактивного regenerate prompt.
7. Bubblewrap собирает два signed TWA APK:
   - `qantrix-english-<tag>.apk` для `ru.qantrix.english` и `https://qantrix.ru/app/`;
   - `qantrix-spanish-<tag>.apk` для `ru.qantrix.spanish` и `https://es.qantrix.ru/app/`.
8. APK и `checksums.txt` загружаются в GitHub Release.

Если `ANDROID_KEYSTORE_BASE64 is required`, значит GitHub secret не добавлен или добавлен не в тот репозиторий.

Если CI падает на вопросе `Do you want Bubblewrap to install the JDK?`, значит workflow запустился без актуальной версии `scripts/build-twa-apks.sh`, где создаётся `${HOME}/.bubblewrap/config.json`, или `JAVA_HOME`/`ANDROID_HOME` не были выставлены предыдущими setup actions.

Если CI падает с `cli ERROR EISDIR: illegal operation on a directory, read`, значит Bubblewrap получил директорию вместо файла manifest. В актуальном `scripts/build-twa-apks.sh` используется `--manifest="${WORKDIR}/twa-manifest.json"`.

Если CI падает на вопросе `No checksum file was found ... would you like to regenerate your project?`, значит `build` запустился без предварительного `bubblewrap update --skipVersionUpgrade`. В актуальном `scripts/build-twa-apks.sh` update запускается перед build внутри `dist/twa-<app>/`.

Если CI падает с `Failed to download icon https://.../app/icons/<app>-512.png ... 404`, значит сборка использует старый скрипт, где icon URL брался с продового домена. В актуальном `scripts/build-twa-apks.sh` для GitHub Actions icon URL берётся из `raw.githubusercontent.com` по `GITHUB_SHA`.

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

## Важное правило

Не терять `qantrix-release.p12` и пароли из `github-actions-secrets.env`. Если keystore потерять, следующие APK с теми же package names нельзя будет установить как обновление поверх уже установленного приложения.
