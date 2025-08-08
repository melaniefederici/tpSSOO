package inicio

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

func ConfigurarLogger(nombre string) {
	logFile, err := os.OpenFile(nombre, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0666)
	if err != nil {
		panic(err)
	}
	mw := io.MultiWriter(os.Stdout, logFile)
	log.SetOutput(mw)
}

func CargarConfiguracion[T any](filePath string) *T {
	var config T

	configFile, err := os.Open(filePath)
	if err != nil {
		log.Fatalf("Error al abrir el archivo de configuración (%s): %v", filePath, err)
	}
	defer configFile.Close()

	jsonParser := json.NewDecoder(configFile)
	if err := jsonParser.Decode(&config); err != nil {
		log.Fatalf("Error al parsear el archivo JSON: %v", err)
	}

	log.Printf("Se cargó correctamente la configuración desde %s", filePath)

	return &config
}

func IniciarServidor(puerto int, handlers map[string]http.HandlerFunc) {
	mux := http.NewServeMux()

	for path, handler := range handlers {
		mux.HandleFunc(path, handler)
	}

	direccion := fmt.Sprintf(":%d", puerto)
	err := http.ListenAndServe(direccion, mux)
	if err != nil {
		log.Fatalf("Error al iniciar el servidor: %v", err)
	}
	log.Printf("Iniciando servidor en puerto: %d", puerto)

}
