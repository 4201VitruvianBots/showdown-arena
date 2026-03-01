#include "comms/udp_comms.hpp"
#include "comms/ethernet.hpp"
#include <WiFiUdp.h>
#include <ArduinoJson.h>

// UDP socket
static WiFiUDP udp;
static String arenaServerIP;
static bool initialized = false;

// Health tracking
static unsigned long lastCommandReceivedMs = 0;
#define UDP_HEALTH_TIMEOUT_MS 2000

// Received command state
static String receivedHubState = "DISABLED";
static float receivedMotorDuty = 0.0f;
static bool receivedRedLight = false;
static bool receivedBlueLight = false;
static int receivedMatchState = 0;

// JSON parsing buffer
static StaticJsonDocument<512> inDoc;

// --- Internal handlers ---

static void handleHubCommand(JsonDocument &doc) {
  if (doc.containsKey("hubState")) {
    receivedHubState = doc["hubState"].as<String>();
  }
  if (doc.containsKey("motorDuty")) {
    receivedMotorDuty = doc["motorDuty"].as<float>();
  }
  if (doc.containsKey("redLight")) {
    receivedRedLight = doc["redLight"].as<bool>();
  }
  if (doc.containsKey("blueLight")) {
    receivedBlueLight = doc["blueLight"].as<bool>();
  }
}

static void handleStackLights(JsonDocument &doc) {
  // Stack lights are handled by the score table ESP32
  // For now, just log receipt. Hardware actuation can be added later.
  // The data is available but the physical stack lights may be
  // driven by the PLC or a separate mechanism.
}

static void handleMatchState(JsonDocument &doc) {
  if (doc.containsKey("state")) {
    receivedMatchState = doc["state"].as<int>();
  }
}

// --- Public API ---

void udp_comms_init(const String &arenaIP) {
  arenaServerIP = arenaIP;
  udp.begin(UDP_LISTEN_PORT);
  initialized = true;
  Serial.printf("UDP Comms: Listening on port %d, sending to %s:%d\n",
                UDP_LISTEN_PORT, arenaIP.c_str(), UDP_SEND_PORT);
}

void udp_comms_run(void) {
  if (!initialized || !isEthernetConnected()) {
    return;
  }

  int packetSize = udp.parsePacket();
  while (packetSize > 0) {
    char buf[512];
    int len = udp.read(buf, sizeof(buf) - 1);
    if (len > 0) {
      buf[len] = '\0';

      DeserializationError err = deserializeJson(inDoc, buf);
      if (err) {
        Serial.printf("UDP Comms: JSON parse error: %s\n", err.c_str());
      } else {
        // Update health timestamp on any valid message
        lastCommandReceivedMs = millis();

        String msgType = inDoc["type"].as<String>();
        if (msgType == "hub_command") {
          handleHubCommand(inDoc);
        } else if (msgType == "stack_lights") {
          handleStackLights(inDoc);
        } else if (msgType == "match_state") {
          handleMatchState(inDoc);
        } else {
          Serial.printf("UDP Comms: Unknown message type: %s\n", msgType.c_str());
        }
      }
    }

    packetSize = udp.parsePacket();
  }
}

void udp_comms_sendEstopState(const ButtonStates &states, const String &role) {
  if (!initialized || !isEthernetConnected()) {
    return;
  }

  StaticJsonDocument<300> doc;
  doc["type"] = "estop_state";
  doc["role"] = role;

  if (role == "FMS_TABLE") {
    doc["fieldEStop"] = states.fieldEStop;
    // FMS_TABLE doesn't have station estops, but send empty array for consistency
    JsonArray stations = doc.createNestedArray("stations");
    for (int i = 0; i < 3; i++) {
      JsonObject station = stations.createNestedObject();
      station["eStop"] = true;  // Default: not pressed
      station["aStop"] = false;
    }
  } else {
    doc["fieldEStop"] = false; // Alliance stations don't report field estop
    JsonArray stations = doc.createNestedArray("stations");

    JsonObject s1 = stations.createNestedObject();
    s1["eStop"] = states.station1EStop;
    s1["aStop"] = states.station1AStop;

    JsonObject s2 = stations.createNestedObject();
    s2["eStop"] = states.station2EStop;
    s2["aStop"] = states.station2AStop;

    JsonObject s3 = stations.createNestedObject();
    s3["eStop"] = states.station3EStop;
    s3["aStop"] = states.station3AStop;
  }

  char buf[300];
  size_t len = serializeJson(doc, buf, sizeof(buf));

  IPAddress serverIP;
  if (serverIP.fromString(arenaServerIP)) {
    udp.beginPacket(serverIP, UDP_SEND_PORT);
    udp.write((const uint8_t *)buf, len);
    udp.endPacket();
  }
}

void udp_comms_sendScoreReport(uint32_t score, const String &role) {
  if (!initialized || !isEthernetConnected()) {
    return;
  }

  StaticJsonDocument<128> doc;
  doc["type"] = "score_report";
  doc["role"] = role;
  doc["score"] = score;

  char buf[128];
  size_t len = serializeJson(doc, buf, sizeof(buf));

  IPAddress serverIP;
  if (serverIP.fromString(arenaServerIP)) {
    udp.beginPacket(serverIP, UDP_SEND_PORT);
    udp.write((const uint8_t *)buf, len);
    udp.endPacket();
  }
}

bool udp_comms_isServerConnected(void) {
  if (lastCommandReceivedMs == 0) {
    return false;
  }
  return (millis() - lastCommandReceivedMs) < UDP_HEALTH_TIMEOUT_MS;
}

// --- Received command accessors ---

String udp_comms_getHubState(void) {
  return receivedHubState;
}

float udp_comms_getMotorDuty(void) {
  return receivedMotorDuty;
}

bool udp_comms_getRedLight(void) {
  return receivedRedLight;
}

bool udp_comms_getBlueLight(void) {
  return receivedBlueLight;
}

int udp_comms_getMatchState(void) {
  return receivedMatchState;
}
