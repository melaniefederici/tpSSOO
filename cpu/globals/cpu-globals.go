package globals

import (
	"sync"

	"github.com/sisoputnfrba/tp-golang/utils/mensajes"
)

type CPUConfig struct {
	PuertoCPUDispatch  int    `json:"port_cpu_dispatch"`
	PuertoCPUInterrupt int    `json:"port_cpu_interrupt"`
	IPCPU              string `json:"ip_cpu"`
	IPMemoria          string `json:"ip_memory"`
	PuertoMemoria      int    `json:"port_memory"`
	IPKernel           string `json:"ip_kernel"`
	PuertoKernel       int    `json:"port_kernel"`
	TLBEntries         int    `json:"tlb_entries"`
	TLBReplacement     string `json:"tlb_replacement"`
	CacheEntries       int    `json:"cache_entries"`
	CacheReplacement   string `json:"cache_replacement"`
	CacheDelay         int    `json:"cache_delay"`
	LogLevel           string `json:"log_level"`
	TamanioPagina      int    `json:"tam_pagina"`
	CantEntradasTabla  int    `json:"cant_entrdas_tabla"` // chequear
	CantNiveles        int    `json:"cant_niveles"`       //chequear
	Identificador      int    `json:"identificador"`
}

var Config *CPUConfig

var (
	Bitmap      = make(map[int][]bool)
	TablaMarcos = make(map[int][]int)
	Mutex       sync.Mutex
)

var CanalProcesoAEjecutar = make(chan mensajes.ProcesoAEjecutar, 1)

type PeticionEscritura struct {
	PID       int    `json:"pid"`
	DirFisica int    `json:"dir_fisica"`
	Cadena    string `json:"valor"`
}

var HayInterrupcion bool
