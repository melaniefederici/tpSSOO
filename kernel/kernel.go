package main

import (
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/sisoputnfrba/tp-golang/kernel/globals"
	"github.com/sisoputnfrba/tp-golang/kernel/planificadores"
	"github.com/sisoputnfrba/tp-golang/kernel/utilskernel"
	"github.com/sisoputnfrba/tp-golang/utils/inicio"
	"github.com/sisoputnfrba/tp-golang/utils/mensajes"
)

func main() {
	// test
	inicio.ConfigurarLogger("kernel.log")

	if len(os.Args) < 3 {
		log.Fatalf("Faltan argumentos: ./kernel [archivo_pseudocodigo] [tamanio_proceso]")
	}

	globals.Config = inicio.CargarConfiguracion[globals.KernelConfig]("kernel-config.json")

	var nombreArchivo string = os.Args[1]
	tamanio, _ := strconv.Atoi(os.Args[2])
	globals.CPUDisponibles = make(chan *globals.CPU, 10)

	globals.ColaDiscoIO = &globals.ColaDisco{
		Cola: make([]*globals.PCB, 0),
	}

	// hilo para atender solicitudes
	go func() {
		handlers := map[string]http.HandlerFunc{
			"/mensaje":          mensajes.RecibirMensaje,
			"/paquetes":         mensajes.RecibirPaquetes,
			"/finIO":            utilskernel.RecibirFinIO, // para 'desbloquear' un proceso de IO
			"/registrarCPU":     utilskernel.RegistrarCPU,
			"/recibirResultado": utilskernel.RecibirResultadoDeCPU,
			"/registrarIO":      utilskernel.RegistrarIO,
			"/manejarINITPROC":  utilskernel.ManejarInitProc,
			"/desconexionIO":    utilskernel.ManejarDesconexionIO,

			//"/solicitudProcesoAEjecutar": RecibirProcesoAEjecutar,
			// "/interrupt"				: ManejarInterrupcion, endpoint de cpu..

			// BRYAN - Explico resumidamente la modificacion en paquetes:
			// LA funcion original de RecibirPaquetes, solo 'loguea'
			// La funcion de HandlerPaquetesKernel, 'loguea' e interpreta contenido

			// Me sirve para recibir solicitudes http, por ej la de IO y asi poder seguir operando (en este caso,
			// para desbloquear un proceso con IO terminado y que debe volver a la cola de ready...)
		}
		inicio.IniciarServidor(globals.Config.PuertoKernel, handlers)
	}()

	go func() {
		planificadores.PlanificarLP()
	}()

	go func() {
		planificadores.PlanificarCP()
	}()

	utilskernel.CrearPCB(nombreArchivo, tamanio)

	mensajes.EnviarMensaje(globals.Config.IpMemoria, globals.Config.PuertoMemoria, "Hola, soy el Kernel")

	select {}
}
