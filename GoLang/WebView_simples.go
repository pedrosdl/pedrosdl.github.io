package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/inkeliz/gowebview"
)

func main() {
	// === Configuração dos argumentos via linha de comando ===
	var (
		largura = flag.Int("largura", 800, "Largura da janela")
		altura  = flag.Int("altura", 400, "Altura da janela")
		url     = flag.String("url", "https://google.com", "URL a ser carregada")
		titulo  = flag.String("titulo", "Hello World", "Título da janela")
		debug   = flag.Bool("debug", true, "Ativar modo debug")
		ajuda   = flag.Bool("help", false, "Mostrar ajuda")
	)

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Uso: %s [opções]\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Opções:\n")
		flag.PrintDefaults()
	}

	flag.Parse()

	if *ajuda {
		flag.Usage()
		os.Exit(0)
	}

	// Validação básica
	if *largura <= 0 || *altura <= 0 {
		fmt.Fprintf(os.Stderr, "Erro: largura e altura devem ser maiores que zero.\n")
		os.Exit(1)
	}

	// === Criação da janela com os parâmetros fornecidos ===
	w, err := gowebview.New(&gowebview.Config{
		URL:   *url,
		Debug: *debug,
		WindowConfig: &gowebview.WindowConfig{
			Title: *titulo,
			Size: &gowebview.Point{
				X: int64(*largura), // Conversão de int → int64
				Y: int64(*altura),  // Conversão de int → int64
			},
		},
	})
	if err != nil {
		panic(err)
	}

	defer w.Destroy()
	w.Run()
}
