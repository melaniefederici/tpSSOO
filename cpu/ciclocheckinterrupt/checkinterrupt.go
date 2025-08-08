package ciclocheckinterrupt

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/sisoputnfrba/tp-golang/cpu/globals"
)

type IntMsg struct {
	PID int `json:"pid"`
	PC  int `json:"pc"`
}

func CheckInterrupt(pid int) (bool, int) {
	addr := formatAddr(globals.Config.IPKernel, globals.Config.PuertoKernel)

	conn, err := net.DialTimeout("tcp", addr, 100*time.Millisecond)
	if err != nil {
		return false, 0
	}
	defer conn.Close()

	conn.SetReadDeadline(time.Now().Add(50 * time.Millisecond))

	var msg IntMsg
	if err := json.NewDecoder(conn).Decode(&msg); err != nil {
		log.Printf("Error al decodificar el mensaje de interrupción: %v", err)

		buf := make([]byte, 1024)
		n, _ := conn.Read(buf)
		log.Printf("Contenido recibido antes del error: %s", string(buf[:n]))

		return false, 0
	}

	if msg.PID == pid {
		return true, msg.PC
	}
	return false, 0
}

func formatAddr(ip string, port int) string {
	if strings.Contains(ip, ":") {

		return fmt.Sprintf("[%s]:%d", ip, port)
	}

	return fmt.Sprintf("%s:%d", ip, port)
}

type Interrupcion struct {
	PID int `json:"pid"`
}

func RecibirInterrupcion(w http.ResponseWriter, r *http.Request) {
	var pidAInterrumpir Interrupcion
	err := json.NewDecoder(r.Body).Decode(&pidAInterrumpir)
	if err != nil {
		log.Printf("Error al decodificar el PID a interrumpir: %s", err.Error())
		w.WriteHeader(http.StatusBadRequest)
	}
	log.Printf("######## HAY INTERRUPCION PARA EL PID: %d", pidAInterrumpir.PID)
	globals.HayInterrupcion = true

	w.WriteHeader(http.StatusOK)
}
