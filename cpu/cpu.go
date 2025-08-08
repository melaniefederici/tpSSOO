package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"strconv"
	"time"

	"github.com/sisoputnfrba/tp-golang/cpu/cache"
	"github.com/sisoputnfrba/tp-golang/cpu/ciclocheckinterrupt"
	"github.com/sisoputnfrba/tp-golang/cpu/ciclodecode"
	"github.com/sisoputnfrba/tp-golang/cpu/cicloexecute"
	"github.com/sisoputnfrba/tp-golang/cpu/ciclofetch"
	"github.com/sisoputnfrba/tp-golang/cpu/globals"
	"github.com/sisoputnfrba/tp-golang/cpu/mmu"
	"github.com/sisoputnfrba/tp-golang/cpu/tlb"
	"github.com/sisoputnfrba/tp-golang/cpu/utilscpu"
	"github.com/sisoputnfrba/tp-golang/utils/inicio"
	"github.com/sisoputnfrba/tp-golang/utils/mensajes"
)

func main() {

	inicio.ConfigurarLogger("cpu.log")

	if len(os.Args) < 2 {
		log.Fatal("Faltan argumentos: ./cpu [identificador]")
	}

	identificador_cpu, _ := strconv.Atoi(os.Args[1])
	pathConfig := "cpu" + os.Args[1] + ".json"
	globals.Config = inicio.CargarConfiguracion[globals.CPUConfig](pathConfig)

	utilscpu.ObtenerConfigMemoria()

	tlb.Init(globals.Config.TLBEntries, globals.Config.TLBReplacement)
	cache.Init(globals.Config.CacheEntries, globals.Config.CacheReplacement, globals.Config.CacheDelay)

	go func() {
		dispatchHandlers := map[string]http.HandlerFunc{
			"/solicitudProcesoAEjecutar": utilscpu.RecibirProcesoAEjecutar,
		}
		inicio.IniciarServidor(globals.Config.PuertoCPUDispatch, dispatchHandlers)
	}()

	go func() {
		interruptHandlers := map[string]http.HandlerFunc{
			"/interrupcion": ciclocheckinterrupt.RecibirInterrupcion,
		}
		inicio.IniciarServidor(globals.Config.PuertoCPUInterrupt, interruptHandlers)
	}()

	//mensajes.EnviarMensaje(globals.Config.IPMemoria, globals.Config.PuertoMemoria, "Hola, soy un CPU") //envio mensaje a memoria
	utilscpu.RegistrarEnKernel(globals.Config.IPKernel, globals.Config.PuertoKernel, identificador_cpu)

	for {
		//peticion := utilscpu.EsperarRespuesta()
		peticion := <-globals.CanalProcesoAEjecutar //El kernel le envia mediante handler /solicitudProcesoAEjecutar
		for {                                       //Agrego ciclo otro ciclo for para las instrucciones, sino cada vez que termine de ejecutar 1 instruccion iba a esperar otro proceso
			instruccion, parametros := ciclofetch.FetchInstruccion(peticion.PID, peticion.PC)
			log.Printf("## PID: %d - FETCH - Program Counter: %d", peticion.PID, peticion.PC) //---------> LOG MIN Y OBLIGATORIO

			decoded := ciclodecode.DecodeInstruccion(instruccion, parametros)

			if decoded.NecesitaTraduccion {
				/*	for i, p := range decoded.Parametros {
					num, _ := strconv.Atoi(p)
					decoded.Parametros[i] = fmt.Sprint(mmu.TraducirDireccion(num, peticion.PID))
				}*/
				dirLogica, _ := strconv.Atoi(parametros[0])
				decoded.Parametros[0] = fmt.Sprint(mmu.TraducirDireccion(dirLogica, peticion.PID)) //1er parametro de decoded dir fisica
			}
			log.Printf("## PID: %d - Ejecutando: %s - %v", peticion.PID, decoded.Tipo, decoded.Parametros) // REVISAR ---------> LOG MIN Y OBLIGATORIO

			nuevoPC := cicloexecute.ExecuteInstruccion(decoded.Tipo, decoded.Parametros, peticion.PID, peticion.PC)

			// Comprobar si hay interrupción
			if globals.HayInterrupcion { //inter, pcInt := ciclocheckinterrupt.CheckInterrupt(peticion.PID); inter {
				log.Println("## Llega interrupcion al puerto Interrupt") //---------> LOG MIN Y OBLIGATORIO

				log.Printf("Kernel interrumpio PID: %d ", peticion.PID)
				resultado := mensajes.ResultadoCPU{
					PID:              peticion.PID,
					PC:               nuevoPC,
					Motivo:           mensajes.MotivoDesalojo,
					Args:             nil,
					IdentificadorCPU: globals.Config.Identificador,
				}
				utilscpu.EnviarResultadoCPU(globals.Config.IPKernel, globals.Config.PuertoKernel, resultado)
				globals.HayInterrupcion = false
				break
			}

			peticion.PC = nuevoPC //actualizo PC

			if decoded.Tipo == "EXIT" || decoded.Tipo == "IO" || decoded.Tipo == "DUMP_MEMORY" { //En el caso de alguna syscall bloqueante, se termina el ciclo de ejecucion
				break // y espera a que kernel le envie otro proceso
			}

			time.Sleep(100 * time.Millisecond)
		}
	}
}
