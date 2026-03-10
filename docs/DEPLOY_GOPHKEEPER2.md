# Деплой gophkeeper2 на сервер (как pointscounter)

Чеклист по документу **SERVER_AND_CI_SETUP_FOR_NEW_PROJECT.md**: один хост, общий PostgreSQL и nginx, отдельный runner для репо gophkeeper2.

---

## 1. Cloudflare DNS (домен ampleev.com)

Добавить **одну A-запись**:

| Type | Name            | Content       | Proxy / TTL |
|------|-----------------|---------------|-------------|
| A    | api-gophkeeper2 | 212.193.50.194 | DNS only, Auto |

Полное имя хоста: **api-gophkeeper2.ampleev.com**.

(Остальные записи не трогать.)

---

## 2. Порты (на том же сервере, что и pointscounter)

| Окружение | HTTP API |
|-----------|----------|
| Прод      | 8083     |
| Тест      | 8084     |

Метрики пока не используются. Убедиться, что 8083 и 8084 свободны (на сервере: `ss -tlnp | grep -E '8083|8084'`).

---

## 3. Базы данных (PostgreSQL на сервере)

На том же инстансе PostgreSQL, что и для pointscounter:

```bash
sudo -u postgres psql -c "CREATE DATABASE gophkeeper2;"
sudo -u postgres psql -c "CREATE DATABASE gophkeeper2_test;"
# Если используете отдельного пользователя (как gopher для pointscounter):
# sudo -u postgres psql -c "CREATE USER gophkeeper2_user WITH PASSWORD '...';"
# sudo -u postgres psql -c "GRANT ALL PRIVILEGES ON DATABASE gophkeeper2 TO gophkeeper2_user;"
# sudo -u postgres psql -c "GRANT ALL PRIVILEGES ON DATABASE gophkeeper2_test TO gophkeeper2_user;"
```

Либо использовать существующего пользователя (например `postgres`) и выдать ему права на новые БД. В секретах GitHub тогда задать `DATABASE_URL` и `DATABASE_URL_TEST` в формате `postgresql://user:password@localhost:5432/gophkeeper2?sslmode=disable`.

---

## 4. Каталоги и код на сервере

```bash
sudo mkdir -p /opt/gophkeeper2
sudo chown "$USER" /opt/gophkeeper2
cd /opt/gophkeeper2
git clone https://github.com/<ваш-аккаунт>/gophkeeper2.git
cd gophkeeper2
```

Деплой-скрипт в workflow делает `git fetch` + `git reset --hard origin/main` в этом каталоге, билд и рестарт сервисов. Имя каталога в workflow: `PROJECT_DIR=/opt/gophkeeper2/gophkeeper2`.

---

## 5. Systemd

**Прод** — `/etc/systemd/system/gophkeeper2.service`:

```ini
[Unit]
Description=gophkeeper2 Backend API (production)
After=network.target postgresql.service

[Service]
Type=simple
User=root
WorkingDirectory=/opt/gophkeeper2/gophkeeper2
EnvironmentFile=/opt/gophkeeper2/gophkeeper2/.env
ExecStart=/opt/gophkeeper2/gophkeeper2/gophkeeper2
Restart=always
RestartSec=5
Environment=PORT=8083

[Install]
WantedBy=multi-user.target
```

**Тест** — `/etc/systemd/system/gophkeeper2-test.service`:

```ini
[Unit]
Description=gophkeeper2 Backend (TEST - port 8084)
After=network.target postgresql.service

[Service]
Type=simple
User=root
WorkingDirectory=/opt/gophkeeper2/gophkeeper2
EnvironmentFile=/opt/gophkeeper2/gophkeeper2/.env.test
ExecStart=/opt/gophkeeper2/gophkeeper2/gophkeeper2
Restart=on-failure
RestartSec=5
Environment=PORT=8084

[Install]
WantedBy=multi-user.target
```

Включение и первый запуск (после первого деплоя, когда уже есть бинарь и .env):

```bash
sudo systemctl daemon-reload
sudo systemctl enable gophkeeper2 gophkeeper2-test
sudo systemctl start gophkeeper2
sudo systemctl start gophkeeper2-test
```

---

## 6. Nginx и SSL

1. SSL-сертификат (Let's Encrypt):

```bash
sudo certbot certonly --nginx -d api-gophkeeper2.ampleev.com
```

2. Конфиг — `/etc/nginx/sites-available/api-gophkeeper2.ampleev.com` (по образцу pointscounter):

```nginx
# Редирект HTTP → HTTPS
server {
    listen 80;
    server_name api-gophkeeper2.ampleev.com;
    return 301 https://$server_name$request_uri;
}

server {
    listen 443 ssl;
    server_name api-gophkeeper2.ampleev.com;

    ssl_certificate     /etc/letsencrypt/live/api-gophkeeper2.ampleev.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/api-gophkeeper2.ampleev.com/privkey.pem;

    location / {
        proxy_pass http://127.0.0.1:8083;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

3. Включить конфиг и перезагрузить nginx:

```bash
sudo ln -sf /etc/nginx/sites-available/api-gophkeeper2.ampleev.com /etc/nginx/sites-enabled/
sudo nginx -t
sudo systemctl reload nginx
```

---

## 7. Self-hosted runner для репо gophkeeper2

1. В GitHub: репозиторий **gophkeeper2** → Settings → Actions → Runners → New self-hosted runner → Linux, x64.
2. На сервере:

```bash
sudo mkdir -p /opt/actions-runner-gophkeeper2
cd /opt/actions-runner-gophkeeper2
# Скачать и распаковать runner (ссылка и токен из GitHub)
sudo ./config.sh --url https://github.com/<ваш-аккаунт>/gophkeeper2 --token <TOKEN>
sudo ./svc.sh install
sudo ./svc.sh start
```

3. Права: пользователь раннера должен иметь доступ на запись в `/opt/gophkeeper2/gophkeeper2` и право выполнять `sudo systemctl restart gophkeeper2` и `sudo systemctl restart gophkeeper2-test` (при необходимости — правило в sudoers).

---

## 8. Секреты GitHub (Settings → Secrets and variables → Actions)

Добавить в репо **gophkeeper2**:

| Secret             | Описание |
|--------------------|----------|
| DATABASE_URL       | Строка подключения к БД прод, например `postgresql://user:pass@localhost:5432/gophkeeper2?sslmode=disable` |
| DATABASE_URL_TEST  | Строка подключения к БД тест, например `postgresql://user:pass@localhost:5432/gophkeeper2_test?sslmode=disable` |
| SECRET_KEY         | Секретный ключ для JWT (произвольная строка достаточной длины) |

В workflow при деплое из них собираются файлы `.env` и `.env.test` на сервере.

---

## 9. Проверка после первого деплоя

- На сервере:
  - `cd /opt/gophkeeper2/gophkeeper2 && git log -1`
  - `systemctl status gophkeeper2`
  - `curl -s -o /dev/null -w "%{http_code}" http://localhost:8083/health` → 200
- Снаружи:
  - `curl -s -o /dev/null -w "%{http_code}" https://api-gophkeeper2.ampleev.com/health` → 200

В iOS-приложении для прода указать base URL: `https://api-gophkeeper2.ampleev.com`.
