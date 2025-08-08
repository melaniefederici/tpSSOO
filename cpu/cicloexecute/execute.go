package cicloexecute

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/sisoputnfrba/tp-golang/cpu/cache"
	"github.com/sisoputnfrba/tp-golang/cpu/globals"
	"github.com/sisoputnfrba/tp-golang/cpu/utilscpu"
	"github.com/sisoputnfrba/tp-golang/utils/mensajes"
)

func ExecuteInstruccion(instruccion string, parametros []string, pid int, pc int) int {

	inst := strings.ToUpper(instruccion) //por si llega en minus
	log.Printf("## EXECUTE - Instrucción: %s | Parámetros: %v", instruccion, parametros)

	nuevoPC := pc + 1

	switch inst {
	case "NOOP":
		log.Println("Instrucción NOOP: No hace nada.")

	case "WRITE":
		if !utilscpu.ValidarParametros("WRITE", 2, parametros) {
			log.Fatal("Parametros invalidos")
		}

		dirFisica, _ := strconv.Atoi(parametros[0])
		dato := parametros[1]

		solicitudEscritura := globals.PeticionEscritura{
			PID:       pid,
			DirFisica: dirFisica,
			Cadena:    dato,
		}
		utilscpu.SolicitudEscritura(globals.Config.IPMemoria, globals.Config.PuertoMemoria, solicitudEscritura)

		tamPagina := globals.Config.TamanioPagina
		nroPagina := dirFisica / tamPagina
		desplazamiento := dirFisica % tamPagina

		// Busco en cache o pido memoria
		contenido, hit := cache.LeerPagina(pid, nroPagina)
		if !hit {
			contenido = utilscpu.PedirPaginaAMemoria(pid, dirFisica)
			//cache.AgregarEntrada(pid, nroPagina, contenido, false) //AGREGAR ESTO REY
			log.Printf("Recibi contenido de memoria: %v", contenido)
		}

		// Copiar sólo la cantidad exacta de bytes del dato en el desplazamiento
		datoBytes := []byte(dato)
		copy(contenido[desplazamiento:], datoBytes)

		cache.EscribirPagina(pid, nroPagina, contenido)
		//contenidoEscritoEnCache, _ := cache.LeerPagina(pid, nroPagina)
		log.Printf("Contenido escrito en cache fue: %v", contenido)

		//copy(contenido[desplazamiento:], []byte(dato))

		log.Printf("PID: %d - Acción: WRITE - Dirección Física: %d - Valor: %s", pid, dirFisica, dato) //---------> LOG MIN Y OBLIGATORIO

		//mensajes.EnviarMensaje(globals.Config.IPMemoria, globals.Config.PuertoMemoria, fmt.Sprintf("WRITE %d %s", dirFisica, dato))

	case "READ":
		if !utilscpu.ValidarParametros("READ", 2, parametros) {
			log.Fatal("Parametros invalidos")
		}

		dirFisica, _ := strconv.Atoi(parametros[0])
		tamano, _ := strconv.Atoi(parametros[1])

		tamPagina := globals.Config.TamanioPagina
		nroPagina := dirFisica / tamPagina
		desplazamiento := dirFisica % tamPagina

		// uso la cacahe
		contenido, hit := cache.LeerPagina(pid, nroPagina)
		if !hit {
			contenido = utilscpu.PedirPaginaAMemoria(pid, dirFisica)
			cache.AgregarEntrada(pid, nroPagina, contenido, false)
		}
		log.Printf("El despplazamiento es de: %d y el tamaño del arg: %d", desplazamiento, tamano)
		valor := contenido[desplazamiento : desplazamiento+tamano]

		log.Printf("PID: %d - Acción: READ - Dirección Física: %d - Valor: %s", pid, dirFisica, valor) //---------> LOG MIN Y OBLIGATORIO

		mensajes.EnviarMensaje(globals.Config.IPMemoria, globals.Config.PuertoMemoria, fmt.Sprintf("READ %d %d", dirFisica, tamano))

	case "GOTO":
		if !utilscpu.ValidarParametros("GOTO", 1, parametros) {
			log.Fatal("Parametros invalidos")
		}

		nuevoPC, err := strconv.Atoi(parametros[0])
		if err != nil {
			log.Printf("Error al convertir GOTO a número: %s", err)
			return pc
		}
		log.Printf("GOTO: Actualizando PC a %d", nuevoPC)
		return nuevoPC

	case "IO":
		if !utilscpu.ValidarParametros("IO", 2, parametros) {
			log.Fatal("Parametros invalidos")
		}

		dispositivo := parametros[0]
		tiempo := parametros[1]
		argumentos := []string{dispositivo, tiempo}
		log.Printf("IO: Dispositivo: %s, Tiempo: %s", dispositivo, tiempo)

		resultado := mensajes.ResultadoCPU{
			PID:              pid,
			PC:               nuevoPC,
			Motivo:           mensajes.MotivoIO,
			Args:             argumentos,
			IdentificadorCPU: globals.Config.Identificador,
		}

		utilscpu.EnviarResultadoCPU(globals.Config.IPKernel, globals.Config.PuertoKernel, resultado)
		return nuevoPC

	case "INIT_PROC":
		if !utilscpu.ValidarParametros("INIT_PROC", 2, parametros) {
			log.Fatal("Parametros invalidos")
		}

		archivo := parametros[0]
		tamano := parametros[1]
		pidAEnviar := strconv.Itoa(pid)

		log.Printf("INIT_PROC: Archivo: %s, Tamaño: %s", archivo, tamano)

		datos := []string{archivo, tamano, pidAEnviar}
		utilscpu.EnviarInitProc(globals.Config.IPKernel, globals.Config.PuertoKernel, datos)

		return nuevoPC

	case "DUMP_MEMORY":
		log.Println("DUMP_MEMORY: Realizando un volcado de memoria.")

		resultado := mensajes.ResultadoCPU{
			PID:              pid,
			PC:               nuevoPC,
			Motivo:           mensajes.MotivoDumpMemory,
			Args:             nil,
			IdentificadorCPU: globals.Config.Identificador,
		}

		utilscpu.EnviarResultadoCPU(globals.Config.IPKernel, globals.Config.PuertoKernel, resultado)

	case "EXIT":
		log.Println("EXIT: Finalizando el proceso.")
		utilscpu.FinalizarProceso(pid)

		//  limpio entradas del proceso en la TLB y en la cache cuando desalojo
		//tlb.LimpiarTLB(pid)
		//cache.LimpiarProceso(pid)

		resultado := mensajes.ResultadoCPU{
			PID:              pid,
			PC:               nuevoPC,
			Motivo:           mensajes.MotivoExit,
			Args:             nil,
			IdentificadorCPU: globals.Config.Identificador,
		}

		utilscpu.EnviarResultadoCPU(globals.Config.IPKernel, globals.Config.PuertoKernel, resultado)

	default:
		log.Printf("Instrucción desconocida: %s", instruccion)
	}

	return nuevoPC
}
