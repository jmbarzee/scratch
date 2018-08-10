package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"syscall"
	"unsafe"
)

const n = 10

func main() {
	t := int(unsafe.Sizeof(int64(0)) * n)

	// Initialize mapped file
	map_file, err := os.OpenFile("/tmp/test.dat", os.O_RDWR|os.O_TRUNC|os.O_SYNC, 666)
	if err != nil {
		panic(err)
	}
	_, err = map_file.Seek(int64(t-1), io.SeekStart)
	if err != nil {
		panic(err)
	}
	blank := make([]byte, 1)
	_, err = map_file.Write(blank)
	if err != nil {
		panic(err)
	}

	fmt.Println("\n***** Mapping file to Pointer (expect blanks) *****")
	mmap, err := syscall.Mmap(int(map_file.Fd()), 0, int(t), syscall.PROT_READ|syscall.PROT_WRITE, syscall.MAP_SHARED)
	if err != nil {
		panic(err)
	}
	map_array := (*[n]int64)(unsafe.Pointer(&mmap[0]))
	fmt.Printf("Pointer: %v\n", *map_array)
	fmt.Printf("File:    %v\n", intsFromFile(map_file))

	fmt.Println("\n***** Writing through Pointer (expect n*2) *****")
	for i := int64(0); i < n; i++ {
		map_array[i] = i + i
	}
	fmt.Printf("Pointer: %v\n", *map_array)
	fmt.Printf("File:    %v\n", intsFromFile(map_file))

	fmt.Println("\n***** Writing through File (expect n^2) *****")
	other_array := make([]int64, n)
	for i := int64(0); i < n; i++ {
		other_array[i] = i * i
	}
	_, err = map_file.Seek(0, io.SeekStart)
	if err != nil {
		panic(err)
	}
	err = binary.Write(map_file, binary.LittleEndian, other_array)
	if err != nil {
		panic(err)
	}
	err = map_file.Sync()
	if err != nil {
		panic(err)
	}
	fmt.Printf("Pointer: %v\n", *map_array)
	fmt.Printf("File:    %v\n", intsFromFile(map_file))

	// unmap memory and close file
	err = syscall.Munmap(mmap)
	if err != nil {
		panic(err)
	}
	err = map_file.Close()
	if err != nil {
		panic(err)
	}
}

func intsFromFile(map_file *os.File) []int64 {
	_, err := map_file.Seek(0, io.SeekStart)
	if err != nil {
		panic(err)
	}
	b, err := ioutil.ReadAll(map_file)
	if err != nil {
		panic(err)
	}
	ints := make([]int64, n)
	err = binary.Read(bytes.NewReader(b), binary.LittleEndian, &ints)
	if err != nil {
		panic(err)
	}
	return ints
}
