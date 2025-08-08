package utilsio

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/sisoputnfrba/tp-golang/io/globals"
	"github.com/sisoputnfrba/tp-golang/utils/inicio"
)

// Funcion para Handshake con el Kernel
func HandshakeKernel(nombre string) {
	registro := globals.RegistroIO{
		Nombre: globals.Config.Nombre,
		IP:     globals.Config.IPIO,
		Puerto: globals.Config.PuertoIO,
		Tipo:   globals.Config.Tipo,
	}

	// Serializacion, salgo si falla
	body, err := json.Marshal(registro)
	if err != nil {
		os.Exit(1)
	}

	// Envio a Kernel, salgo si falla
	url := fmt.Sprintf("http://%s:%d/registrarIO", globals.Config.IPKernel, globals.Config.PuertoKernel)
	_, err = http.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		os.Exit(1)
	}
}

// Funcion Handler para atender peticiones
func ManejarPeticionIO(w http.ResponseWriter, r *http.Request) {
	var peticion globals.PeticionIO

	//Decodifico json
	err := json.NewDecoder(r.Body).Decode(&peticion)
	if err != nil {
		http.Error(w, "Peticion no reconocida", http.StatusBadRequest)
		return
	}

	//Inicio el usleep
	log.Printf("## PID: %d - Inicio de IO - Tiempo: %d", peticion.PID, peticion.Tiempo)
	time.Sleep(time.Duration(peticion.Tiempo) * time.Millisecond)
	log.Printf("## PID: %d - Fin de IO", peticion.PID)

	//Aviso al Kernel que la IO finalizo
	AvisarFinIO(globals.Config.IPKernel, globals.Config.PuertoKernel, peticion.PID)
	w.WriteHeader(http.StatusOK)
}

func AvisarFinIO(ip string, puerto int, pid int) {
	paquete := struct {
		PID int `json:"pid"`
	}{
		PID: pid,
	}

	body, err := json.Marshal(paquete)
	if err != nil {
		log.Printf("Error codificando mensaje: %s", err.Error())
		return
	}

	url := fmt.Sprintf("http://%s:%d/finIO", ip, puerto)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		log.Printf("Error enviando mensaje a ip:%s puerto:%d - %s", ip, puerto, err.Error())
		return
	}
	defer resp.Body.Close()

	log.Printf("Respuesta del servidor: %s", resp.Status)
}

func NotificarDesconexion(ip string, puerto int, nombre string) {
	payload := struct {
		Nombre string `json:"nombre"`
	}{
		Nombre: nombre,
	}

	body, _ := json.Marshal(payload)
	url := fmt.Sprintf("http://%s:%d/desconexionIO", ip, puerto)
	_, err := http.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		log.Printf("Error notificando desconexión al Kernel: %s", err.Error())
	}
	log.Printf("Kernel fue notificado de la desconexion del dispositivo: %s", nombre)
}

func IniciarSeniales(nombre string) {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-sigs
		log.Printf("Desconectando IO...  %v", sig)
		NotificarDesconexion(globals.Config.IPKernel, globals.Config.PuertoKernel, nombre)
		log.Printf("IO desconectado")
		os.Exit(0)
	}()
}

func CargarProximaConfig(dir string) *globals.IOConfig {
	files, err := os.ReadDir(dir)
	if err != nil {
		log.Fatalf("No se pudo leer el directorio de configuraciones: %v", err)
	}

	for _, file := range files {
		path := filepath.Join(dir, file.Name())
		config := inicio.CargarConfiguracion[globals.IOConfig](path)

		addr := fmt.Sprintf("%s:%d", config.IPIO, config.PuertoIO)
		listener, err := net.Listen("tcp", addr)
		if err == nil {
			listener.Close()
			log.Printf("Usando configuración: %s", path)
			return config
		}
	}

	log.Fatal("No hay configuraciones de IO disponibles con puertos libres")
	return nil
}
