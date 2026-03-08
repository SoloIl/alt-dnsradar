package main

import (
	"fmt"
	"os"
)

func currentLang() string {
	for i := 0; i < len(os.Args); i++ {
		if os.Args[i] == "--lang" && i+1 < len(os.Args) {
			if os.Args[i+1] == "ru" {
				return "ru"
			}
			return "en"
		}
	}

	if flagSettings.Lang != nil {
		switch *flagSettings.Lang {
		case "ru":
			return "ru"
		default:
			return "en"
		}
	}

	return "en"
}

func msgUsageTitle() string {
	if currentLang() == "ru" {
		return "Использование:"
	}
	return "Usage:"
}

func msgExamplesTitle() string {
	if currentLang() == "ru" {
		return "Примеры:"
	}
	return "Examples:"
}

func msgRecommendedExample() string {
	if currentLang() == "ru" {
		return "  dnsradar instagram.com (рекомендуемый вариант, по умолчанию)"
	}
	return "  dnsradar instagram.com (recommended, default)"
}

func msgSecondaryExample() string {
	if currentLang() == "ru" {
		return "  dnsradar youtube.com --all"
	}
	return "  dnsradar youtube.com --all"
}

func msgDefaultBehaviorTitle() string {
	if currentLang() == "ru" {
		return "Что будет сделано по умолчанию:"
	}
	return "Default behavior:"
}

func msgDefaultBehaviorLines() []string {
	if currentLang() == "ru" {
		return []string{
			"будет выполнена DNS-диагностика через Local DNS, Google UDP, Google DoH и Cloudflare DoH",
			"для начальных endpoint будет выполнена TCP/TLS-диагностика",
			"TCP проверяется с таймаутом 3 секунды, тремя тестами (median latency)",
			"будет выполнено ECS-сканирование по 540 публичным подсетям",
			"измерение задержки (latency) выполняется в 20 потоков",
			"для 5 самых быстрых endpoint выполняются TLS-диагностика и заполняются метаданные от ipinfo.io (лимит ipinfo.io: 1000 запросов в день)",
		}
	}

	return []string{
		"runs DNS diagnostics via Local DNS, Google UDP, Google DoH, and Cloudflare DoH",
		"runs initial TCP/TLS diagnostics for the discovered initial endpoints",
		"uses 3-second TCP timeout and 3 probes with median latency",
		"runs ECS scanning across 540 public subnets",
		"measures latency with 20 worker threads",
		"runs TLS diagnostics and fills ipinfo.io metadata for the top 5 fastest endpoints (ipinfo.io limit: 1000 requests per day)",
	}
}

func msgLogFileFlag() string {
	if currentLang() == "ru" {
		return "имя лог-файла"
	}
	return "optional log file name"
}

func msgResolverFlag() string {
	if currentLang() == "ru" {
		return "DoH-резолвер с поддержкой ECS"
	}
	return "DoH resolver with ECS support"
}

func msgTimeoutFlag() string {
	if currentLang() == "ru" {
		return "таймаут в секундах"
	}
	return "network timeout in seconds"
}

func msgShowAllFlag() string {
	if currentLang() == "ru" {
		return "показать все найденные IP"
	}
	return "show all discovered IPs"
}

func msgVerboseFlag() string {
	if currentLang() == "ru" {
		return "подробный вывод"
	}
	return "verbose output"
}

func msgQuietFlag() string {
	if currentLang() == "ru" {
		return "тихий режим консоли"
	}
	return "quiet console"
}

func msgNoColorFlag() string {
	if currentLang() == "ru" {
		return "отключить цвета в терминале"
	}
	return "disable terminal colors"
}

func msgLangFlag() string {
	if currentLang() == "ru" {
		return "язык интерфейса: en или ru"
	}
	return "ui language: en or ru"
}

func msgProgramBanner(name string, version string) string {
	return fmt.Sprintf("%s v%s", name, version)
}

func msgProcessingURL(domain string) string {
	if currentLang() == "ru" {
		return fmt.Sprintf("Обработка URL %q", domain)
	}
	return fmt.Sprintf("Processing URL %q", domain)
}

func msgDNSDiagnosticsTitle() string {
	if currentLang() == "ru" {
		return "DNS-диагностика"
	}
	return "DNS diagnostics"
}

func msgInitialProbeStatus(count int) string {
	if currentLang() == "ru" {
		return fmt.Sprintf("Запуск начальной TCP/TLS-диагностики для %d уникальных endpoint(s)...", count)
	}
	return fmt.Sprintf("Running initial TCP/TLS diagnostics for %d unique endpoint(s)...", count)
}

func msgInitialEndpointDiagnosticsTitle() string {
	if currentLang() == "ru" {
		return "Начальная диагностика endpoint"
	}
	return "Initial endpoint diagnostics"
}

func msgDNSSummaryTitle() string {
	if currentLang() == "ru" {
		return "Сводка DNS-диагностики"
	}
	return "DNS diagnostic summary"
}

func msgNoIPv4Answers() string {
	if currentLang() == "ru" {
		return "нет IPv4-ответов"
	}
	return "no IPv4 answers"
}

func msgReference() string {
	if currentLang() == "ru" {
		return "эталон"
	}
	return "reference"
}

func msgSharedWithDoHReference() string {
	if currentLang() == "ru" {
		return "совпадает с DoH-эталоном"
	}
	return "shared with DoH reference"
}

func msgNotInDoHReference() string {
	if currentLang() == "ru" {
		return "нет в DoH-эталоне"
	}
	return "not in DoH reference"
}

func msgReferenceUnavailable() string {
	if currentLang() == "ru" {
		return "эталон недоступен"
	}
	return "reference unavailable"
}

func msgGoogleUDPAndGoogleDoHUnavailable() string {
	if currentLang() == "ru" {
		return "сравнение Google UDP и Google DoH недоступно"
	}
	return "Google UDP and Google DoH comparison unavailable"
}

func msgGoogleUDPConsistent() string {
	if currentLang() == "ru" {
		return "Google UDP согласуется с Google DoH"
	}
	return "Google UDP is consistent with Google DoH"
}

func msgGoogleUDPPartialDiffers() string {
	if currentLang() == "ru" {
		return "Google UDP частично отличается от Google DoH; возможны кеш или вариативность CDN"
	}
	return "Google UDP partially differs from Google DoH; possible cache or CDN variance"
}

func msgGoogleUDPMultiSetMismatch() string {
	if currentLang() == "ru" {
		return "Google UDP и Google DoH вернули разные multi-endpoint наборы; возможны кеш, вариативность CDN или перехват"
	}
	return "Google UDP and Google DoH returned different multi-endpoint sets; possible cache, CDN variance, or interception"
}

func msgGoogleUDPStrongMismatch() string {
	if currentLang() == "ru" {
		return "Сильное расхождение между Google UDP и Google DoH; возможен DNS-перехват"
	}
	return "Strong mismatch between Google UDP and Google DoH; possible DNS interception"
}

func msgLocalDNSUnavailable() string {
	if currentLang() == "ru" {
		return "сравнение Local DNS и Google DoH недоступно"
	}
	return "Local DNS comparison with Google DoH unavailable"
}

func msgLocalDNSConsistent() string {
	if currentLang() == "ru" {
		return "Local DNS согласуется с Google DoH"
	}
	return "Local DNS is consistent with Google DoH"
}

func msgLocalDNSPartialDiffers() string {
	if currentLang() == "ru" {
		return "Local DNS частично отличается от Google DoH"
	}
	return "Local DNS partially differs from Google DoH"
}

func msgLocalDNSMultiSetMismatch() string {
	if currentLang() == "ru" {
		return "Local DNS вернул другой multi-endpoint набор по сравнению с Google DoH"
	}
	return "Local DNS returned a different multi-endpoint set from Google DoH"
}

func msgLocalDNSStrongMismatch() string {
	if currentLang() == "ru" {
		return "Local DNS сильно отличается от Google DoH"
	}
	return "Local DNS strongly differs from Google DoH"
}

func msgCloudflareDoHUnavailable() string {
	if currentLang() == "ru" {
		return "сравнение Cloudflare DoH и Google DoH недоступно"
	}
	return "Cloudflare DoH and Google DoH comparison unavailable"
}

func msgCloudflareDoHConsistent() string {
	if currentLang() == "ru" {
		return "Cloudflare DoH согласуется с Google DoH"
	}
	return "Cloudflare DoH is consistent with Google DoH"
}

func msgCloudflareDoHPartialDiffers() string {
	if currentLang() == "ru" {
		return "Cloudflare DoH частично отличается от Google DoH; возможны вариативность CDN или кеш"
	}
	return "Cloudflare DoH partially differs from Google DoH; possible CDN or cache variance"
}

func msgCloudflareDoHMultiSetMismatch() string {
	if currentLang() == "ru" {
		return "Cloudflare DoH и Google DoH вернули разные multi-endpoint наборы; уверенность в эталоне ниже"
	}
	return "Cloudflare DoH and Google DoH returned different multi-endpoint sets; reference confidence is lower"
}

func msgCloudflareDoHStrongMismatch() string {
	if currentLang() == "ru" {
		return "Сильное расхождение между Cloudflare DoH и Google DoH; уверенность в эталоне ниже"
	}
	return "Strong mismatch between Cloudflare DoH and Google DoH; reference confidence is lower"
}

func msgErrorPrefix(err error) string {
	if currentLang() == "ru" {
		return fmt.Sprintf("\nОшибка: %v\n", err)
	}
	return fmt.Sprintf("\nError: %v\n", err)
}

func msgErrorClosingLog(err error) string {
	if currentLang() == "ru" {
		return fmt.Sprintf("Ошибка закрытия лога: %v\n", err)
	}
	return fmt.Sprintf("Error closing log: %v\n", err)
}

func msgStartingECSScan() string {
	if currentLang() == "ru" {
		return "Запуск ECS-сканирования"
	}
	return "Starting ECS scan"
}

func msgTotalECSSubnets(count int) string {
	if currentLang() == "ru" {
		return fmt.Sprintf("Всего ECS-подсетей: %d", count)
	}
	return fmt.Sprintf("Total ECS subnets: %d", count)
}

func msgSuccessfulDNSReplies(count int) string {
	if currentLang() == "ru" {
		return fmt.Sprintf("Успешных DNS-ответов: %d", count)
	}
	return fmt.Sprintf("DNS successful replies: %d", count)
}

func msgUniqueIPDiscovered(count int) string {
	if currentLang() == "ru" {
		return fmt.Sprintf("Уникальных IP найдено: %d", count)
	}
	return fmt.Sprintf("Unique IP discovered: %d", count)
}

func msgPOPClustersDiscovered(count int) string {
	if currentLang() == "ru" {
		return fmt.Sprintf("POP-кластеров найдено: %d", count)
	}
	return fmt.Sprintf("POP clusters discovered: %d", count)
}

func msgGeoRateLimitWarning() string {
	if currentLang() == "ru" {
		return "Достигнут лимит geo lookup; поля location могут быть неполными"
	}
	return "Geo lookup rate limit reached; location fields may be incomplete"
}

func msgNoReachableEdges() string {
	if currentLang() == "ru" {
		return "Доступные endpoint не найдены"
	}
	return "No reachable edges found"
}

func msgPreparingTopEndpointTable(domain string) string {
	if currentLang() == "ru" {
		return fmt.Sprintf("Подготовка таблицы лучших endpoint для %s (geo lookup + TLS diagnostics)...", domain)
	}
	return fmt.Sprintf("Preparing top endpoint table for %s (geo lookup + TLS diagnostics)...", domain)
}

func msgTopFastestEndpoints(domain string) string {
	if currentLang() == "ru" {
		return fmt.Sprintf("Самые быстрые endpoint для %s", domain)
	}
	return fmt.Sprintf("Top fastest endpoints for %s", domain)
}

func msgAllDiscoveredIPs() string {
	if currentLang() == "ru" {
		return "Все найденные IP"
	}
	return "All discovered IPs"
}
