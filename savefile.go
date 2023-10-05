package main

import (
	"encoding/gob"
	"fmt"
	"os"

	"github.com/klauspost/compress/zstd"
)

// Struct of savefile data
type Savefile struct {
	Name 		string
	Data		[]byte
}

// Struct to contain savefiles
type SavefileGroup struct {
	Name 		string
	Savefiles 	[]Savefile
}

// Loads savefile from path and compresses
func LoadSavefile(path string, name string) (*Savefile, error) {
	rawBytes, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	compressedBytes, err := CompressBytes(rawBytes)
	if err != nil {
		return nil, err
	}

	return &Savefile{name, compressedBytes}, nil
}

// Decompresses savefile and exports to file
func (s *Savefile) Export(path string) error {
	decompressedBytes, err := DecompressBytes(s.Data)
	if err != nil {
		return err
	}

	outputFile, err := os.Create(path)
	if err != nil {
		return err
	}

	outputFile.Write(decompressedBytes)
	outputFile.Close()

	return nil
}

// Loads savefile group from file
func LoadSavefileGroup(path string) (*SavefileGroup, error) {
	sg := SavefileGroup{}

	sgFile, err := os.Open(path)
	defer sgFile.Close()
	if err != nil {
		return nil, err
	}

	decoder := gob.NewDecoder(sgFile)
	err = decoder.Decode(&sg)
	if err != nil {
		return nil, err
	}

	return &sg, nil
}

// Exports savefile group to file
func (sg *SavefileGroup) Export(path string) error {
	exportFile, err := os.Create(path)
	defer exportFile.Close()
	if err != nil {
		return err
	}
	
	encoder := gob.NewEncoder(exportFile)
	err = encoder.Encode(sg)
	return err
}

// Searches savefiles in group for a name and returns pointer to savefile
func (sg *SavefileGroup) GetSavefile(name string) (*Savefile, error) {
	for _, savefile := range sg.Savefiles {
		if savefile.Name == name {
			return &savefile, nil
		}
	}

	return nil, fmt.Errorf("Savefile not found: %s", name)
}

// Compression of a byte array by zstd
func CompressBytes(src []byte) ([]byte, error) {
	encoder, err := zstd.NewWriter(nil)
	if err != nil {
		return []byte{}, err
	}
	return encoder.EncodeAll(src, make([]byte, 0, len(src))), nil
}

// Decompression of a byte array by zstd
func DecompressBytes(src []byte) ([]byte, error) {
	decoder, err := zstd.NewReader(nil, zstd.WithDecoderConcurrency(0))
	if err != nil {
		return []byte{}, err
	}
	return decoder.DecodeAll(src, nil)
}