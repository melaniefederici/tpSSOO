package utilskernel

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/sisoputnfrba/tp-golang/kernel/globals"
)

func ocuparYMandarIO(io *globals.IOConectado, proceso *globals.PCB) {
	//io.Ocupado = true
	io.EnUso = proceso
	peticionIO := struct {
		PID    int `json:"pid"`
		Tiempo int `json:"tiempo"`
	}{
		PID:    proceso.PID,
		Tiempo: proceso.TiempoIO,
	}

	body, err := json.Marshal(peticionIO)
	if err != nil {
		log.Printf("Error codificando solicitud de IO: %s", err.Error())
		return
	}

	url := fmt.Sprintf("http://%s:%d/io", io.IP, io.Puerto)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(body))

	if err != nil {
		log.Printf("Error enviando solicitud de IO: %s", err.Error())
		return
	}

	defer resp.Body.Close()

	//log.Printf("Se ha enviado IO PID: %d %d", proceso.PID, io.Puerto)

}

func RegistrarIO(w http.ResponseWriter, r *http.Request) {
	var ioRecibido struct {
		Nombre string `json:"nombre"`
		IP     string `json:"ip"`
		Puerto int    `json:"puerto"`
		Tipo   string `json:"tipo"`
	}

	err := json.NewDecoder(r.Body).Decode(&ioRecibido)
	if err != nil {
		http.Error(w, "Error al decodificar IO", http.StatusBadRequest)
		return
	}

	globals.DispositivosIO[ioRecibido.Nombre] = &globals.IOConectado{
		IP:     ioRecibido.IP,
		Puerto: ioRecibido.Puerto,
		Cola:   globals.ColaDiscoIO,
		// Ocupado: false
		EnUso:  nil,
		Nombre: ioRecibido.Nombre,
		Tipo:   ioRecibido.Tipo,
	}

	log.Printf("Dispositivo IO registrado: %s en %s: %d", ioRecibido.Nombre, ioRecibido.IP, ioRecibido.Puerto)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))

}

/*func DesbloquearProcesoDeIO(pid int) {
	var dispositivoEncontrado *globals.IOConectado
	var proceso *globals.PCB

	/*for nombre, dispositivo := range globals.DispositivosIO {
		dispositivo.Cola.MutexCola.Lock()
		for i, pcb := range dispositivo.Cola.Cola {
			if pcb.PID == pid {
				proceso = pcb
				dispositivo.Cola.Cola = append(dispositivo.Cola.Cola[:i], dispositivo.Cola.Cola[i+1:]...)
				dispositivoEncontrado = dispositivo
				log.Printf("## (%d) - Desbloqueado de IO: %s", pid, nombre)
				break
			}
		}
		dispositivo.Cola.MutexCola.Unlock()


		if proceso != nil {
			break
		}
	}
	dispositivo := BuscarDispositivoPorPID(pid)
	dispositivo.Cola.MutexCola.Lock()

	if dispositivo.EnUso != nil && dispositivo.EnUso.PID == pid {
		dispositivo.EnUso = nil

		if len(dispositivo.Cola.Cola) > 0 {
			proximo := dispositivo.Cola.Cola[0]
			dispositivo.Cola.Cola = dispositivo.Cola.Cola[1:]
			dispositivo.EnUso = proximo
			go ocuparYMandarIO(dispositivo, proximo)
		}
	} else {
		log.Printf("PID %d no es el proceso que está usando el IO actualmente", pid)
	}

	dispositivo.Cola.MutexCola.Unlock()

	/*if proceso == nil || dispositivoEncontrado == nil {
		log.Printf("No se encontró el proceso %d en ninguna cola de IO", pid)
		return
	}

	// 1. Mover el proceso a READY
	MoverACola(proximo, globals.ColaREADY)

	// 2. Liberar el dispositivo
	//dispositivoEncontrado.Ocupado = false

	//IRIA EN OTRA FUNCION
	// 3.Si hay otro proceso esperando, despacharlo
	dispositivoEncontrado.Cola.MutexCola.Lock()

	if len(dispositivoEncontrado.Cola.Cola) > 0 {
		siguiente := dispositivoEncontrado.Cola.Cola[0]
		dispositivoEncontrado.Cola.Cola = dispositivoEncontrado.Cola.Cola[1:]

		log.Printf("## (%d) - Se asigna a IO liberado", siguiente.PID)
		ocuparYMandarIO(dispositivoEncontrado, siguiente)

	}
	dispositivoEncontrado.Cola.MutexCola.Unlock()

}*/

func DesbloquearProcesoDeIO(pid int) {
	dispositivo := BuscarDispositivoPorPID(pid)
	if dispositivo == nil {
		log.Printf("No se encontró dispositivo usando el PID %d", pid)
		return
	}

	dispositivo.Cola.MutexCola.Lock()

	if dispositivo.EnUso != nil && dispositivo.EnUso.PID == pid {
		// Liberar el dispositivo
		procesoDesbloqueado := dispositivo.EnUso
		dispositivo.EnUso = nil

		procesoDesbloqueado.MutexProc.Lock()
		if procesoDesbloqueado.TimerSuspension != nil {
			procesoDesbloqueado.TimerSuspension.Stop()
			procesoDesbloqueado.TimerSuspension = nil
		}
		procesoDesbloqueado.MutexProc.Unlock()

		if procesoDesbloqueado.Estado == globals.EstadoSuspBlock {
			MoverACola(procesoDesbloqueado, globals.ColaSuspReady)
			log.Printf("## (%d) finalizó IO y pasa a SUSP READY", procesoDesbloqueado.PID)
		} else {
			MoverACola(procesoDesbloqueado, globals.ColaREADY)
			log.Printf("## (%d) finalizó IO y pasa a READY", procesoDesbloqueado.PID)

		}

		// Si hay otro proceso esperando, asignarlo al dispositivo
		if len(dispositivo.Cola.Cola) > 0 {
			siguiente := dispositivo.Cola.Cola[0]
			dispositivo.Cola.Cola = dispositivo.Cola.Cola[1:]
			dispositivo.EnUso = siguiente
			log.Printf("## (%d) - Bloqueado por IO: %s (%s)", siguiente.PID, dispositivo.Tipo, dispositivo.Nombre)

			go ocuparYMandarIO(dispositivo, siguiente)
		}

	} else {
		log.Printf("PID %d no es el proceso que está usando el IO actualmente", pid)
	}

	dispositivo.Cola.MutexCola.Unlock()
}

func BuscarDispositivoPorPID(pid int) *globals.IOConectado {
	for _, dispositivo := range globals.DispositivosIO {
		dispositivo.Cola.MutexCola.Lock()
		if dispositivo.EnUso != nil && dispositivo.EnUso.PID == pid {
			dispositivo.Cola.MutexCola.Unlock()
			return dispositivo
		}
		dispositivo.Cola.MutexCola.Unlock()
	}
	return nil
}

func RecibirFinIO(w http.ResponseWriter, r *http.Request) {
	var datos struct {
		PID int `json:"pid"`
	}

	err := json.NewDecoder(r.Body).Decode(&datos)
	if err != nil {
		http.Error(w, "Error al decodificar JSON", http.StatusBadRequest)
		log.Printf("Error al decodificar JSON: %s", err.Error())
		return
	}
	DesbloquearProcesoDeIO(datos.PID)

}

func ManejarDesconexionIO(w http.ResponseWriter, r *http.Request) {
	var datos struct {
		Nombre string `json:"nombre"`
	}

	err := json.NewDecoder(r.Body).Decode(&datos)
	if err != nil {
		http.Error(w, "Error al decodificar desconexión IO", http.StatusBadRequest)
		return
	}

	dispositivo, existe := globals.DispositivosIO[datos.Nombre]
	if !existe {
		log.Printf("No se encontró el dispositivo IO: %s", datos.Nombre)
		w.WriteHeader(http.StatusNotFound)
		return
	}

	log.Printf("## Desconexión de IO '%s' recibida. Finalizando procesos asociados...", datos.Nombre)

	if dispositivo.EnUso != nil {
		FinalizarProceso(dispositivo.EnUso)
	}

	globals.ColaDiscoIO.MutexCola.Lock()
	for _, proceso := range globals.ColaDiscoIO.Cola {
		FinalizarProceso(proceso)
	}
	globals.ColaDiscoIO.Cola = nil
	globals.ColaDiscoIO.MutexCola.Unlock()

	delete(globals.DispositivosIO, datos.Nombre)

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}
