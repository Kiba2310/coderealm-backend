@echo off
echo Inditom a backend mikroszervizeket...

if not exist ".bin" mkdir .bin

echo Epites: API Gateway...
go build -o .bin\api-gateway.exe services\api-gateway\main.go
echo Epites: Learning Module...
go build -o .bin\learning-module.exe services\learning-module-service\cmd\main.go
echo Epites: Test Service...
go build -o .bin\test-service.exe services\test-service\cmd\main.go
echo Epites: User Service...
go build -o .bin\user-service.exe services\user-service\cmd\main.go
echo Epites: Group Service...
go build -o .bin\group-service.exe services\group-service\cmd\main.go

echo Inditas...

:: API Gateway (Központi elosztó - 8081)
start "API Gateway (8081)" cmd /k ".bin\api-gateway.exe"

:: Learning Module Service (Leckék és kurzusok - 8082)
start "Learning Module Service (8082)" cmd /k ".bin\learning-module.exe"

:: Test Service (Tesztek es vizsgak - 8083)
start "Test Service (8083)" cmd /k ".bin\test-service.exe"

:: User Service (Felhasználó és jogosultság kezelés - 8084)
start "User Service (8084)" cmd /k ".bin\user-service.exe"

:: Group Service (Csoport kezelés - 8085)
start "Group Service (8085)" cmd /k ".bin\group-service.exe"

echo Minden szerviz elindult! Külön ablakokban futnak.
echo A leallitashoz egyszeruen zard be a felugro ablakokat!
