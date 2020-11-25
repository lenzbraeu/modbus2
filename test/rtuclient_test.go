// Copyright 2014 Quoc-Viet Nguyen. All rights reserved.
// This software may be modified and distributed under the terms
// of the BSD license.  See the LICENSE file for details.

package test

import (
	"log"
	"os"
	"sync"
	"testing"

	modbus "go.lenzbraeu.de/modbus2"
)

const (
	rtuDevice = "/dev/pts/2"
)

func TestRTUClient(t *testing.T) {
	// Diagslave does not support broadcast id.
	handler := modbus.NewRTUClientHandler(rtuDevice)
	ClientTestAll(t, modbus.NewClient(17, handler))
}

type result struct {
	bytes []byte
	err   error
}

func TestRTUClientAdvancedUsage(t *testing.T) {
	handler := modbus.NewRTUClientHandler(rtuDevice)
	handler.BaudRate = 19200
	handler.DataBits = 8
	handler.Parity = "E"
	handler.StopBits = 1
	handler.Logger = log.New(os.Stdout, "rtu: ", log.LstdFlags)
	err := handler.Connect()
	if err != nil {
		t.Fatal(err)
	}
	defer handler.Close()

	var wg sync.WaitGroup
	results := make(chan result)
	slaveIDs := []byte{
		11,
		24,
		32,
		47,
	}

	for _, slaveID := range slaveIDs {
		client := modbus.NewClient(slaveID, handler)

		wg.Add(1)
		go func(slaveID byte) {
			bytes, err := client.ReadDiscreteInputs(15, 2)
			results <- result{bytes, err}
			wg.Done()
		}(slaveID)

		wg.Add(1)
		go func(slaveID byte) {
			bytes, err := client.ReadWriteMultipleRegisters(0, 2, 2, 2, []byte{1, 2, 3, 4})
			results <- result{bytes, err}
			wg.Done()
		}(slaveID)
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	for r := range results {
		log.Printf("r: %+v", r)
		if r.err != nil || r.bytes == nil {
			t.Fatal(r.err, r.bytes)
		}
	}
}
