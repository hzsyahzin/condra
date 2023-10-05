package main

import (
	"bytes"
	"os"
	"testing"
	"time"
)

func Test(t *testing.T) {
	TestSavefile(t)

	process, err := GetProcess("DARKSOULS.exe")
	if err != nil {
		panic(err)
	}

	var igtPointer uintptr

	if memory, err := GetModuleMemory(process); err != nil {
		panic(err)
	} else {
		igtPointer, err = FindPointer(process, &memory, 
			"8b 0d ? ? ? ? 8b 41 30 8b 4d 64", 2, 0)
		if err != nil {
			panic(err)
		}
	}

	getAsyncKeyState := GetAsyncKeyStateProc()
	keycodes := []Keycode{
		{"F1", 0x70},
		{"F2", 0x71},
		{"F3", 0x72},
		{"PgUp", 0x21},
		{"PgDown", 0x22},
	}

	for {
		for i, code := range keycodes {
			ret, _, _ := getAsyncKeyState.Call(uintptr(code.Code))
			if (ret & 0x01) != 0 {
				t.Logf("Key %s pressed.", keycodes[i].Key)
			}
		}

		igt, err := ReadData[int32](process, igtPointer, 0x68)
		if err != nil {
			panic("IGT unable to be fetched.")
		}

		igtTime := time.Unix(0, 
			int64(igt.(uint32)) * int64(time.Millisecond)).UTC().Format("15:04:05.000")
		
		t.Logf("IGT:\t%s\n", igtTime)

		time.Sleep(200 * time.Millisecond)
	}
}

// TestIGT
func _(t *testing.T) {
	// Fetch process and module handles
	process, err := GetProcess("DARKSOULS.exe")
	if err != nil {
		panic(err)
	}

	// Declare pointers
	var (
		igtPointer uintptr
	)

	// Get pointers from memory; ensures memory only available in this scope
	if memory, err := GetModuleMemory(process); err != nil {
		panic(err)
	} else {
		igtPointer, err = FindPointer(process, &memory, 
			"8b 0d ? ? ? ? 8b 41 30 8b 4d 64", 2, 0)
		if err != nil {
			panic(err)
		}
	}

	for {
		igt, err := ReadData[int32](process, igtPointer, 0x68)
		if err != nil {
			panic("IGT unable to be fetched.")
		}

		igtTime := time.Unix(0, 
			int64(igt.(uint32)) * int64(time.Millisecond)).UTC().Format("15:04:05.000")
		
		t.Logf("IGT:\t%s\n", igtTime)

		time.Sleep(500 * time.Millisecond)
	}
}

// TestHotkeys
func _(t *testing.T) {
	getAsyncKeyState := GetAsyncKeyStateProc()
	keycodes := []Keycode{
		{"F1", 0x70},
		{"F2", 0x71},
		{"F3", 0x72},
		{"PgUp", 0x21},
		{"PgDown", 0x22},
	}

	for {
		for i, code := range keycodes {
			ret, _, _ := getAsyncKeyState.Call(uintptr(code.Code))
			if (ret & 0x01) != 0 {
				t.Logf("Key %s pressed.", keycodes[i].Key)
			}
		}

		time.Sleep(200 * time.Millisecond)
	}
}

// TestSavefile
func TestSavefile(t *testing.T) {
	testInput := "./test/test.sl2"
	testOutput := "./test/savefileTest.condra"

	_, err := os.Open(testOutput)
	if err == os.ErrNotExist {
		t.Logf("Deleting test file at: %s.", testOutput)
        os.Remove(testOutput)
	}

	savefileGroup := SavefileGroup{"Test Group", []Savefile{}}
	t.Logf("Savefile created.")

	var savefile *Savefile
	savefile, err = LoadSavefile(testInput, "Savefile 1")
	if err != nil {
		t.Error("Failed to load savefile.")
	}
	savefileGroup.Savefiles = append(savefileGroup.Savefiles, *savefile)
	t.Log("Imported savefile 1")

	savefile, err = LoadSavefile(testInput, "Savefile 2")
	if err != nil {
		t.Error("Failed to load savefile.")
	}
	savefileGroup.Savefiles = append(savefileGroup.Savefiles, *savefile)
	t.Log("Imported savefile 2")

	err = savefileGroup.Export(testOutput)
	if err != nil {
		t.Error("Failed to export savefile group.")
	}

	loadedSavefileGroup, err := LoadSavefileGroup(testOutput)
	if err != nil {
		t.Error("Failed to load savefile group.")
	}
	t.Log("Loaded savefile group.")

	if !bytes.Equal(loadedSavefileGroup.Savefiles[0].Data, 
		savefileGroup.Savefiles[0].Data) {
		t.Error("Loaded savefile does not match existing")
	}
	t.Log("Loaded savefile matches existing")
}