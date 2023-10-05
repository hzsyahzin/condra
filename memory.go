package main

import (
	"encoding/binary"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"
	"unsafe"

	windows "golang.org/x/sys/windows"
)

// Struct of relevant process handles
type Process struct {
	processHandle windows.Handle
	moduleHandle  windows.Handle
}

// Get the process and module handles
func GetProcess(processName string) (Process, error) {
	var process Process
	pid, err := GetProcessID(processName)
	if err != nil {
		return Process{}, fmt.Errorf("Process \"%s\" not found.", processName)
	}

	fmt.Printf("DarkSouls PID:\t%d\n", pid)

	process.processHandle, err = windows.OpenProcess(windows.PROCESS_VM_READ | 
		windows.PROCESS_QUERY_INFORMATION, false, pid)
	if err != nil {
		return Process{}, fmt.Errorf("Unable to open process.")
	}

	fmt.Printf("Process handle:\t%d\n", process.processHandle)

	process.moduleHandle, err = GetModuleHandle(pid)
	if err != nil {
		return Process{}, fmt.Errorf("Unable to get main module handle.")
	}
	fmt.Printf("Module handle:\t%d\n", process.moduleHandle)

	return process, nil
}

// Scans an aob for a pattern and returns its index
func FindPointer(process Process, memory *[]byte, pattern string, offsets ...int) (uintptr, error) {
	patternPointer, err := ScanForPattern(*memory, pattern)
	if err != nil {
		return 0, fmt.Errorf("AOB not found: %s", err)
	}
	fmt.Printf("Pattern found:\t%x\n", patternPointer)

	pointer, err := GetMemoryPointer(process, patternPointer, offsets)
	if err != nil {
		return 0, fmt.Errorf("Failed to fetch pointer: %s", err)
	}
	fmt.Printf("Pointer found:\t%x\n", pointer)

	return pointer, nil
}

// Converts pattern string to array for convenience when scanning.
func PatternStrToInt(pattern string) ([]int16, error) {
	splitPattern := strings.Split(pattern, " ")
	outputPattern := make([]int16, len(splitPattern))
	for i, code := range splitPattern {
		if code == "?" {
			outputPattern[i] = -1
		} else {
			tmp, err := strconv.ParseUint(code, 16, 32)
			if err != nil {
				return nil, err
			}
			outputPattern[i] = int16(tmp)
		}
	}

	return outputPattern, nil
}

// Searches block of memory for a matching pattern of bytes.
func ScanForPattern(memory []byte, patternStr string) (uintptr, error) {
	start := time.Now()

	pattern, err := PatternStrToInt(patternStr)
	if err != nil {
		return 0, err
	}
	
	for i := 0; i < len(memory) - len(pattern); i += 1 {
		for j := 0; j < len(pattern); j += 1 {
			if pattern[j] != -1 && byte(pattern[j]) != memory[i + j] {
				break
			} else if j == len(pattern) - 1 {
				fmt.Printf("Time taken:\t%v\n", time.Since(start))
				return uintptr(i), nil
			}
		}
	}

	return 0, fmt.Errorf("AOB not found")
}

// Each pointer offset relative to the current pointer in a pointer chain.
// The first one reads in the module code to get the first pointer.
// Each subsequent one reads directly into memory to get the next pointer.
func GetMemoryPointer(process Process, patternPointer uintptr, pointerOffsets []int) (uintptr, error) {
	var pointerBytes []byte

	currentPointer := uintptr(process.moduleHandle) + uintptr(patternPointer)
	for _, offset := range pointerOffsets {
		pointerBytes = make([]byte, 4)
		err := windows.ReadProcessMemory(process.processHandle,
			currentPointer + uintptr(offset),
			&pointerBytes[0], uintptr(len(pointerBytes)), new(uintptr))
		if err != nil {
			return 0, err
		}
		currentPointer = uintptr(binary.NativeEndian.Uint32(pointerBytes))
	}

	return uintptr(binary.NativeEndian.Uint32(pointerBytes)), nil
}

// Reads memory of Sizeof(T) bytes and consequently converts to type T.
func ReadData[T any](process Process, memoryPointer uintptr, dataOffset int) (any, error) {
	kind := reflect.TypeOf(*new(T))
	memBytes, err := ReadBytes(process, memoryPointer, dataOffset, kind.Size())
	if err != nil {
		return 0, err
	}

	switch kind.Kind() {
	case reflect.Int32:
		return binary.NativeEndian.Uint32(memBytes), nil
	default:
		return memBytes, nil
	}
}

// Reads certain number of bytes from memory given a pointer offset and length.
func ReadBytes(process Process, memoryPointer uintptr, dataOffset int, length uintptr) ([]byte, error) {
	memBytes := make([]byte, length)
	err := windows.ReadProcessMemory(process.processHandle,
		memoryPointer + uintptr(dataOffset),
		&memBytes[0], length, new(uintptr))
	if err != nil {
		return []byte{}, err
	}

	return memBytes, nil
}

// Given the process and memory handles, return entire block of memory in bytes.
func GetModuleMemory(process Process) ([]byte, error) {
	moduleInfo := windows.ModuleInfo{}
	err := windows.GetModuleInformation(process.processHandle, 
		process.moduleHandle, &moduleInfo, uint32(unsafe.Sizeof(windows.ModuleInfo{})))
	if err != nil {
		return []byte{}, err
	}

	data := make([]byte, moduleInfo.SizeOfImage)
	err = windows.ReadProcessMemory(process.processHandle, uintptr(process.moduleHandle), 
		&data[0], uintptr(moduleInfo.SizeOfImage), new(uintptr))
	if err != nil {
		return []byte{}, err
	}

	return data, nil
}

// Given the process handle, get memory handle (address) of main module
func GetModuleHandle(pid uint32) (windows.Handle, error) {
	toolHelp, err := windows.CreateToolhelp32Snapshot(windows.TH32CS_SNAPMODULE, pid)
	if err != nil {
		return 0, err
	}

	moduleEntry := windows.ModuleEntry32{Size: uint32(unsafe.Sizeof(windows.ModuleEntry32{}))}
	err = windows.Module32First(toolHelp, &moduleEntry)
	if err != nil {
		return 0, err
	}

	return moduleEntry.ModuleHandle, nil
}

// Takes the name of the process, enumerates processes and searches for name
func GetProcessID(name string) (uint32, error) {
	toolHelp, err := windows.CreateToolhelp32Snapshot(windows.TH32CS_SNAPPROCESS, 0)
	if err != nil {
		return 0, err
	}

	processEntry := windows.ProcessEntry32{Size: uint32(unsafe.Sizeof(windows.ProcessEntry32{}))}
	for {
		err := windows.Process32Next(toolHelp, &processEntry)
		if err != nil {
			return 0, err
		}

		if windows.UTF16ToString(processEntry.ExeFile[:]) == name {
			return processEntry.ProcessID, nil
		}
	}
}