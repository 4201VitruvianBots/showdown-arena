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

// Received command state (from node_command messages)
static String receivedHubState = "DISABLED";
static float receivedMotorDuty = 0.0f;
static String receivedLedPattern = "off";
static int receivedMatchState = 0;
static bool receivedStackRed = false;
static bool receivedStackBlue = false;
static bool receivedStackOrange = false;
static bool receivedStackGreen = false;
static bool receivedStackBuzzer = false;

// JSON parsing buffer
static StaticJsonDocument<1024> inDoc;

// --- Internal handler for unified node_command ---

static void handleNodeCommand(JsonDocument &doc) {
  // Hub fields (relevant for hub roles)
  if (doc.containsKey("hubState")) {
    receivedHubState = doc["hubState"].as<String>();
  }
  if (doc.containsKey("motorDuty")) {
    receivedMotorDuty = doc["motorDuty"].as<float>();
  }
  if (doc.containsKey("ledPattern")) {
    receivedLedPattern = doc["ledPattern"].as<String>();
  }

  // Stack light fields (relevant for estop/table roles)
  if (doc.containsKey("stackRed")) {
    receivedStackRed = doc["stackRed"].as<bool>();
  }
  if (doc.containsKey("stackBlue")) {
    receivedStackBlue = doc["stackBlue"].as<bool>();
  }
  if (doc.containsKey("stackOrange")) {
    receivedStackOrange = doc["stackOrange"].as<bool>();
  }
  if (doc.containsKey("stackGreen")) {
    receivedStackGreen = doc["stackGreen"].as<bool>();
  }
  if (doc.containsKey("stackBuzzer")) {
    receivedStackBuzzer = doc["stackBuzzer"].as<bool>();
  }

  // Match state (relevant for all roles)
  if (doc.containsKey("matchState")) {
    receivedMatchState = doc["matchState"].as<int>();
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
    char buf[1024];
    int len = udp.read(buf, sizeof(buf) - 1);
    if (len > 0) {
      buf[len] = '\0';
      inDoc.clear();
      DeserializationError err = deserializeJson(inDoc, buf);
      if (err) {
        Serial.printf("UDP Comms: JSON parse error: %s\n", err.c_str());
      } else {
        // Update health timestamp on any valid message
        lastCommandReceivedMs = millis();

        String msgType = inDoc["type"].as<String>();
        if (msgType == "node_command") {
          handleNodeCommand(inDoc);
        } else {
          Serial.printf("UDP Comms: Unknown message type: %s\n", msgType.c_str());
        }
      }
    }

    packetSize = udp.parsePacket();
  }
}

void udp_comms_sendNodeStatus(const ButtonStates &states, uint32_t score, const String &role) {
  if (!initialized || !isEthernetConnected()) {
    return;
  }

  StaticJsonDocument<384> doc;
  doc["type"] = "node_status";
  doc["role"] = role;

  // Field estop (relevant for FMS_TABLE)
  if (role == "FMS_TABLE") {
    doc["fieldEStop"] = states.fieldEStop;
  } else {
    doc["fieldEStop"] = false;
  }

  // Station estops (relevant for alliance roles)
  JsonArray stations = doc.createNestedArray("stations");
  if (role == "RED_ALLIANCE" || role == "BLUE_ALLIANCE") {
    JsonObject s1 = stations.createNestedObject();
    s1["eStop"] = states.station1EStop;
    s1["aStop"] = states.station1AStop;

    JsonObject s2 = stations.createNestedObject();
    s2["eStop"] = states.station2EStop;
    s2["aStop"] = states.station2AStop;

    JsonObject s3 = stations.createNestedObject();
    s3["eStop"] = states.station3EStop;
    s3["aStop"] = states.station3AStop;
  } else {
    // Non-alliance roles send empty station data
    for (int i = 0; i < 3; i++) {
      JsonObject station = stations.createNestedObject();
      station["eStop"] = false;
      station["aStop"] = false;
    }
  }

  // Score (relevant for hub roles)
  doc["score"] = (int)score;

  char buf[384];
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

String udp_comms_getLedPattern(void) {
  return receivedLedPattern;
}

int udp_comms_getMatchState(void) {
  return receivedMatchState;
}

bool udp_comms_getStackRed(void) {
  return receivedStackRed;
}

bool udp_comms_getStackBlue(void) {
  return receivedStackBlue;
}

bool udp_comms_getStackOrange(void) {
  return receivedStackOrange;
}

bool udp_comms_getStackGreen(void) {
  return receivedStackGreen;
}

bool udp_comms_getStackBuzzer(void) {
  return receivedStackBuzzer;
}
