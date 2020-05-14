package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
)

func pad(b []byte) {
	for i := 0; i < len(b); i++ {
		b[i] = 0xFF
	}
}

func diff(oldPath, newPath, outDir string, offset, blockSize int) ([]int, error) {

	oldF, err := os.Open(oldPath)
	if err != nil {
		return nil, err
	}
	defer oldF.Close()

	newF, err := os.Open(newPath)
	if err != nil {
		return nil, err
	}
	defer newF.Close()

	oldBlock := make([]byte, blockSize)
	newBlock := make([]byte, blockSize)
	var oldEOF, newEOF bool = false, false
	var currentPatch *os.File
	var list []int

	for !newEOF {
		if !oldEOF {
			i, err := oldF.Read(oldBlock)
			if err == io.EOF {
				pad(oldBlock[i:])
				err = nil
				oldEOF = true
			}
			if err != nil {
				return nil, err
			}
		}
		i, err := newF.Read(newBlock)
		if err == io.EOF {
			pad(newBlock[i:])
			newEOF = true
			err = nil
		}
		if err != nil {
			return nil, err
		}
		if oldF == nil || bytes.Compare(oldBlock, newBlock) != 0 {
			if currentPatch == nil {
				currentPatch, err = os.OpenFile(filepath.Join(outDir, fmt.Sprintf("0x%x.bin", offset)), os.O_CREATE+os.O_WRONLY, 0666)
				if err != nil {
					return nil, err
				}
				list = append(list, offset)
			}
			_, err = currentPatch.Write(newBlock)
			if err != nil {
				return nil, err
			}
			if oldEOF {
				oldF = nil
			}

		} else {
			if currentPatch != nil {
				currentPatch.Close()
				currentPatch = nil
			}
		}
		offset += blockSize
	}
	if currentPatch != nil {
		currentPatch.Close()
	}

	return list, nil
}

func main() {

	list, err := diff("/tmp/n1.bin", "/tmp/n2.bin", "./out", 0x20000, 0x1000)
	if err != nil {
		log.Fatal(err)
	}

	for _, i := range list {
		fmt.Printf("0x%x out/0x%x.bin ", i, i)
	}

	fmt.Println()
}
