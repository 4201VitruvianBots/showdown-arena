// Copyright 20## Team ###. All Rights Reserved.
// Author: cpapplefamily@gmail.com (Corey Applegate)
//
// ESP32 IO handlers using UDP JSON communication.
// Replaces the previous TCP/HTTP-based health checking and data exchange.
//
// Architecture:
//   - Go listens on udpListenPort (5301) for JSON messages from ESP32s (many→1)
//   - Go sends JSON commands to each ESP32 IP on udpSendPort (5300) (1→many)
//   - Each ESP32 identifies itself by "role" in its messages
//   - Health is determined by receiving any valid message within the timeout window

package plc

import (
	"encoding/json"
	"log"
	"net"
	"strings"
	"sync"
	"time"
)

const (
	udpListenPort      = 5301 // Port Go listens on for ESP32→Go messages
	udpSendPort        = 5300 // Port ESP32s listen on for Go→ESP32 commands
	udpHealthTimeoutMs = 2000 // Mark unhealthy if no message received for this long
	udpSendIntervalMs  = 100  // How often to send commands to ESP32s
	udpReadBufferSize  = 2048
)

// Esp32 defines the interface for ESP32 communication.
type Esp32 interface {
	Run()

	// Health configuration
	IsScoreTableIOEnabled() bool
	IsRedEstopsEnabled() bool
	IsBlueEstopsEnabled() bool
	IsScoreTableHealthy() bool
	IsRedEstopsHealthy() bool
	IsBlueEstopsHealthy() bool

	// Address configuration
	SetScoreTableAddress(string)
	SetRedAllianceStationEstopAddress(string)
	SetBlueAllianceStationEstopAddress(string)

	// Data received from ESP32s
	GetFieldEStop() bool
	GetTeamEStops() ([3]bool, [3]bool)
	GetTeamAStops() ([3]bool, [3]bool)
	GetRedHubScore() int
	GetBlueHubScore() int

	// Commands to send to ESP32s
	SendHubCommand(hubState string, motorDuty float64, redLight, blueLight bool)
	SendStackLights(red, blue, orange, green, buzzer bool)
	SendMatchState(state int)
}

// --- Inbound message types (ESP32 → Go) ---

// EstopStationState represents one driver station's stop button states.
type EstopStationState struct {
	EStop bool `json:"eStop"`
	AStop bool `json:"aStop"`
}

// EstopStateMessage is sent by ESP32s to report estop/astop button states.
type EstopStateMessage struct {
	Type       string              `json:"type"` // "estop_state"
	Role       string              `json:"role"` // "RED_ALLIANCE", "BLUE_ALLIANCE", "FMS_TABLE"
	FieldEStop bool                `json:"fieldEStop"`
	Stations   [3]EstopStationState `json:"stations"`
}

// ScoreReportMessage is sent by ESP32s to report scoring data.
type ScoreReportMessage struct {
	Type  string `json:"type"`  // "score_report"
	Role  string `json:"role"`  // "RED_HUB", "BLUE_HUB"
	Score int    `json:"score"`
}

// --- Outbound message types (Go → ESP32) ---

// HubCommandMessage commands a scoring hub ESP32.
type HubCommandMessage struct {
	Type      string  `json:"type"`      // "hub_command"
	HubState  string  `json:"hubState"`  // "DISABLED", "SCORING_ACTIVE", "SCORING_INACTIVE", etc.
	MotorDuty float64 `json:"motorDuty"`
	RedLight  bool    `json:"redLight"`
	BlueLight bool    `json:"blueLight"`
}

// StackLightsMessage commands the score table stack lights.
type StackLightsMessage struct {
	Type   string `json:"type"` // "stack_lights"
	Red    bool   `json:"red"`
	Blue   bool   `json:"blue"`
	Orange bool   `json:"orange"`
	Green  bool   `json:"green"`
	Buzzer bool   `json:"buzzer"`
}

// MatchStateMessage sends the current match state to ESP32s.
type MatchStateMessage struct {
	Type  string `json:"type"`  // "match_state"
	State int    `json:"state"`
}

// InboundMessage is used for initial JSON unmarshalling to determine message type.
type InboundMessage struct {
	Type string `json:"type"`
}

// Esp32IO manages bidirectional UDP communication with ESP32 field nodes.
type Esp32IO struct {
	// Configured IP addresses
	ScoreTableIP         string
	RedAllianceEstopsIP  string
	BlueAllianceEstopsIP string

	// Health tracking — last time a valid message was received from each node
	scoreTableLastHeard     time.Time
	redEstopsLastHeard      time.Time
	blueEstopsLastHeard     time.Time

	// Received state from ESP32s
	fieldEStop    bool
	redEStops     [3]bool
	redAStops     [3]bool
	blueEStops    [3]bool
	blueAStops    [3]bool
	redHubScore   int
	blueHubScore  int

	// Outbound command state (set by the arena, sent periodically)
	hubCommand  HubCommandMessage
	stackLights StackLightsMessage
	matchState  MatchStateMessage

	// UDP connection
	listenConn *net.UDPConn

	mu sync.Mutex
}

func (esp32 *Esp32IO) SetScoreTableAddress(address string) {
	address = strings.TrimSpace(address)
	if address != "" && net.ParseIP(address) == nil {
		log.Printf("Invalid Score Table IP address: %s", address)
		return
	}
	esp32.mu.Lock()
	esp32.ScoreTableIP = address
	esp32.mu.Unlock()
	if address != "" {
		log.Printf("Set Score Table IP to: %s", address)
	}
}

func (esp32 *Esp32IO) SetRedAllianceStationEstopAddress(address string) {
	address = strings.TrimSpace(address)
	if address != "" && net.ParseIP(address) == nil {
		log.Printf("Invalid Red Alliance Estops IP address: %s", address)
		return
	}
	esp32.mu.Lock()
	esp32.RedAllianceEstopsIP = address
	esp32.mu.Unlock()
	if address != "" {
		log.Printf("Set Red Alliance Estops IP to: %s", address)
	}
}

func (esp32 *Esp32IO) SetBlueAllianceStationEstopAddress(address string) {
	address = strings.TrimSpace(address)
	if address != "" && net.ParseIP(address) == nil {
		log.Printf("Invalid Blue Alliance Estops IP address: %s", address)
		return
	}
	esp32.mu.Lock()
	esp32.BlueAllianceEstopsIP = address
	esp32.mu.Unlock()
	if address != "" {
		log.Printf("Set Blue Alliance Estops IP to: %s", address)
	}
}

// --- Health / Enabled checks ---

func (esp32 *Esp32IO) IsScoreTableIOEnabled() bool {
	esp32.mu.Lock()
	defer esp32.mu.Unlock()
	return esp32.ScoreTableIP != ""
}

func (esp32 *Esp32IO) IsRedEstopsEnabled() bool {
	esp32.mu.Lock()
	defer esp32.mu.Unlock()
	return esp32.RedAllianceEstopsIP != ""
}

func (esp32 *Esp32IO) IsBlueEstopsEnabled() bool {
	esp32.mu.Lock()
	defer esp32.mu.Unlock()
	return esp32.BlueAllianceEstopsIP != ""
}

func (esp32 *Esp32IO) IsScoreTableHealthy() bool {
	esp32.mu.Lock()
	defer esp32.mu.Unlock()
	if esp32.ScoreTableIP == "" {
		return false
	}
	return time.Since(esp32.scoreTableLastHeard) < time.Millisecond*udpHealthTimeoutMs
}

func (esp32 *Esp32IO) IsRedEstopsHealthy() bool {
	esp32.mu.Lock()
	defer esp32.mu.Unlock()
	if esp32.RedAllianceEstopsIP == "" {
		return false
	}
	return time.Since(esp32.redEstopsLastHeard) < time.Millisecond*udpHealthTimeoutMs
}

func (esp32 *Esp32IO) IsBlueEstopsHealthy() bool {
	esp32.mu.Lock()
	defer esp32.mu.Unlock()
	if esp32.BlueAllianceEstopsIP == "" {
		return false
	}
	return time.Since(esp32.blueEstopsLastHeard) < time.Millisecond*udpHealthTimeoutMs
}

// --- Data accessors (read state received from ESP32s) ---

func (esp32 *Esp32IO) GetFieldEStop() bool {
	esp32.mu.Lock()
	defer esp32.mu.Unlock()
	return esp32.fieldEStop
}

func (esp32 *Esp32IO) GetTeamEStops() ([3]bool, [3]bool) {
	esp32.mu.Lock()
	defer esp32.mu.Unlock()
	return esp32.redEStops, esp32.blueEStops
}

func (esp32 *Esp32IO) GetTeamAStops() ([3]bool, [3]bool) {
	esp32.mu.Lock()
	defer esp32.mu.Unlock()
	return esp32.redAStops, esp32.blueAStops
}

func (esp32 *Esp32IO) GetRedHubScore() int {
	esp32.mu.Lock()
	defer esp32.mu.Unlock()
	return esp32.redHubScore
}

func (esp32 *Esp32IO) GetBlueHubScore() int {
	esp32.mu.Lock()
	defer esp32.mu.Unlock()
	return esp32.blueHubScore
}

// --- Command setters (called by arena, sent to ESP32s periodically) ---

func (esp32 *Esp32IO) SendHubCommand(hubState string, motorDuty float64, redLight, blueLight bool) {
	esp32.mu.Lock()
	defer esp32.mu.Unlock()
	esp32.hubCommand = HubCommandMessage{
		Type:      "hub_command",
		HubState:  hubState,
		MotorDuty: motorDuty,
		RedLight:  redLight,
		BlueLight: blueLight,
	}
}

func (esp32 *Esp32IO) SendStackLights(red, blue, orange, green, buzzer bool) {
	esp32.mu.Lock()
	defer esp32.mu.Unlock()
	esp32.stackLights = StackLightsMessage{
		Type:   "stack_lights",
		Red:    red,
		Blue:   blue,
		Orange: orange,
		Green:  green,
		Buzzer: buzzer,
	}
}

func (esp32 *Esp32IO) SendMatchState(state int) {
	esp32.mu.Lock()
	defer esp32.mu.Unlock()
	esp32.matchState = MatchStateMessage{
		Type:  "match_state",
		State: state,
	}
}

// --- Main run loop ---

// Run starts the UDP listener and periodic sender. Should be called as a goroutine.
func (esp32 *Esp32IO) Run() {
	// Initialize default command states
	esp32.hubCommand = HubCommandMessage{Type: "hub_command", HubState: "DISABLED"}
	esp32.stackLights = StackLightsMessage{Type: "stack_lights"}
	esp32.matchState = MatchStateMessage{Type: "match_state"}

	// Start the UDP listener
	go esp32.runListener()

	// Periodically send commands to all configured ESP32s
	for {
		startTime := time.Now()
		esp32.sendCommandsToAll()
		elapsed := time.Since(startTime)
		if remaining := time.Millisecond*udpSendIntervalMs - elapsed; remaining > 0 {
			time.Sleep(remaining)
		}
	}
}

// runListener listens for incoming UDP messages from ESP32s on udpListenPort.
func (esp32 *Esp32IO) runListener() {
	addr := &net.UDPAddr{Port: udpListenPort}
	var err error
	esp32.listenConn, err = net.ListenUDP("udp", addr)
	if err != nil {
		log.Printf("ESP32 UDP: Failed to listen on port %d: %v", udpListenPort, err)
		return
	}
	defer esp32.listenConn.Close()
	log.Printf("ESP32 UDP: Listening on port %d", udpListenPort)

	buf := make([]byte, udpReadBufferSize)
	for {
		n, remoteAddr, err := esp32.listenConn.ReadFromUDP(buf)
		if err != nil {
			log.Printf("ESP32 UDP: Read error: %v", err)
			continue
		}

		esp32.handleInboundMessage(buf[:n], remoteAddr)
	}
}

// handleInboundMessage parses and processes a single inbound UDP message.
func (esp32 *Esp32IO) handleInboundMessage(data []byte, remoteAddr *net.UDPAddr) {
	// Peek at the type field
	var msg InboundMessage
	if err := json.Unmarshal(data, &msg); err != nil {
		log.Printf("ESP32 UDP: Invalid JSON from %s: %v", remoteAddr, err)
		return
	}

	switch msg.Type {
	case "estop_state":
		var estopMsg EstopStateMessage
		if err := json.Unmarshal(data, &estopMsg); err != nil {
			log.Printf("ESP32 UDP: Failed to parse estop_state from %s: %v", remoteAddr, err)
			return
		}
		esp32.handleEstopState(&estopMsg, remoteAddr)

	case "score_report":
		var scoreMsg ScoreReportMessage
		if err := json.Unmarshal(data, &scoreMsg); err != nil {
			log.Printf("ESP32 UDP: Failed to parse score_report from %s: %v", remoteAddr, err)
			return
		}
		esp32.handleScoreReport(&scoreMsg, remoteAddr)

	default:
		log.Printf("ESP32 UDP: Unknown message type '%s' from %s", msg.Type, remoteAddr)
	}
}

// handleEstopState processes an estop state message from an ESP32.
func (esp32 *Esp32IO) handleEstopState(msg *EstopStateMessage, remoteAddr *net.UDPAddr) {
	esp32.mu.Lock()
	defer esp32.mu.Unlock()

	switch msg.Role {
	case "FMS_TABLE":
		esp32.fieldEStop = msg.FieldEStop
		esp32.scoreTableLastHeard = time.Now()
	case "RED_ALLIANCE":
		for i := 0; i < 3; i++ {
			esp32.redEStops[i] = msg.Stations[i].EStop
			esp32.redAStops[i] = msg.Stations[i].AStop
		}
		esp32.redEstopsLastHeard = time.Now()
	case "BLUE_ALLIANCE":
		for i := 0; i < 3; i++ {
			esp32.blueEStops[i] = msg.Stations[i].EStop
			esp32.blueAStops[i] = msg.Stations[i].AStop
		}
		esp32.blueEstopsLastHeard = time.Now()
	default:
		log.Printf("ESP32 UDP: Unknown estop role '%s' from %s", msg.Role, remoteAddr)
	}
}

// handleScoreReport processes a score report message from an ESP32.
func (esp32 *Esp32IO) handleScoreReport(msg *ScoreReportMessage, remoteAddr *net.UDPAddr) {
	esp32.mu.Lock()
	defer esp32.mu.Unlock()

	switch msg.Role {
	case "RED_HUB":
		esp32.redHubScore = msg.Score
		// Update health for score table (the hub ESP32 is co-located)
		esp32.scoreTableLastHeard = time.Now()
	case "BLUE_HUB":
		esp32.blueHubScore = msg.Score
		esp32.scoreTableLastHeard = time.Now()
	default:
		log.Printf("ESP32 UDP: Unknown score role '%s' from %s", msg.Role, remoteAddr)
	}
}

// sendCommandsToAll sends the current command state to all configured ESP32s.
func (esp32 *Esp32IO) sendCommandsToAll() {
	esp32.mu.Lock()
	hubCmd := esp32.hubCommand
	stackCmd := esp32.stackLights
	matchCmd := esp32.matchState
	scoreTableIP := esp32.ScoreTableIP
	redEstopsIP := esp32.RedAllianceEstopsIP
	blueEstopsIP := esp32.BlueAllianceEstopsIP
	esp32.mu.Unlock()

	// Send hub command + match state to the score table ESP32 (which also handles the hub)
	if scoreTableIP != "" {
		esp32.sendJSON(scoreTableIP, hubCmd)
		esp32.sendJSON(scoreTableIP, matchCmd)
		esp32.sendJSON(scoreTableIP, stackCmd)
	}

	// Send stack lights + match state to alliance estop ESP32s
	if redEstopsIP != "" {
		esp32.sendJSON(redEstopsIP, stackCmd)
		esp32.sendJSON(redEstopsIP, matchCmd)
	}
	if blueEstopsIP != "" {
		esp32.sendJSON(blueEstopsIP, stackCmd)
		esp32.sendJSON(blueEstopsIP, matchCmd)
	}
}

// sendJSON marshals a message to JSON and sends it via UDP to the given IP.
func (esp32 *Esp32IO) sendJSON(ip string, msg interface{}) {
	data, err := json.Marshal(msg)
	if err != nil {
		log.Printf("ESP32 UDP: Failed to marshal message for %s: %v", ip, err)
		return
	}

	addr := &net.UDPAddr{IP: net.ParseIP(ip), Port: udpSendPort}
	conn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		log.Printf("ESP32 UDP: Failed to dial %s: %v", ip, err)
		return
	}
	defer conn.Close()

	_, err = conn.Write(data)
	if err != nil {
		log.Printf("ESP32 UDP: Failed to send to %s: %v", ip, err)
	}
}
