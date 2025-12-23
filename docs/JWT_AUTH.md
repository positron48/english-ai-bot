# JWT Authentication

Проект использует JWT (JSON Web Tokens) для аутентификации вместо cookie-based сессий.

## Настройка

Установите переменную окружения `WEBAPP_JWT_SECRET` (или используйте `WEBAPP_SESSION_SECRET` как fallback):

```bash
export WEBAPP_JWT_SECRET="your-secret-key-here"
```

Опционально можно настроить время жизни токена:
```bash
export WEBAPP_JWT_TTL_HOURS=720  # По умолчанию 720 часов (30 дней)
```

## Использование

### 1. Получение JWT токена

#### Через OTP:
```bash
# Запрос OTP кода
curl -X POST http://localhost:8184/auth/request_otp \
  -d "username=your_username"

# Проверка OTP и получение токена
curl -X POST http://localhost:8184/auth/otp \
  -d "user_id=123" \
  -d "code=123456"
```

Ответ:
```json
{
  "success": true,
  "message": "Authentication successful",
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "token_type": "Bearer"
}
```

#### Через Telegram WebApp:
```bash
curl -X POST http://localhost:8184/auth/telegram \
  -d "initData=your_telegram_init_data"
```

### 2. Использование токена

Добавьте токен в заголовок `Authorization`:

```bash
curl -X GET http://localhost:8184/app/dashboard \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
```

### 3. В Swagger UI

1. Выполните запрос на `/auth/otp` или `/auth/telegram`
2. Скопируйте полученный `token` из ответа
3. Нажмите кнопку "Authorize" в Swagger UI
4. Введите: `Bearer <ваш_токен>`
5. Теперь все запросы будут использовать этот токен

## Структура токена

JWT токен содержит:
- `user_id` - ID пользователя
- `exp` - Время истечения токена
- `iat` - Время выдачи токена
- `iss` - Issuer (всегда "english-bot")

## Безопасность

- Токены подписываются с использованием HMAC-SHA256
- Токены имеют срок действия (по умолчанию 30 дней)
- Храните секретный ключ в безопасности
- Используйте HTTPS в production

## Обратная совместимость

Система поддерживает обратную совместимость с cookie-based сессиями во время миграции. Если токен не найден в заголовке Authorization, система попытается использовать cookie сессию.

