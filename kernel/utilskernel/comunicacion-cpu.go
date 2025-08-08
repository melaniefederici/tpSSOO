package utilskernel

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/sisoputnfrba/tp-golang/kernel/globals"
	"github.com/sisoputnfrba/tp-golang/utils/mensajes"
)

func EnviarProcesoACpu(ip string, puerto int, procesoAEjecutar mensajes.ProcesoAEjecutar) {
	body, err := json.Marshal(procesoAEjecutar)
	if err != nil {
		log.Printf("Error codificando proceso: %s", err.Error())
	}

	url := fmt.Sprintf("http://%s:%d/solicitudProcesoAEjecutar", ip, puerto)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		log.Printf("Error enviando proceso a CPU: %s", err.Error())
	}
	defer resp.Body.Close()

	var respuesta struct {
		OK bool `json:"ok"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&respuesta); err != nil {
		log.Printf("Error al decodificar respuesta de CPU: %s", err.Error())
	}

	//log.Printf("Respuesta de CPU: %v", respuesta.OK)
}

func RegistrarCPU(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var cpu *globals.CPU
	err := decoder.Decode(&cpu)
	if err != nil {
		log.Printf("error al decodificar mensaje: %s\n", err.Error())
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("error al decodificar mensaje"))
		return
	}

	cpu.Ocupado = false
	cpu.ProcesoEjecutando = nil

	globals.CPUDisponibles <- cpu

	globals.CPUsConectadas = append(globals.CPUsConectadas, cpu)

	log.Printf("Se registró la CPU con identificador: %d", cpu.Identificador)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}

func RecibirResultadoDeCPU(w http.ResponseWriter, r *http.Request) {
	var resultado mensajes.ResultadoCPU
	err := json.NewDecoder(r.Body).Decode(&resultado)
	if err != nil {
		log.Printf("Error al decodificar resultado CPU: %s", err.Error())
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	//	log.Printf("CPU devolvió PID: %d | PC: %d | Motivo: %s", resultado.PID, resultado.PC, resultado.Motivo)

	go ManejarResultadoCPU(resultado)

	w.WriteHeader(http.StatusOK)
}

func ManejarResultadoCPU(resultado mensajes.ResultadoCPU) {
	proceso := BuscarPCBPorPID(resultado.PID)
	if proceso == nil {
		log.Printf("No se encontró proceso con PID %d", resultado.PID)
		return
	}
	proceso.PC = resultado.PC

	switch globals.Config.PlanificacionCP {
	case "FIFO":
	case "SJF", "SRT":
		if resultado.Motivo == mensajes.MotivoIO || resultado.Motivo == mensajes.MotivoExit {
			log.Printf("Estimacion anterior para PID (%d) - : %f", proceso.PID, proceso.UltimaEstimacion)
			tiempoEjecutado := time.Now().Sub(proceso.InicioExec)
			ejecutadoMs := float64(tiempoEjecutado.Milliseconds())
			log.Printf("Tiempo ejecutado para PID (%d) - : %f", proceso.PID, ejecutadoMs)
			proceso.EstimacionRafaga = float64(ejecutadoMs)*globals.Config.Alpha + float64(proceso.UltimaEstimacion)*(1-globals.Config.Alpha)
			log.Printf("Rafaga estimada para PID (%d) - : %f", proceso.PID, proceso.EstimacionRafaga)
		}
	}

	switch resultado.Motivo {
	case mensajes.MotivoIO:
		//log.Printf("Proceso %d pasó a BLOCK por IO", resultado.PID) // segun logs obligatorios: ## (<PID>) - Bloqueado por IO: <DISPOSITIVO_IO>
		//args : [0] NOMBRE [1] TIEMPO porque asi lo manda la CPU
		log.Printf("## (%d) - Solicitó syscall: IO", resultado.PID)

		tipoSolicitado := resultado.Args[0]
		var dispositivoLibre *globals.IOConectado

		encontrado := false

		for _, dispositivo := range globals.DispositivosIO {
			if dispositivo.Tipo == tipoSolicitado {
				encontrado = true
				if dispositivo.EnUso == nil && dispositivoLibre == nil {
					dispositivoLibre = dispositivo
				}
			}
		}

		// Si no hay dispositivos del tipo solicitado, finaliza el proceso
		if !encontrado {
			log.Printf("## (%d) - Se solicita IO tipo %s pero no hay dispositivos conectados de ese tipo. Finalizando proceso.", proceso.PID, tipoSolicitado)
			FinalizarProceso(proceso)
			break
		}

		MoverACola(proceso, globals.ColaBLOCK)
		proceso.MutexProc.Lock()

		if proceso.TimerSuspension == nil {
			proceso.TimerSuspension = time.AfterFunc(time.Duration(globals.Config.TiempoSuspension)*time.Millisecond, func() {
				PlanificarMP(proceso.PID)
			})
		}
		proceso.MutexProc.Unlock()

		tiempo, _ := strconv.Atoi(resultado.Args[1])
		proceso.TiempoIO = tiempo

		if dispositivoLibre != nil {
			// Hay dispositivo libre, se lo asignás directo
			dispositivoLibre.EnUso = proceso
			go ocuparYMandarIO(dispositivoLibre, proceso)
			log.Printf("## (%d) - Bloqueado por IO: %s (%s)", proceso.PID, tipoSolicitado, dispositivoLibre.Nombre)
		} else {
			// Todos ocupados, lo mandás a la cola global del tipo
			globals.ColaDiscoIO.MutexCola.Lock()
			globals.ColaDiscoIO.Cola = append(globals.ColaDiscoIO.Cola, proceso)
			globals.ColaDiscoIO.MutexCola.Unlock()

			log.Printf("## (%d) - En espera por IO tipo %s (todos ocupados)", proceso.PID, tipoSolicitado)
		}

	/*	for _, dispositivo := range globals.DispositivosIO {
			if dispositivo.Tipo == tipoSolicitado {
				//dispositivo.Cola.MutexCola.Lock()
				if dispositivo.EnUso == nil{
					dispositivoLibre = dispositivo
					//dispositivo.Cola.MutexCola.Unlock()
					break
				}
				//dispositivo.Cola.MutexCola.Unlock()
			}
		}

		nombre := dispositivoLibre.Nombre
		log.Printf("## (%d) - Solicitó syscall: IO", resultado.PID)
		dispositivo, ok := globals.DispositivosIO[nombre]

		if !ok {
			log.Printf("Dispositivo IO no encontrado: %s", resultado.Args[0])
			FinalizarProceso(proceso)
			break
		}

		MoverACola(proceso, globals.ColaBLOCK)
		log.Printf("## (%d) - Bloqueado por IO: %s", proceso.PID, resultado.Args[0])
		go PlanificarMP(proceso.PID)

		dispositivo.Cola.MutexCola.Lock()
		dispositivo.Cola.Cola = append(dispositivo.Cola.Cola, proceso)
		dispositivo.Cola.MutexCola.Unlock()

		tiempo, _ := strconv.Atoi(resultado.Args[1])
		proceso.TiempoIO = tiempo
		/*if !dispositivo.Ocupado {
			ocuparYMandarIO(dispositivo, proceso)
		}

		dispositivo.Cola.MutexCola.Lock()

		if dispositivo.EnUso == nil {
			proceso := dispositivo.Cola.Cola[0]
			dispositivo.Cola.Cola = dispositivo.Cola.Cola[1:]
			dispositivo.EnUso = proceso
			go ocuparYMandarIO(dispositivo, proceso)
		}

		dispositivo.Cola.MutexCola.Unlock()*/

	case mensajes.MotivoExit:
		//	log.Printf("## (%d) - Finaliza el proceso", resultado.PID)
		log.Printf("## (%d) - Solicitó syscall: EXIT", resultado.PID)
		FinalizarProceso(proceso)

	case mensajes.MotivoDumpMemory:
		log.Printf("## (%d) - Solicitó syscall: DUMP_MEMORY", resultado.PID)
		MoverACola(proceso, globals.ColaBLOCK)
		RealizarDumpDeMemoria(proceso)

	case mensajes.MotivoDesalojo:
		log.Printf("## (%d) - Desalojado por algoritmo SJF/SRT", resultado.PID)
		proceso.InterrupcionEnviada = false
		globals.DesalojoHecho <- struct{}{}
		MoverACola(proceso, globals.ColaREADY)

	default:
		log.Printf("Motivo desconocido recibido de CPU: %s", resultado.Motivo)
	}
	cpu := BuscarCPUPorID(globals.CPUsConectadas, resultado.IdentificadorCPU)
	cpu.Ocupado = false
	//cpu.ProcesoEjecutando = nil
	globals.CPUDisponibles <- cpu
}

func ManejarInitProc(w http.ResponseWriter, r *http.Request) {
	var datos []string
	err := json.NewDecoder(r.Body).Decode(&datos)
	if err != nil {
		log.Printf("Error al decodificar datos INITPROC: %v", err)
		http.Error(w, "JSON inválido", http.StatusBadRequest)
		return
	}
	log.Printf("## (%s) - Solicitó syscall: INIT_PROC", datos[2])
	nombreArchivo := datos[0]
	tamanio, _ := strconv.Atoi(datos[1])

	CrearPCB(nombreArchivo, tamanio)

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}

func BuscarCPUPorID(listaCPUs []*globals.CPU, id int) *globals.CPU {
	for i := range listaCPUs {
		if listaCPUs[i].Identificador == id {
			return listaCPUs[i]
		}
	}
	return nil
}

func FinalizarProceso(proceso *globals.PCB) {
	MoverACola(proceso, globals.ColaEXIT)
	if SolicitudFinalizarProceso(globals.Config.IpMemoria, globals.Config.PuertoMemoria, proceso.PID) {
		EliminarPCBPorPID(proceso.PID)
		log.Print("Voy a mandar la señal")
		globals.SenialMemoriaLiberada <- struct{}{}
		log.Print("Senial de meoria liberada enviada")
	}
	log.Printf("## (%d) - Finaliza el proceso", proceso.PID)
	log.Printf("## (%d) - Métricas de estado: NEW (%d) (%d ms), READY (%d) (%d ms), EXEC (%d) (%d ms), BLOCKED (%d) (%d ms), SUSP_BLOCK (%d) (%d ms), SUSP_READY (%d) (%d ms), EXIT (%d) (%d ms)",
		proceso.PID,
		proceso.ME[globals.EstadoNew], proceso.MT[globals.EstadoNew],
		proceso.ME[globals.EstadoReady], proceso.MT[globals.EstadoReady],
		proceso.ME[globals.EstadoExec], proceso.MT[globals.EstadoExec],
		proceso.ME[globals.EstadoBlocked], proceso.MT[globals.EstadoBlocked],
		proceso.ME[globals.EstadoSuspBlock], proceso.MT[globals.EstadoSuspBlock],
		proceso.ME[globals.EstadoSuspReady], proceso.MT[globals.EstadoSuspReady],
		proceso.ME[globals.EstadoExit], proceso.MT[globals.EstadoExit],
	)

}
