# CodeRealm – MI-támogatott, játékosított programozás-oktató platform (Backend)

## 1. Projektösszefoglaló és célkitűzés

A **CodeRealm** egy modern, interaktív programozás-oktató rendszer, amely a tanári feladatkiadást, az automatizált kódértékelést, a tanulói motivációt támogató játékosítást (XP, szintek, napi kihívások, virtuális bolt) és a fokozatos, pedagógiailag ellenőrzött MI-alapú segítségnyújtást ötvözi egyetlen integrált keretrendszerben.

A szakdolgozat központi fókusza a kódértékeléshez kapcsolt, adaptív és fokozatos MI-segítségnyújtási modell kidolgozása, amely Python és további nyelveken támogatja a diákokat a programozási hibák feltárásában és megértésében anélkül, hogy a kész megoldást indokolatlanul felfedné.

## 2. Rendszerarchitektúra és mikroszolgáltatások

A backend Go nyelven megvalósított, mikroszolgáltatás-alapú tiszta architektúrára (Clean Architecture) épül. Az elosztott felépítés garantálja az egyes funkcionális területek független működését, tesztelhetőségét és a kódfuttatás biztonságos izolációját.

```
                    ┌──────────────────────────┐
                    │       API Gateway        │  :8081
                    │ (Keycloak Auth / Router) │
                    └─────────────┬────────────┘
                                  │
       ┌──────────────┬───────────┼───────────┬──────────────┐
       ▼              ▼           ▼           ▼              ▼
┌─────────────┐ ┌───────────┐ ┌─────────┐ ┌──────────┐ ┌───────────┐
│  Learning   │ │   Test    │ │  User   │ │  Group   │ │  Piston   │
│   Module    │ │  Service  │ │ Service │ │ Service  │ │  Sandbox  │
│    :8082    │ │   :8083   │ │  :8084  │ │  :8085   │ │ Execution │
└──────┬──────┘ └─────┬─────┘ └────┬────┘ └────┬─────┘ └───────────┘
       │              │            │           │
       └──────────────┴─────┬──────┴───────────┘
                            ▼
              ┌───────────────────────────┐
              │ PostgreSQL / Redis / RMQ  │
              └───────────────────────────┘
```

### Szolgáltatások áttekintése:

1. **API Gateway (`services/api-gateway`, port: `8081`):**
   - Központi belépési pont a kliensek számára.
   - JWT validáció és Keycloak OAuth2 / OpenID Connect tokenellenőrzés.
   - Biztonsági fejlécek, CORS-kezelés és kérésátirányítás a belső szolgáltatások felé.

2. **Learning Module Service (`services/learning-module-service`, port: `8082`):**
   - Kurzusok, fejezetek és leckék kezelése hierarchikus struktúrában.
   - Elméleti tananyagok, feladatleírások és kezdeti kódsablonok kiszolgálása.
   - Adatbázis-migrációk és strukturált lecke-előrehaladás követése.

3. **Test Service (`services/test-service`, port: `8083`):**
   - Vizsgák, tesztek és kódértékelő feladatok felügyelete.
   - Automatikus kódkiértékelés vezérlése a sandbox motoron keresztül.
   - Rejtett és publikus tesztesetek szétválasztott kezelése.

4. **User Service (`services/user-service`, port: `8084`):**
   - Felhasználói profilok, jogosultságok és aktivitások nyilvántartása.
   - Játékosítási motor: tapasztalati pontok (XP), szintek, ranglisták.
   - Napi kihívások (Daily Challenges), bejelentkezési naptár (Daily Logins).
   - Virtuális bolt (Storefront), avatárok, keretek és felszerelhető digitális jutalmak.

5. **Group Service (`services/group-service`, port: `8085`):**
   - Iskolai csoportok, osztályok és kurzus-hozzárendelések adminisztrációja.
   - Tanári és tanulói csoporttagságok kezelése.

### Adatinfrastruktúra:
- **PostgreSQL:** Relációs adattárolás különálló táblasémákkal és verziózott SQL migrációkkal.
- **Redis:** Munkamenet-adatok, leckék és ranglisták hatékony gyorsítótárazása.
- **RabbitMQ:** Aszinkron eseménykezelés és üzenetküldés a szolgáltatások között (pl. leckebefejezés, XP-jóváírás).

## 3. Oktatási folyamat és kódfuttatási biztonsági modell

### Oktatási folyamat:
1. **Kurzus és feladatkiadás:** A tanár létrehozza a kurzust és a programozási feladatot referencia-megoldással, publikus mintatesztekkel és rejtett értékelő tesztesetekkel.
2. **Kódolás és beküldés:** A tanuló a felületen megírja a megoldását, majd beküldi ellenőrzésre.
3. **Elkülönített értékelés:** A backend sandboxban lefordítja és lefutatja a kódot a tesztesetekkel. A diák a publikus eredményeket, a kimenetet és az esetleges hibákat látja; a referencia-kód és a rejtett tesztek biztonságban maradnak.
4. **Fokozatos segítségnyújtás:** Elakadás esetén a tanuló fokozatos segítséget vehet igénybe (koncepcionális magyarázat → hibaterület kijelölése → lépésenkénti stratégia).
5. **Haladás és analitika:** A rendszer rögzíti a futási eredményeket, időráfordítást és a segítségkérések számát, átfogó képet adva a tanárnak a diákok előrehaladásáról.

### Biztonsági sandbox modell:
- A tanulói kódok futtatása izolált Piston sandbox környezetben történik.
- **Erőforrás-korlátok:** Szigorú CPU-, memória- és futásidő-korlátozás védi a rendszert a végtelen ciklusoktól, memóriatúlcsordulástól és a fork bombáktól.
- **Hálózati izoláció:** A futtatási környezetből a kimenő és bejövő hálózati forgalom tiltott, így a tanulói kód nem érhet el belső szolgáltatásokat vagy külső szervereket.
- **Ideiglenes munkakönyvtár:** Minden futtatás dedikált, a kiértékelés után azonnal megsemmisített fájlrendszer-munkakörnyezetben fut.

## 4. Helyi fejlesztés és futtatás

### Előfeltételek:
- Go 1.24+ (vagy újabb)
- Docker és Docker Compose

### Beállítás:
1. Másold át a környezeti változókat:
   ```sh
   copy .env.example .env
   ```
2. Indítsd el a háttér-infrastruktúrát (PostgreSQL, Redis, RabbitMQ):
   ```sh
   docker-compose -f docker-compose.infra.yml up -d
   docker-compose -f infra/development/docker-compose.yml up -d
   ```
3. A mikroszolgáltatások fordítása és párhuzamos futtatása Windows környezetben:
   ```cmd
   start.bat
   ```
   A szkript automatikusan lefordítja mind az 5 mikroszolgáltatást a `.bin/` mappába, és külön konzolablakokban elindítja őket a megfelelő portokon (8081–8085).

### Konténeres futtatás:
A szolgáltatások Docker-konténerként is építhetők a gyökérben található `Dockerfile` és a GitHub Actions CI munkafolyamat (`.github/workflows/docker-build.yml`) segítségével.

## 5. Kapcsolódó dokumentáció és repók

- [CodeRealm Frontend GitHub-repó](https://github.com/Kiba2310/coderealm-frontend) – Felhasználói felület és komponensek
- [Frontend MI-mentor dokumentáció](https://github.com/Kiba2310/coderealm-frontend/blob/main/docs/ai-mentor.md) – Böngészőoldali WebLLM mentor architektúra
