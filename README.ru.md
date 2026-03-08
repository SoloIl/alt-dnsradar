# Alt DNSRadar

[English version](./README.md)

Alt DNSRadar — это лёгкая кросс-платформенная CLI-утилита на Go, которая находит альтернативные IP-адреса для домена с помощью EDNS Client Subnet (ECS) probing и ранжирует доступные результаты по реальной TCP-задержке соединения.

Инструмент помогает выявлять IP-адреса, которые могут быть не видны через локальный DNS-резолвер.

Многие современные платформы возвращают разные DNS-ответы в зависимости от географического положения клиента. Обычный запрос к локальному DNS-резолверу показывает лишь небольшую часть доступных endpoint. Имитируя DNS-запросы из множества разных клиентских сетей с помощью ECS, DNSRadar может находить дополнительные IP-адреса, которые обычно скрыты от локального резолвера.

Это делает инструмент полезным для понимания того, как платформа распределяет свою инфраструктуру, как разные резолверы видят один и тот же домен, и какие доступные endpoint отвечают быстрее всего из текущей сети.

## Возможности

- DNS-диагностика через:
  - Local DNS
  - Google DNS over UDP
  - Google DNS over HTTPS
  - Cloudflare DNS over HTTPS
- ECS-сканирование по 540 публичным подсетям
- Обнаружение дополнительных IP-адресов, которые DNS возвращает для разных client network locations
- Измерение TCP-задержки с:
  - 3 проверками на IP
  - медианным значением как итоговым результатом
- TLS-handshake как отдельная диагностика, не участвующая в ранжировании
- Обогащение топ-результатов метаданными Geo / ASN через `ipinfo.io`
- Кросс-платформенная поддержка:
  - Linux
  - macOS
  - Windows
- Минимальный набор зависимостей

## Как это работает

1. Программа запускает DNS-диагностику и сравнивает ответы от Local DNS, Google UDP, Google DoH и Cloudflare DoH.
2. Затем она выполняет ECS-сканирование по 540 публичным подсетям и собирает все уникальные IP, возвращённые DNS.
3. Для каждого найденного IP измеряется TCP-задержка тремя пробами, после чего используется медианное значение.
4. Самые быстрые endpoint обогащаются метаданными из `ipinfo.io`.
5. На выходе программа показывает компактную таблицу с самыми быстрыми доступными endpoint.

## Основные ограничения дизайна

- Единственная метрика ранжирования — TCP connect latency.
- TLS используется только для диагностики и никогда не участвует в ранжировании.
- Геоданные запрашиваются только для верхних результатов.
- Утилита должна оставаться лёгкой и переносимой.

## Готовые пакеты

Готовые prebuilt packages планируются для основных платформ.

Пока здесь оставлены dummy-ссылки под будущую release page:

- Linux amd64: `https://github.com/yourname/dnsradar/releases/latest/download/dnsradar-linux-amd64.tar.gz`
- Linux arm64: `https://github.com/yourname/dnsradar/releases/latest/download/dnsradar-linux-arm64.tar.gz`
- macOS amd64: `https://github.com/yourname/dnsradar/releases/latest/download/dnsradar-darwin-amd64.tar.gz`
- macOS arm64: `https://github.com/yourname/dnsradar/releases/latest/download/dnsradar-darwin-arm64.tar.gz`
- Windows amd64: `https://github.com/yourname/dnsradar/releases/latest/download/dnsradar-windows-amd64.zip`
- Windows arm64: `https://github.com/yourname/dnsradar/releases/latest/download/dnsradar-windows-arm64.zip`

## Установка из исходников

### Что нужно заранее

На системе должен быть установлен Go.

Проверь это командой:

```bash
go version
```

### Linux

1. Установите Go через пакетный менеджер вашего дистрибутива или с официального сайта Go.
2. Клонируйте репозиторий:

```bash
git clone https://github.com/yourname/dnsradar.git
cd dnsradar
```

3. Соберите бинарник:

```bash
go build -o dnsradar .
```

4. Запустите его:

```bash
./dnsradar example.com
```

### macOS

1. Установите Go с официального сайта или через Homebrew:

```bash
brew install go
```

2. Клонируйте репозиторий:

```bash
git clone https://github.com/yourname/dnsradar.git
cd dnsradar
```

3. Соберите бинарник:

```bash
go build -o dnsradar .
```

4. Запустите его:

```bash
./dnsradar example.com
```

### Windows

1. Установите Go через официальный Windows installer.
2. Откройте PowerShell или Command Prompt.
3. Клонируйте репозиторий:

```powershell
git clone https://github.com/yourname/dnsradar.git
cd dnsradar
```

4. Соберите бинарник:

```powershell
go build -o dnsradar.exe .
```

5. Запустите его:

```powershell
.\dnsradar.exe example.com
```

## Быстрый старт

Если вы уже скачали готовый пакет:

### Linux / macOS

1. Распакуйте архив.
2. Откройте terminal в этой папке.
3. Запустите:

```bash
./dnsradar example.com
```

### Windows

1. Распакуйте архив.
2. Откройте PowerShell в этой папке.
3. Запустите:

```powershell
.\dnsradar.exe example.com
```

## Использование

### Показать help

```bash
go run . --help
```

### Русский UI

```bash
go run . example.com --lang ru
```

### Обычный запуск

```bash
dnsradar example.com
```

При таком запуске утилита:

- выполнит DNS-диагностику через Local DNS, Google UDP, Google DoH и Cloudflare DoH
- выполнит начальную TCP/TLS-диагностику для найденных initial endpoint
- будет использовать 3-секундный TCP timeout и 3 TCP-проверки с медианным значением
- выполнит ECS-сканирование по 540 публичным подсетям
- измерит задержку в 20 потоков
- выполнит TLS-диагностику и заполнит метаданные `ipinfo.io` для 5 самых быстрых endpoint

### Примеры

Показать все найденные IP:

```bash
dnsradar example.com --all
```

Записать компактный лог-файл:

```bash
dnsradar example.com -l dnsradar.log
```

Отключить цвета:

```bash
dnsradar example.com --no-color
```

## Пример вывода

Иллюстративный вывод для `youtube.com`:

```text
Alt DNSRadar v0.12

Обработка URL "youtube.com"

DNS-диагностика
----------------------------
Запуск начальной TCP/TLS-диагностики для 5 уникальных endpoint(s)...

Начальная диагностика endpoint

SOURCE            IP               TCP       TLS      NOTE
--------------------------------------------------------------------------------
Local DNS         142.251.38.110   23ms      TIMEOUT  совпадает с DoH-эталоном
Google UDP        142.251.38.110   23ms      TIMEOUT  совпадает с DoH-эталоном
Google DoH        173.194.222.136  19ms      TIMEOUT  эталон
Google DoH        173.194.222.190  22ms      TIMEOUT  эталон
Google DoH        173.194.222.91   23ms      TIMEOUT  эталон
Google DoH        173.194.222.93   26ms      TIMEOUT  эталон
Cloudflare DoH    142.251.38.110   23ms      TIMEOUT  эталон

Сводка DNS-диагностики
- Local DNS вернул другой multi-endpoint набор по сравнению с Google DoH
- Google UDP и Google DoH вернули разные multi-endpoint наборы; возможны кеш, вариативность CDN или перехват
- Cloudflare DoH и Google DoH вернули разные multi-endpoint наборы; уверенность в эталоне ниже

-------------------------------------------

Запуск ECS-сканирования
Всего ECS-подсетей: 540

ECS scan 540/540 [================================================| 100 %]

Успешных DNS-ответов: 540
Уникальных IP найдено: 402

TCP latency 402/402 [================================================| 100 %]

Подготовка таблицы лучших endpoint для youtube.com (geo lookup + TLS diagnostics)...

Самые быстрые endpoint для youtube.com

IP               TCP     TLS      CDN            ASN      LOCATION
--------------------------------------------------------------------------------
192.178.25.14    22ms    TIMEOUT  Google         AS15169  US   Mountain View
142.251.38.110   22ms    TIMEOUT  Google         AS15169  US   Mountain View
142.251.142.238  22ms    TIMEOUT  Google         AS15169  SE   Stockholm
172.217.19.238   23ms    TIMEOUT  Google         AS15169  SE   Stockholm
172.217.20.174   23ms    TIMEOUT  Google         AS15169  US   Mountain View
```

Реальный вывод зависит от домена, сетевых условий, поведения резолверов и доступности endpoint.

## Примечания

- Запросы метаданных к `ipinfo.io` ограничены правилами внешнего сервиса. Сейчас утилита запрашивает данные только для 5 самых быстрых endpoint, чтобы уменьшить расход лимита (1000 запросов в день).
- Для multi-endpoint доменов DNS-ответы разных резолверов могут отличаться без признаков вмешательства.
- Некоторые сети или резолверы могут ограничивать ECS-поведение или возвращать неполные результаты.

## Ограничения

- ECS-сканирование использует грубую сетку и может не обнаружить все возможные endpoint.
- Точность метаданных зависит от базы `ipinfo.io`.
- TLS-диагностика может зависеть от middlebox или DPI-систем и должна интерпретироваться как диагностика, а не как метрика ранжирования.

## Тесты

Запуск unit tests:

```bash
go test ./...
```

## Лицензия

MIT
