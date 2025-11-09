package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"sync"

	"github.com/inkeliz/gowebview"
)

func main() {
	// === Flags ===
	host := flag.String("host", "localhost", "Host para o servidor HTTP")
	port := flag.String("port", "41234", "Porta para o servidor HTTP")
	dir := flag.String("dir", ".", "Diretório para servir arquivos")
	width := flag.Int("width", 800, "Largura da janela")
	height := flag.Int("height", 400, "Altura da janela")
	title := flag.String("title", "Servidor Local", "Título da janela")
	debug := flag.Bool("debug", true, "Ativar modo debug")
	help := flag.Bool("help", false, "Mostrar ajuda")
	flag.Parse()

	if *help {
		flag.Usage()
		os.Exit(0)
	}

	if *width <= 0 || *height <= 0 {
		fmt.Fprintf(os.Stderr, "Erro: width e height devem ser maiores que zero.\n")
		os.Exit(1)
	}

	// === Salvar PID ===
	pid := os.Getpid()
	execPath, err := os.Executable()
	if err != nil {
		log.Fatalf("Erro ao obter caminho do executável: %v", err)
	}
	execDir := filepath.Dir(execPath)
	pidFilePath := filepath.Join(execDir, "server.prcID")

	err = os.WriteFile(pidFilePath, []byte(fmt.Sprint(pid)), 0644)
	if err != nil {
		log.Fatalf("Erro ao salvar PID no arquivo: %v", err)
	}
	defer os.Remove(pidFilePath) // Remove ao encerrar

	// === Configurar servidor HTTP ===
	address := fmt.Sprintf("%s:%s", *host, *port)
	url := fmt.Sprintf("http://%s", address)

	var wg sync.WaitGroup
	var server *http.Server
	shutdown := make(chan struct{})

	// Handler customizado
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/stop_server.html" {
			fmt.Fprintln(w, "Servidor será encerrado...")
			go func() {
				log.Println("Recebido pedido para parar o servidor via /stop_server.html")
				close(shutdown)
			}()
			return
		}
		// Serve arquivos do diretório
		http.FileServer(http.Dir(*dir)).ServeHTTP(w, r)
	})

	server = &http.Server{
		Addr:    address,
		Handler: mux,
	}

	// Iniciar servidor em goroutine
	wg.Add(1)
	go func() {
		defer wg.Done()
		log.Printf("Servidor HTTP rodando em %s", url)
		log.Printf("Acesse /stop_server.html para encerrar remotamente.")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Erro no servidor: %v", err)
		}
	}()

	// === Aguardar servidor estar pronto ===
	// (Opcional: pode adicionar um health check, mas para simplicidade, damos um pequeno delay)
	// Ou melhor: tentamos conectar via loop, mas aqui vamos direto

	// === Criar WebView ===
	w, err := gowebview.New(&gowebview.Config{
		URL:   url,
		Debug: *debug,
		WindowConfig: &gowebview.WindowConfig{
			Title: *title,
			Size: &gowebview.Point{
				X: int64(*width),
				Y: int64(*height),
			},
		},
	})
	if err != nil {
		log.Fatalf("Erro ao criar janela WebView: %v", err)
	}
	defer w.Destroy()

	// === Canal para encerramento ===
	quit := make(chan struct{})

	// Captura fechamento da janela
	go func() {
		w.Run()
		log.Println("Janela WebView fechada pelo usuário.")
		close(quit)
	}()

	// Captura sinais do sistema (Ctrl+C)
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		<-c
		log.Println("Sinal de interrupção recebido (Ctrl+C).")
		close(quit)
	}()

	// === Aguardar encerramento (janela fechada, sinal, ou /stop_server.html) ===
	select {
	case <-quit:
		log.Println("Encerrando por fechamento da janela ou sinal...")
	case <-shutdown:
		log.Println("Encerrando por solicitação via /stop_server.html...")
	}

	// === Encerrar servidor graciosamente ===
	log.Println("Encerrando servidor HTTP...")
	if err := server.Close(); err != nil {
		log.Printf("Erro ao fechar servidor: %v", err)
	}

	// Aguardar finalização
	wg.Wait()
	log.Println("Servidor encerrado com sucesso. PID file removido.")
}
