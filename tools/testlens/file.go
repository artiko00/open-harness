package main

import "os"

// fileExists checks if a regular file exists at the given path. Exige que sea
// regular para que un candidato de test que resulte ser un FIFO/dispositivo no
// llegue a contieneMarcador, cuyo os.Open bloquearía sobre un named pipe.
func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.Mode().IsRegular()
}