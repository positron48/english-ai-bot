# Настройка Nginx и SSL для English Bot

Полная инструкция по настройке Nginx с SSL сертификатом от Let's Encrypt для English Bot.

## Предварительные требования

1. **Домен настроен на ваш сервер**
   - DNS A-запись для `english.positroid.tech` должна указывать на IP вашего сервера
   - Проверьте: `dig english.positroid.tech` или `nslookup english.positroid.tech`

2. **Nginx установлен**
   ```bash
   sudo apt update
   sudo apt install nginx -y
   ```

3. **Certbot установлен**
   ```bash
   sudo apt install certbot python3-certbot-nginx -y
   ```

4. **Бот запущен и слушает порт 8184**
   - Убедитесь, что бот работает: `curl http://localhost:8184/health`
   - Если бот не запущен, запустите его перед настройкой SSL

## Шаг 1: Установка конфигурации Nginx

1. **Скопируйте конфигурацию:**
   ```bash
   sudo cp nginx.conf.example /etc/nginx/sites-available/english.positroid.tech
   ```

2. **Создайте симлинк:**
   ```bash
   sudo ln -s /etc/nginx/sites-available/english.positroid.tech /etc/nginx/sites-enabled/
   ```

3. **Проверьте конфигурацию:**
   ```bash
   sudo nginx -t
   ```
   Должно вывести: `nginx: configuration file /etc/nginx/nginx.conf test is successful`

4. **Перезагрузите Nginx:**
   ```bash
   sudo systemctl reload nginx
   ```

5. **Проверьте, что сайт доступен по HTTP:**
   ```bash
   curl http://english.positroid.tech
   ```
   Должно вернуть: `English Bot is running!`

## Шаг 2: Получение SSL сертификата

Certbot автоматически:
- Получит SSL сертификат от Let's Encrypt
- Настроит Nginx для использования HTTPS
- Настроит автоматическое перенаправление HTTP → HTTPS
- Настроит автоматическое обновление сертификата

**Выполните команду:**
```bash
sudo certbot --nginx -d english.positroid.tech
```

**Во время выполнения certbot спросит:**

1. **Email для уведомлений:**
   ```
   Enter email address (used for urgent renewal and security notices)
   ```
   Введите ваш email адрес

2. **Согласие с условиями:**
   ```
   (A)gree/(C)ancel: A
   ```

3. **Согласие на получение новостей (опционально):**
   ```
   (Y)es/(N)o: N
   ```

4. **Перенаправление HTTP на HTTPS:**
   ```
   Please choose whether or not to redirect HTTP traffic to HTTPS, removing HTTP access.
   1: No redirect - Make no further changes to the webserver configuration.
   2: Redirect - Make all requests redirect to secure HTTPS access.
   ```
   Выберите **2** (Redirect) для автоматического перенаправления HTTP на HTTPS

**После успешного выполнения вы увидите:**
```
Successfully received certificate.
Certificate is saved at: /etc/letsencrypt/live/english.positroid.tech/fullchain.pem
Key is saved at:         /etc/letsencrypt/live/english.positroid.tech/privkey.pem
This certificate expires on YYYY-MM-DD.
These files will be updated when the certificate renews.
Certbot has set up a scheduled task to automatically renew this certificate in the background.

- - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
If you like Certbot, please consider supporting our work by:
 * Donating to ISRG / Let's Encrypt:   https://letsencrypt.org/donate
 * Donating to EFF:                    https://eff.org/donate-le
- - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
```

## Шаг 3: Проверка работы

1. **Проверьте HTTPS:**
   ```bash
   curl https://english.positroid.tech
   ```
   Должно вернуть: `English Bot is running!`

2. **Проверьте автоматическое перенаправление:**
   ```bash
   curl -I http://english.positroid.tech
   ```
   Должно вернуть: `HTTP/1.1 301 Moved Permanently` с заголовком `Location: https://english.positroid.tech/...`

3. **Проверьте webhook:**
   ```bash
   curl -X POST https://english.positroid.tech/webhook
   ```
   (Может вернуть ошибку, но это нормально - главное, что endpoint доступен)

4. **Проверьте Mini App:**
   Откройте в браузере: `https://english.positroid.tech/app`

## Шаг 4: Автоматическое обновление сертификата

Certbot автоматически настроил задачу для обновления сертификата. Проверьте:

```bash
sudo systemctl status certbot.timer
```

Должно быть: `Active: active (waiting)`

**Проверьте, что обновление работает:**
```bash
sudo certbot renew --dry-run
```

Если команда выполнилась без ошибок, автоматическое обновление настроено правильно.

## Шаг 5: Настройка бота

Убедитесь, что в `.env` файле бота указаны правильные настройки:

```env
# Webhook настройки
TELEGRAM_WEBHOOK_ENABLE=true
TELEGRAM_WEBHOOK_DOMAIN=https://english.positroid.tech
TELEGRAM_WEBHOOK_PATH=/webhook

# Mini App настройки
WEBAPP_PUBLIC_URL=https://english.positroid.tech

# Порт сервера
SERVER_ADDRESS=:8184
```

После изменения `.env` перезапустите бота:
```bash
make restart
# или
systemctl --user restart english-bot
```

## Управление сертификатом

### Просмотр информации о сертификате
```bash
sudo certbot certificates
```

### Обновление сертификата вручную
```bash
sudo certbot renew
```

### Отзыв сертификата (если нужно)
```bash
sudo certbot revoke --cert-path /etc/letsencrypt/live/english.positroid.tech/cert.pem
```

### Удаление конфигурации certbot
```bash
sudo certbot delete --cert-name english.positroid.tech
```

## Устранение проблем

### Проблема: "Failed to obtain certificate"

**Причины:**
- DNS не настроен или еще не распространился
- Порт 80 заблокирован файрволом
- Домен уже используется другим сертификатом

**Решение:**
1. Проверьте DNS: `dig english.positroid.tech`
2. Проверьте, что порт 80 открыт: `sudo ufw allow 80/tcp`
3. Проверьте, что Nginx слушает порт 80: `sudo netstat -tuln | grep :80`

### Проблема: "Connection refused" при проверке webhook

**Причина:** Бот не запущен или слушает другой порт

**Решение:**
1. Проверьте, что бот запущен: `systemctl --user status english-bot`
2. Проверьте порт: `netstat -tuln | grep 8184`
3. Проверьте логи бота: `journalctl --user -u english-bot -f`

### Проблема: Сертификат не обновляется автоматически

**Решение:**
1. Проверьте таймер: `sudo systemctl status certbot.timer`
2. Включите таймер: `sudo systemctl enable certbot.timer`
3. Запустите таймер: `sudo systemctl start certbot.timer`

### Проблема: Nginx не перезагружается после certbot

**Решение:**
```bash
sudo nginx -t  # Проверьте конфигурацию
sudo systemctl reload nginx  # Перезагрузите вручную
```

## Дополнительные настройки безопасности

После настройки SSL рекомендуется:

1. **Настроить HSTS (HTTP Strict Transport Security):**
   Certbot уже добавит базовые настройки, но можно улучшить в `/etc/nginx/sites-available/english.positroid.tech`:
   ```nginx
   add_header Strict-Transport-Security "max-age=31536000; includeSubDomains" always;
   ```

2. **Ограничить версии SSL/TLS:**
   Certbot уже настроит безопасные версии, но можно проверить в конфиге.

3. **Настроить файрвол:**
   ```bash
   sudo ufw allow 80/tcp
   sudo ufw allow 443/tcp
   sudo ufw enable
   ```

## Полезные команды

```bash
# Проверка конфигурации Nginx
sudo nginx -t

# Перезагрузка Nginx
sudo systemctl reload nginx

# Просмотр логов Nginx
sudo tail -f /var/log/nginx/english.positroid.tech-access.log
sudo tail -f /var/log/nginx/english.positroid.tech-error.log

# Проверка статуса certbot
sudo systemctl status certbot.timer

# Тестовое обновление сертификата
sudo certbot renew --dry-run

# Просмотр всех сертификатов
sudo certbot certificates
```

## Что делает certbot

После выполнения `certbot --nginx -d english.positroid.tech`, certbot:

1. ✅ Добавит в конфиг Nginx:
   - `listen 443 ssl;`
   - `ssl_certificate /etc/letsencrypt/live/english.positroid.tech/fullchain.pem;`
   - `ssl_certificate_key /etc/letsencrypt/live/english.positroid.tech/privkey.pem;`
   - `include /etc/letsencrypt/options-ssl-nginx.conf;`
   - `ssl_dhparam /etc/letsencrypt/ssl-dhparams.pem;`

2. ✅ Создаст второй `server` блок для HTTP (порт 80) с автоматическим перенаправлением на HTTPS

3. ✅ Настроит автоматическое обновление сертификата через systemd timer

**Важно:** Не редактируйте строки, помеченные `# managed by Certbot` вручную - certbot будет перезаписывать их при обновлении.

