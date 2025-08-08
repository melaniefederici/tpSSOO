package main

import (
	"log"
	"net/http"
	"os"

	"github.com/sisoputnfrba/tp-golang/io/globals"
	"github.com/sisoputnfrba/tp-golang/io/utilsio"
	"github.com/sisoputnfrba/tp-golang/utils/inicio"
	"github.com/sisoputnfrba/tp-golang/utils/mensajes"
)

func main() {

	// Chekpoint 1: Configuracion de logs, carga de configuracion y envio de mensaje al Kernel
	inicio.ConfigurarLogger("io.log")
	globals.Config = utilsio.CargarProximaConfig("configs/")

	mensajes.EnviarMensaje(globals.Config.IPKernel, globals.Config.PuertoKernel, "Hola soy el IO")

	// Checkpoint 2:
	//Leer por argumento el nombre del dispositivo
	if len(os.Args) < 2 {
		log.Fatal("No se encontró el nombre del IO")
	}
	var nombre string = os.Args[1] // nombre pasado por argumento

	//Handshake con Kernel
	utilsio.HandshakeKernel(nombre)

	//Inicio señales
	utilsio.IniciarSeniales(globals.Config.Nombre)

	//Peticiones del Kernel
	//Inicio servidor http
	inicio.IniciarServidor(globals.Config.PuertoIO, map[string]http.HandlerFunc{
		"/io": utilsio.ManejarPeticionIO,
	})
}
