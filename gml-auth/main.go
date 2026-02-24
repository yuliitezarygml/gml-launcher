package main

import (
	"fmt"
	"gml-auth/config"
	"gml-auth/handlers"
	"gml-auth/news"
	"gml-auth/storage"
	"log"
	"net"
	"net/http"
	"time"
)

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		lw := &responseWriter{ResponseWriter: w, code: http.StatusOK}
		next.ServeHTTP(lw, r)
		log.Printf("[%s] %s %s → %d (%s)",
			r.RemoteAddr, r.Method, r.URL.Path, lw.code, time.Since(start))
	})
}

type responseWriter struct {
	http.ResponseWriter
	code int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.code = code
	rw.ResponseWriter.WriteHeader(code)
}

func localIPs() []string {
	var ips []string
	ifaces, _ := net.Interfaces()
	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}
		addrs, _ := iface.Addrs()
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			if ip != nil && ip.To4() != nil {
				ips = append(ips, ip.String())
			}
		}
	}
	return ips
}

func buildCache(provider news.Provider, interval time.Duration) *news.Cache {
	cache := news.NewCache(provider, interval)
	cache.Start()
	return cache
}

func registerNews(mux *http.ServeMux, path string, cache *news.Cache) {
	if cache != nil {
		mux.HandleFunc(path, handlers.NewNewsHandler(cache).List)
	} else {
		mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte("[]\n"))
		})
	}
}

func main() {
	const port = "5003"

	store := storage.New("data/users.json")
	authHandler := handlers.NewAuthHandler(store)
	adminHandler := handlers.NewAdminHandler(store)

	mux := http.NewServeMux()
	mux.HandleFunc("/", authHandler.SignIn)
	mux.HandleFunc("/api/v1/integrations/auth/signin", authHandler.SignIn)
	mux.HandleFunc("/api/v1/users/refresh", authHandler.Refresh)
	mux.Handle("/admin/users", adminHandler)
	mux.Handle("/admin/users/", adminHandler)

	// Новости
	cfg, err := config.Load("config.json")
	if err != nil {
		log.Printf("[news] config.json не найден, /api/news/* вернёт []")
	}
	interval := time.Duration(cfg.News.RefreshSeconds) * time.Second

	var tgCache, dcCache *news.Cache
	if cfg.News.Telegram.Token != "" {
		tgCache = buildCache(news.NewTelegramProvider(cfg.News.Telegram.Token, cfg.News.Telegram.Channel), interval)
		log.Printf("[news] Telegram: %s → /api/news/telegram", cfg.News.Telegram.Channel)
	}
	if cfg.News.Discord.Token != "" {
		dcCache = buildCache(news.NewDiscordProvider(cfg.News.Discord.Token, cfg.News.Discord.Channel), interval)
		log.Printf("[news] Discord: %s → /api/news/discord", cfg.News.Discord.Channel)
	}

	registerNews(mux, "/api/news/telegram", tgCache)
	registerNews(mux, "/api/news/discord", dcCache)

	// /api/news — объединённая лента (оба источника)
	switch {
	case tgCache != nil && dcCache != nil:
		combined := buildCache(news.NewMultiProvider(
			news.NewTelegramProvider(cfg.News.Telegram.Token, cfg.News.Telegram.Channel),
			news.NewDiscordProvider(cfg.News.Discord.Token, cfg.News.Discord.Channel),
		), interval)
		mux.HandleFunc("/api/news", handlers.NewNewsHandler(combined).List)
	case tgCache != nil:
		mux.HandleFunc("/api/news", handlers.NewNewsHandler(tgCache).List)
	case dcCache != nil:
		mux.HandleFunc("/api/news", handlers.NewNewsHandler(dcCache).List)
	default:
		mux.HandleFunc("/api/news", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte("[]\n"))
		})
	}

	fmt.Println("===========================================")
	fmt.Println("  GML Auth Server")
	fmt.Println("===========================================")
	fmt.Printf("  Порт : %s\n", port)
	for _, ip := range localIPs() {
		fmt.Printf("  Адрес: http://%s:%s\n", ip, port)
	}
	fmt.Printf("  Адрес: http://127.0.0.1:%s  (localhost)\n", port)
	fmt.Println("===========================================")
	log.Printf("Сервер запущен, ожидаю подключения на :%s", port)

	log.Fatal(http.ListenAndServe(":"+port, loggingMiddleware(mux)))
}
