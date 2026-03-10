# Настройка сервера и CI/CD для нового бэкенда (как pointscounter)

Документ для разработчика **второго** бэкенда на том же сервере: как устроен текущий проект (pointscounter), как запускаются тесты при коммитах, как настроен автоматический деплой на тест и прод, и пошаговая инструкция, чтобы повторить эту схему для нового приложения.

---

## Содержание

1. [Как устроен текущий проект (pointscounter)](#1-как-устроен-текущий-проект-pointscounter)
2. [CI/CD: тесты и деплой](#2-cicd-тесты-и-деплой)
3. [Сервер: каталоги, сервисы, nginx, БД](#3-сервер-каталоги-сервисы-nginx-бд)
4. [Пошаговая настройка нового проекта](#4-пошаговая-настройка-нового-проекта)
5. [Что уточнить перед стартом](#5-что-уточнить-перед-стартом)

---

## 1. Как устроен текущий проект (pointscounter)

### Репозиторий и ветки

- Репозиторий: `cheepcounter-backend` (Go).
- Рабочая ветка: `main`. Деплой в прод и тест происходит при **push в main**.
- Тесты в CI запускаются на **pull request в main** (без деплоя).

### Что где лежит на сервере

| Назначение | Путь |
|------------|------|
| Код прод/тест | `/opt/pointscounter/pointscounter-backend` |
| Self-hosted GitHub Actions runner | `/opt/actions-runner` |
| Nginx конфиги | `/etc/nginx/sites-available/`, симлинки в `sites-enabled/` |
| Systemd юниты | `/etc/systemd/system/` |
| SSL (Let's Encrypt) | `/etc/letsencrypt/live/<домен>/` |

### Порты (pointscounter)

| Окружение | HTTP API | Метрики Prometheus |
|-----------|----------|--------------------|
| Прод | 8081 | 2112 |
| Тест | 8082 | 2113 |

### Сервисы systemd

- **pointscounter** — прод: бинарь `pointscounter`, порт 8081, `.env` (или переменные в юните).
- **pointscounter-test** — тест: тот же бинарь, `EnvironmentFile=.env.test`, порт 8082.

Оба используют один и тот же каталог `/opt/pointscounter/pointscounter-backend` и разные файлы окружения.

### Базы данных (PostgreSQL на том же сервере)

- Прод: база `cheepcounter` (или имя из `DATABASE_URL`).
- Тест: база `cheepcounter_test` (из `DATABASE_URL_TEST`).

Миграции применяются при старте приложения к той БД, на которую указывает `DATABASE_URL` в соответствующем `.env` / `.env.test`.

### Nginx

- Домен API: `api-pointscounter.ampleev.com`.
- Конфиг: `/etc/nginx/sites-available/api-pointscounter.ampleev.com`.
- Проксирование: HTTP/HTTPS → `localhost:8081`, WebSocket `/ws` → `localhost:8081/ws`.
- SSL: Let's Encrypt (certbot).

Тестовый бэкенд (8082) может быть вынесен на отдельный поддомен (например `api-test-pointscounter.ampleev.com`) или оставлен только по порту — см. `docs/TEST_BACKEND_SETUP.md`.

---

## 2. CI/CD: тесты и деплой

### Тесты (GitHub Actions)

- **Файл:** `.github/workflows/test.yml`
- **Триггер:** `pull_request` в ветку `main`.
- **Где выполняется:** `ubuntu-latest` (хосты GitHub).
- **Что делается:** checkout → Go 1.22 → Postgres 15 (service container) → `go mod tidy` → миграции (`scripts/run_migrations.go`) → `go vet` и `go test`.
- **Важно:** тесты **не** запускаются при push в main — при push в main стартует только деплой. Чтобы тесты были «при коммитах», они срабатывают на PR в main; мерж в main после успешного PR запускает деплой.

Если нужно запускать тесты ещё и при push в main (например, блокировать деплой при падении тестов), можно добавить вызов того же job или отдельный workflow `on: push: branches: [main]` с тестами и `needs` в деплое.

### Деплой (GitHub Actions, self-hosted)

- **Файл:** `.github/workflows/deploy-selfhosted.yml`
- **Триггер:** `push` в `main` и `workflow_dispatch`.
- **Где выполняется:** **self-hosted** runner на том же сервере (`runs-on: [self-hosted, linux]`).
- **Что делается:**
  - `cd` в `PROJECT_DIR` (например `/opt/pointscounter/pointscounter-backend`), `git fetch` + `git reset --hard origin/main`
  - `go build` бинарей (основной сервер + утилиты при необходимости)
  - Сборка `.env` и `.env.test` из секретов репозитория (DATABASE_URL, DATABASE_URL_TEST, APNS_*, DEEPSEEK_API_KEY, ADMIN_API_KEY и т.д.)
  - Опционально: backfill, синхронизация таблиц (например `app_copy` прод → тест)
  - `sudo systemctl restart pointscounter` и `pointscounter-test`
  - Health-check: `curl -fsS http://localhost:8081/health`

Секреты хранятся в **GitHub → Settings → Secrets and variables → Actions** и подставляются в `.env`/`.env.test` на сервере (в workflow), не хранятся в репо.

Подробнее про установку self-hosted runner: **`docs/SELF_HOSTED_RUNNER.md`**.

---

## 3. Сервер: каталоги, сервисы, nginx, БД

### Общая схема

```
Сервер (один физический/виртуальный хост)
├── /opt/pointscounter/pointscounter-backend   # код + бинари pointscounter
├── /opt/actions-runner                        # GitHub Actions runner (для этого репо)
├── PostgreSQL (systemd: postgresql / postgres)
│   ├── БД cheepcounter      → прод
│   └── БД cheepcounter_test → тест
├── systemd
│   ├── pointscounter       → прод, порт 8081
│   └── pointscounter-test  → тест, порт 8082
└── nginx
    └── api-pointscounter.ampleev.com → proxy to localhost:8081
```

### Nginx (пример для pointscounter)

- Файл: `/etc/nginx/sites-available/api-pointscounter.ampleev.com`
- Включение: `sudo ln -sf .../sites-available/api-pointscounter... sites-enabled/`, затем `nginx -t` и `systemctl reload nginx`.
- В конфиге: редирект 80→443, server 443 с ssl, proxy_pass на `http://localhost:8081` для API и на `http://localhost:8081/ws` для WebSocket с таймаутами для long-lived соединений.

Полный пример конфига nginx и systemd для pointscounter приведён в **`MIGRATION_GUIDE.md`** (шаги 4 и 6).

### Проверка после деплоя

- Прод: `curl -s -o /dev/null -w "%{http_code}" https://api-pointscounter.ampleev.com/health` → 200.
- Локально на сервере: `curl -s http://localhost:8081/health`, `systemctl status pointscounter`, `journalctl -u pointscounter -n 50`.

При проблемах с подключением из приложения см. **`docs/CONNECTION_ERROR_TROUBLESHOOTING.md`**.

---

## 4. Пошаговая настройка нового проекта

Ниже — чеклист для **второго** бэкенда на том же сервере. Подставь свои имена вместо плейсхолдеров (например `NEWAPP` → короткое имя приложения, `api-newapp.ampleev.com` → домен API).

### 4.1 Решение по репозиторию и раннеру

- **Вариант A:** новый бэкенд в **отдельном репозитории**. Тогда либо:
  - зарегистрировать **второй self-hosted runner** на сервере для этого репо (второй каталог, например `/opt/actions-runner-newapp`), либо
  - использовать **один раннер**, привязанный к нескольким репозиториям (если так разрешено в организации), и в каждом репо свой workflow с разными `PROJECT_DIR`.
- **Вариант B:** новый бэкенд в **том же репо** (монорепо). Тогда один workflow может деплоить оба бэкенда (два блока шагов или два job) с разными `PROJECT_DIR` и разными systemd-сервисами.

Далее предполагается **отдельный репо** и отдельные каталог/сервисы/порты.

### 4.2 Подготовка репозитория нового проекта

1. Создать репо (например `newapp-backend`), клонировать на сервер в `/opt/newapp/newapp-backend` (или аналог).
2. Добавить в репо:
   - **Тесты:** `.github/workflows/test.yml` по образцу pointscounter: `on: pull_request: branches: [main]`, Postgres service, Go, миграции, `go vet` и `go test`.
   - **Деплой:** `.github/workflows/deploy-selfhosted.yml` по образцу pointscounter, с подстановкой:
     - `PROJECT_DIR=/opt/newapp/newapp-backend`
     - имена бинарей и команд (`go build -o newapp ./cmd/server` и т.д.)
     - формирование `.env` и `.env.test` из секретов **этого** репо (DATABASE_URL, DATABASE_URL_TEST и остальные переменные нового приложения)
     - рестарт сервисов: `sudo systemctl restart newapp` и `newapp-test` (если нужен тест)
     - health-check на портах нового приложения (например 8083 для прод, 8084 для тест).

### 4.3 Порты для второго приложения

Чтобы не пересекаться с pointscounter:

| Окружение | HTTP API | Метрики (если нужны) |
|-----------|----------|----------------------|
| Прод newapp | 8083 | 2114 |
| Тест newapp | 8084 | 2115 |

В `.env` / `.env.test` или в systemd задать `PORT=8083` и `PORT=8084` соответственно.

### 4.4 Базы данных

На том же PostgreSQL:

```bash
sudo -u postgres psql -c "CREATE DATABASE newapp;"
sudo -u postgres psql -c "CREATE DATABASE newapp_test;"
# Выдать права пользователю из DATABASE_URL:
sudo -u postgres psql -c "GRANT ALL PRIVILEGES ON DATABASE newapp TO your_db_user;"
sudo -u postgres psql -c "GRANT ALL PRIVILEGES ON DATABASE newapp_test TO your_db_user;"
```

В GitHub Secrets репо нового приложения добавить:

- `DATABASE_URL` — подключение к `newapp`
- `DATABASE_URL_TEST` — подключение к `newapp_test`

### 4.5 Systemd

Создать два юнита по аналогии с pointscounter.

**Прод** — `/etc/systemd/system/newapp.service`:

```ini
[Unit]
Description=NewApp Backend API (production)
After=network.target postgresql.service

[Service]
Type=simple
User=root
WorkingDirectory=/opt/newapp/newapp-backend
EnvironmentFile=/opt/newapp/newapp-backend/.env
ExecStart=/opt/newapp/newapp-backend/newapp
Restart=always
RestartSec=5
Environment=PORT=8083

[Install]
WantedBy=multi-user.target
```

**Тест** — `/etc/systemd/system/newapp-test.service`:

```ini
[Unit]
Description=NewApp Backend (TEST - port 8084)
After=network.target postgresql.service

[Service]
Type=simple
User=root
WorkingDirectory=/opt/newapp/newapp-backend
EnvironmentFile=/opt/newapp/newapp-backend/.env.test
ExecStart=/opt/newapp/newapp-backend/newapp
Restart=on-failure
RestartSec=5

[Install]
WantedBy=multi-user.target
```

В `.env.test` задать `PORT=8084` (и при необходимости `METRICS_PORT=2115`).

```bash
sudo systemctl daemon-reload
sudo systemctl enable newapp newapp-test
sudo systemctl start newapp
sudo systemctl start newapp-test
```

### 4.6 Nginx и домен

1. Домен: например `api-newapp.ampleev.com`. В DNS (Cloudflare и т.п.) создать A-запись на IP сервера.
2. SSL: `sudo certbot certonly --nginx -d api-newapp.ampleev.com` (или `--standalone` при первом разе).
3. Конфиг: `/etc/nginx/sites-available/api-newapp.ampleev.com` — по образцу pointscounter, заменив:
   - `server_name` на `api-newapp.ampleev.com`
   - `proxy_pass http://localhost:8081` на `http://localhost:8083`
   - пути к сертификатам на `/etc/letsencrypt/live/api-newapp.ampleev.com/...`
4. Включить: `sudo ln -sf /etc/nginx/sites-available/api-newapp.ampleev.com /etc/nginx/sites-enabled/`, `sudo nginx -t`, `sudo systemctl reload nginx`.

### 4.7 Self-hosted runner для репо нового приложения

Если репо отдельный — на сервере завести второй раннер (чтобы не смешивать с pointscounter):

1. В GitHub: **newapp-backend** → Settings → Actions → Runners → New self-hosted runner → Linux, x64.
2. На сервере: новый каталог, например `sudo mkdir -p /opt/actions-runner-newapp`, скачать и распаковать runner, выполнить `./config.sh --url https://github.com/<org>/newapp-backend --token <TOKEN>`.
3. Запуск как сервис: `sudo ./svc.sh install`, `sudo ./svc.sh start`.
4. Права: пользователь, под которым крутится раннер, должен иметь доступ на запись в `/opt/newapp/newapp-backend` и право выполнять `sudo systemctl restart newapp` (и при необходимости `newapp-test`) — при необходимости настроить sudoers.

Если решаешь использовать **один** раннер для двух репо — нужно привязать один runner к обоим репозиториям (если твой GitHub это позволяет) и в каждом репо в workflow указывать свой `PROJECT_DIR` и свои команды рестарта.

### 4.8 Секреты нового репо

В **Settings → Secrets and variables → Actions** репо нового приложения добавить все переменные, которые деплой подставляет в `.env`/`.env.test` (DATABASE_URL, DATABASE_URL_TEST, ключи API и т.д.). Имена могут совпадать с pointscounter (например DATABASE_URL), но значения — свои (другие БД, при необходимости другие ключи).

### 4.9 Проверка

- После первого деплоя: на сервере `git log -1` в `/opt/newapp/newapp-backend`, `systemctl status newapp`, `curl -s http://localhost:8083/health`.
- Снаружи: `curl -s -o /dev/null -w "%{http_code}" https://api-newapp.ampleev.com/health` → 200.
- В приложении указать base URL `https://api-newapp.ampleev.com` (прод) и при необходимости тестовый URL (поддомен или порт по аналогии с `docs/TEST_BACKEND_SETUP.md`).

---

## 5. Что уточнить перед стартом

Перед тем как повторять схему для нового проекта, полезно зафиксировать:

1. **Имя и репо:** короткое имя сервиса (например `newapp`), полное имя репо и организация/владелец (для URL раннера и секретов).
2. **Один сервер или два:** оба бэкенда на одном сервере (как в этом документе) или новый бэкенд на отдельном — от этого зависит количество раннеров и nginx.
3. **Один раннер или два:** один self-hosted runner на два репо (и общий `PROJECT_DIR` только в смысле «в каждом workflow свой путь») или два каталога раннеров и два привязки к разным репо.
4. **Тесты при push в main:** оставить только тесты на PR или добавить запуск тестов при push в main и блокировать деплой при падении (например через `needs` в deploy job).
5. **Домены:** точные имена доменов для прод (и при необходимости тест) API, чтобы сразу заложить их в nginx и в приложение.
6. **Специфичные бинари и шаги:** помимо основного сервера нужны ли отдельные утилиты (как у pointscounter: blockchain-config, nickname-manager, backfill и т.д.) и скрипты после деплоя (синхронизация таблиц, обновление Grafana и т.п.) — их нужно отразить в своём `deploy-selfhosted.yml`.

После ответов на эти пункты можно один раз пройти раздел 4 по шагам и получить второй бэкенд с тестами на PR и автоматическим деплоем на тест и прод при push в main.

---

## Ссылки на документы (в этом репо)

| Документ | О чём |
|----------|--------|
| `docs/SELF_HOSTED_RUNNER.md` | Установка и настройка self-hosted GitHub Actions runner на сервере |
| `docs/TEST_BACKEND_SETUP.md` | Тестовый бэкенд (порт 8082, БД, systemd, синхронизация app_copy) |
| `docs/CONNECTION_ERROR_TROUBLESHOOTING.md` | Что проверять при ошибках подключения в приложении |
| `MIGRATION_GUIDE.md` | Пример полного переезда домена: nginx, systemd, certbot, проверки |
| `docs/HIGH_LOAD_DIAGNOSIS.md` | Диагностика высокой нагрузки на сервере |
