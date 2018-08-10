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

func main() {
	const n = 10
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

	// grab pointer for memory mapped file
	fmt.Println("\n***** Mapping file to Pointer *****")
	mmap, err := syscall.Mmap(int(map_file.Fd()), 0, int(t), syscall.PROT_READ|syscall.PROT_WRITE, syscall.MAP_SHARED)
	if err != nil {
		panic(err)
	}
	map_array := (*[n]int64)(unsafe.Pointer(&mmap[0]))

	// read through pointer
	fmt.Printf("%b\n", *map_array)

	// read through file
	_, err = map_file.Seek(0, io.SeekStart)
	if err != nil {
		panic(err)
	}
	b1, err := ioutil.ReadAll(map_file)
	if err != nil {
		panic(err)
	}
	i1 := make([]int64, n)
	err = binary.Read(bytes.NewReader(b1), binary.LittleEndian, &i1)
	fmt.Printf("%b\n", i1)

	// write through pointer
	fmt.Println("\n***** Writing through Pointer *****")
	for i := int64(0); i < n; i++ {
		map_array[i] = i * i
	}

	// read through pointer
	fmt.Printf("%b\n", *map_array)

	// read through file
	_, err = map_file.Seek(0, io.SeekStart)
	if err != nil {
		panic(err)
	}
	b2, err := ioutil.ReadAll(map_file)
	if err != nil {
		panic(err)
	}
	i2 := make([]int64, n)
	err = binary.Read(bytes.NewReader(b2), binary.LittleEndian, &i2)
	fmt.Printf("%b\n", i2)

	// write through file
	fmt.Println("\n***** Writing through File *****")
	other_array := make([]int64, n)
	for i := int64(0); i < n; i++ {
		other_array[i] = i + i
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

	// read through pointer
	fmt.Printf("%b\n", *map_array)

	// read through file
	_, err = map_file.Seek(0, io.SeekStart)
	if err != nil {
		panic(err)
	}
	b3, err := ioutil.ReadAll(map_file)
	if err != nil {
		panic(err)
	}
	i3 := make([]int64, n)
	err = binary.Read(bytes.NewReader(b3), binary.LittleEndian, &i3)
	fmt.Printf("%b\n", i3)

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
