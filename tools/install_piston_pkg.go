//go:build ignore
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"


	"github.com/joho/godotenv"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Használat: go run install_piston_pkg.go <nyelv> [verzió]")
		fmt.Println("Példa: go run install_piston_pkg.go python 3.12.0")
		fmt.Println("Ha a verzió hiányzik, '*' (legújabb) lesz használva.")
		return
	}

	language := os.Args[1]
	version := "*"
	if len(os.Args) >= 3 {
		version = os.Args[2]
	}

	// Betöltjük a .env fájlt, ha van, hogy kiolvassuk a Piston URL-t
	_ = godotenv.Load(".env")
	
	// Használjuk a PISTON_LOCAL_URL-t az .env-ből, de átalakítjuk a packages végpontra
	// Mivel a PISTON_LOCAL_URL általában "/api/v2/execute", ezt le kell cserélnünk
	// De egyszerűbb hardkódolni az alap URL-t vagy külön env változót használni.
	// Most fixáljuk a használt szerverre a kérést:
	url := "https://pistonapi.pollak.info/api/v2/packages"

	payloadMap := map[string]string{
		"language": language,
		"version":  version,
	}
	payloadBytes, _ := json.Marshal(payloadMap)

	fmt.Printf("Telepítés indítása: %s (%s) a %s szerveren...\n", language, version, url)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(payloadBytes))
	if err != nil {
		fmt.Println("Hiba a kérés létrehozásakor:", err)
		return
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Hálózati hiba:", err)
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	fmt.Printf("Válasz kód: %d\n", resp.StatusCode)
	fmt.Printf("Válasz: %s\n", string(body))
}
