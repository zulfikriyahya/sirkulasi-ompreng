/*
 * Reader RFID MBG: ESP32-C3 Super Mini + RC522
 *
 * Kontrak serial (USB CDC, 115200 baud):
 *   - Setiap tap kartu mengirim satu baris: 10 digit desimal + '\n'   contoh: 0012345678
 *   - Saat boot / RC522 pulih  : "READY"
 *   - RC522 tidak terdeteksi   : "ERR:RC522"  (dikirim sekali, firmware terus mencoba pulih)
 * Aplikasi hanya menerima baris yang cocok persis ^\d{10}$ sebagai tap.
 *
 * Library : MFRC522 (GithubCommunity / miguelbalboa)
 * Board   : ESP32C3 Dev Module, "USB CDC On Boot" = Enabled
 */

#include <SPI.h>
#include <MFRC522.h>

// ---- Pin (ESP32-C3 Super Mini -> RC522) ------------------------------------
constexpr uint8_t PIN_SCK  = 4;   // RC522 SCK
constexpr uint8_t PIN_MISO = 5;   // RC522 MISO
constexpr uint8_t PIN_MOSI = 6;   // RC522 MOSI
constexpr uint8_t PIN_SS   = 7;   // RC522 SDA (SS)
constexpr uint8_t PIN_RST  = 3;   // RC522 RST
constexpr uint8_t PIN_LED  = 8;   // LED onboard Super Mini (aktif LOW)

// ---- Pengaturan -------------------------------------------------------------
// Urutan byte UID agar angka sama dengan reader HID 13.56 MHz yang dipakai.
// Jika angka dari reader HID dan dari alat ini berbeda untuk kartu yang sama,
// ubah nilai ini (true <-> false).
constexpr bool     UID_LITTLE_ENDIAN = true;
constexpr uint32_t DEBOUNCE_MS       = 1500;   // kartu sama diabaikan dalam jeda ini
constexpr uint32_t HEALTH_CHECK_MS   = 5000;   // cek kesehatan modul RC522

MFRC522 rfid(PIN_SS, PIN_RST);

uint32_t lastUid      = 0xFFFFFFFF;
uint32_t lastTapMs    = 0;
uint32_t lastCheckMs  = 0;
bool     rc522Healthy = false;

static bool rc522Present() {
  byte v = rfid.PCD_ReadRegister(MFRC522::VersionReg);
  return v != 0x00 && v != 0xFF;
}

static void blink(uint16_t ms) {
  digitalWrite(PIN_LED, LOW);   // nyala
  delay(ms);
  digitalWrite(PIN_LED, HIGH);  // mati
}

// Pakai 4 byte PERTAMA UID (sama dengan reader HID). UID < 4 byte diisi nol di belakang.
static uint32_t uidToNumber(const MFRC522::Uid &uid) {
  byte b[4] = {0, 0, 0, 0};
  for (uint8_t i = 0; i < 4 && i < uid.size; i++) b[i] = uid.uidByte[i];
  if (UID_LITTLE_ENDIAN) {
    return (uint32_t)b[0] | ((uint32_t)b[1] << 8) |
           ((uint32_t)b[2] << 16) | ((uint32_t)b[3] << 24);
  }
  return ((uint32_t)b[0] << 24) | ((uint32_t)b[1] << 16) |
         ((uint32_t)b[2] << 8) | (uint32_t)b[3];
}

static void initReader() {
  rfid.PCD_Init();
  delay(10);
  if (rc522Present()) {
    rfid.PCD_SetAntennaGain(rfid.RxGain_max);
    if (!rc522Healthy) Serial.println("READY");
    rc522Healthy = true;
  } else {
    if (rc522Healthy || lastCheckMs == 0) Serial.println("ERR:RC522");
    rc522Healthy = false;
  }
}

void setup() {
  pinMode(PIN_LED, OUTPUT);
  digitalWrite(PIN_LED, HIGH);

  Serial.begin(115200);
  uint32_t t0 = millis();
  while (!Serial && millis() - t0 < 2000) delay(10);   // tunggu USB, tapi jangan macet

  SPI.begin(PIN_SCK, PIN_MISO, PIN_MOSI, PIN_SS);
  initReader();
  lastCheckMs = millis();
}

void loop() {
  uint32_t now = millis();

  // Cek kesehatan modul secara berkala; coba pulih jika terlepas.
  if (now - lastCheckMs >= HEALTH_CHECK_MS) {
    lastCheckMs = now;
    if (!rc522Present()) {
      initReader();
    } else if (!rc522Healthy) {
      initReader();
    }
  }
  if (!rc522Healthy) {
    delay(50);
    return;
  }

  if (!rfid.PICC_IsNewCardPresent()) return;
  if (!rfid.PICC_ReadCardSerial()) return;

  uint32_t uid = uidToNumber(rfid.uid);

  bool duplicate = (uid == lastUid) && (now - lastTapMs < DEBOUNCE_MS);
  if (!duplicate) {
    char line[16];
    snprintf(line, sizeof(line), "%010lu", (unsigned long)uid);
    Serial.println(line);
    lastUid   = uid;
    lastTapMs = now;
    blink(80);
  }

  rfid.PICC_HaltA();
  rfid.PCD_StopCrypto1();
}
