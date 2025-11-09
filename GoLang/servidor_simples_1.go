package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

func main() {
	host := flag.String("host", "localhost", "Host para o servidor HTTP")
	port := flag.String("port", "41234", "Porta para o servidor HTTP")
	dir := flag.String("dir", ".", "Diretório para servir arquivos")

	flag.Parse()

	pid := os.Getpid()
	execPath, err := os.Executable()
    if err != nil {
        panic(err)
    }
    execDir := filepath.Dir(execPath)	
	pidFilePath := filepath.Join(execDir, "server.prcID")
	err = os.WriteFile(pidFilePath, []byte(fmt.Sprint(pid)), 0644)
	if err != nil {
		log.Fatalf("Erro ao salvar PID no arquivo: %v", err)
	}

	address := fmt.Sprintf("%s:%s", *host, *port)

	server := &http.Server{Addr: address}

	// Faz o handler customizado para checar se tentam acessar "stop_server.html"
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/stop_server.html" {
			fmt.Fprintln(w, "Servidor será encerrado...")
			go func() {
				log.Println("Recebido pedido para parar o servidor...")
				server.Close()
				os.Exit(0)
			}()
			return
		}
		// Serve os arquivos normalmente
		http.FileServer(http.Dir(*dir)).ServeHTTP(w, r)
	})

	log.Printf("Servidor iniciado em http://%s. PID salvo em %s\n", address, pidFilePath)
	err = server.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		log.Fatalf("Erro no servidor: %v", err)
	}
}
