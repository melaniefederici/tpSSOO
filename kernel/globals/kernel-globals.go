package globals

import (
	"sync"
	"time"
)

type EstadoProceso string

const (
	EstadoNew       EstadoProceso = "NEW"
	EstadoReady     EstadoProceso = "READY"
	EstadoExec      EstadoProceso = "EXEC"
	EstadoBlocked   EstadoProceso = "BLOCKED"
	EstadoExit      EstadoProceso = "EXIT"
	EstadoSuspBlock EstadoProceso = "SUSP_BLOCK"
	EstadoSuspReady EstadoProceso = "SUSP_READY"
)

type PCB struct {
	PID                 int
	PC                  int
	Estado              EstadoProceso
	Archivo             string
	Tamanio             int
	ME                  map[EstadoProceso]int //Metrica de estado, mapea estado con veces que estuvo
	MT                  map[EstadoProceso]int //Metrica de tiempo x estado, mapea estado con tiempo que estuvo
	EstimacionRafaga    float64
	UltimaEstimacion    float64
	RafagaReal          int
	InicioEstado        time.Time
	TiempoIO            int
	InicioExec          time.Time
	MutexProc           sync.Mutex
	TimerSuspension     *time.Timer
	InterrupcionEnviada bool
}
type KernelConfig struct {
	IpMemoria         string  `json:"ip_memoria"`
	PuertoMemoria     int     `json:"puerto_memoria"`
	PuertoKernel      int     `json:"puerto_kernel"`
	PlanificacionCP   string  `json:"scheduler_algorithm"`
	PlanificacionLP   string  `json:"ready_ingress_algorithm"`
	Alpha             float64 `json:"alpha"`
	EstimacionInicial float64 `json:"initial_estimate"`
	TiempoSuspension  int     `json:"suspension_time"`
	LogLevel          string  `json:"log_level"`
}

type CPU struct {
	IP                string `json:"ip_cpu"`
	PuertoDispatch    int    `json:"puerto_dispatch"`
	PuertoInterrupt   int    `json:"puerto_interrupt"`
	Identificador     int    `json:"identificador_cpu"`
	Ocupado           bool   `json:"cpu_ocupado"`
	ProcesoEjecutando *PCB   `json:"proceso_ejecutando"`
}

var CPUDisponibles chan *CPU

// puedo mandar 10 señales sin que se bloquee.
var SenialMemoriaLiberada chan struct{} = make(chan struct{}, 100)

type Cola struct {
	Cola      []*PCB
	MutexCola sync.Mutex
	Estado    EstadoProceso
}

//

type IOConectado struct {
	IP     string
	Puerto int
	Cola   *ColaDisco
	//Ocupado bool
	EnUso *PCB
	// PID   int
	Nombre string
	Tipo   string
}

var ColaDiscoIO *ColaDisco

type ColaDisco struct {
	Cola      []*PCB
	MutexCola sync.Mutex
}

var DispositivosIO = make(map[string]*IOConectado)

//

var (
	ColaNEW       = &Cola{Estado: EstadoNew}
	ColaREADY     = &Cola{Estado: EstadoReady}
	ColaEXEC      = &Cola{Estado: EstadoExec}
	ColaBLOCK     = &Cola{Estado: EstadoBlocked}
	ColaEXIT      = &Cola{Estado: EstadoExit}
	ColaSuspBlock = &Cola{Estado: EstadoSuspBlock}
	ColaSuspReady = &Cola{Estado: EstadoSuspReady}
)

var Config *KernelConfig

var PIDActual int

var CPUsConectadas []*CPU

var HayProcesoEnReady = make(chan struct{}, 10)

var DesalojoHecho = make(chan struct{}, 10)
