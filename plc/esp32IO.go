// Copyright 20## Team ###. All Rights Reserved.
// Author: cpapplefamily@gmail.com (Corey Applegate)
//
// ESP32 IO handlers using UDP JSON communication.
//
// Architecture:
//   - Go listens on udpListenPort (5301) for JSON messages from ESP32s (many→1)
//   - Go sends JSON commands to each ESP32 IP on udpSendPort (5300) (1→many)
//   - Each ESP32 identifies itself by "role" in its messages
//   - Health is determined by receiving any valid message within the timeout window
//
// Message Protocol:
//   - One inbound message type: "node_status" (ESP32 → Go)
//   - One outbound message type: "node_command" (Go → ESP32)
//   - All device types use the same message shape; each reads only the fields
//     relevant to its role. Go-side uses separate structs per device type for clarity.

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
	udpHealthTimeoutMs = 1000 // Mark unhealthy if no message received for this long
	udpSendIntervalMs  = 100  // How often to send commands to ESP32s (10Hz)
	udpReadBufferSize  = 2048
)

// Esp32 defines the interface for ESP32 communication.
type Esp32 interface {
	Run()

	// Health configuration — score table / alliance estops
	IsScoreTableIOEnabled() bool
	IsRedEstopsEnabled() bool
	IsBlueEstopsEnabled() bool
	IsScoreTableHealthy() bool
	IsRedEstopsHealthy() bool
	IsBlueEstopsHealthy() bool

	// Health configuration — hubs
	IsRedHubEnabled() bool
	IsBlueHubEnabled() bool
	IsRedHubHealthy() bool
	IsBlueHubHealthy() bool

	// Address configuration
	SetScoreTableAddress(string)
	SetRedAllianceStationEstopAddress(string)
	SetBlueAllianceStationEstopAddress(string)
	SetRedHubAddress(string)
	SetBlueHubAddress(string)

	// Data received from ESP32s
	GetFieldEStop() bool
	GetTeamEStops() ([3]bool, [3]bool)
	GetTeamAStops() ([3]bool, [3]bool)
	GetRedHubScore() int
	GetBlueHubScore() int

	// Commands to send to ESP32s
	SetRedHubState(hubState string, motorDuty float64)
	SetBlueHubState(hubState string, motorDuty float64)
	SetRedHubLedPattern(ledPattern string)
	SetBlueHubLedPattern(ledPattern string)
	SendStackLights(red, blue, orange, green, buzzer bool)
	SendMatchState(state int)

	// Getters for current outbound command state (for debugging/display)
	GetStackLights() (red, blue, orange, green, buzzer bool)
	GetMatchState() int
	GetRedHubCommand() (hubState string, motorDuty float64, ledPattern string)
	GetBlueHubCommand() (hubState string, motorDuty float64, ledPattern string)
}

// --- Inbound message (ESP32 → Go): "node_status" ---

// NodeStationStatus represents one driver station's stop button states within a node_status message.
type NodeStationStatus struct {
	EStop bool `json:"eStop"`
	AStop bool `json:"aStop"`
}

// NodeStatusMessage is the unified inbound message from all ESP32 device types.
// Each device populates the fields relevant to its role:
//   - RED_ALLIANCE / BLUE_ALLIANCE: stations[]
//   - FMS_TABLE: fieldEStop
//   - RED_HUB / BLUE_HUB: score
type NodeStatusMessage struct {
	Type       string               `json:"type"` // Always "node_status"
	Role       string               `json:"role"` // "RED_ALLIANCE", "BLUE_ALLIANCE", "FMS_TABLE", "RED_HUB", "BLUE_HUB"
	FieldEStop bool                 `json:"fieldEStop"`
	Stations   [3]NodeStationStatus `json:"stations"`
	Score      int                  `json:"score"`
}

// --- Outbound message (Go → ESP32): "node_command" ---

// NodeCommandMessage is the unified outbound message sent to all ESP32 device types.
// Each device reads only the fields relevant to its role:
//   - Hubs read: hubState, motorDuty, redLight, blueLight, matchState
//   - Estop panels read: stackRed/Blue/Orange/Green, stackBuzzer, matchState
//   - Score table reads: stackRed/Blue/Orange/Green, stackBuzzer, matchState
type NodeCommandMessage struct {
	Type        string  `json:"type"`     // Always "node_command"
	HubState    string  `json:"hubState,omitempty"` // "DISABLED", "SCORING_ACTIVE", "SCORING_INACTIVE", etc.
	MotorDuty   float64 `json:"motorDuty"`
	LedPattern  string  `json:"ledPattern,omitempty"`
	StackRed    bool    `json:"stackRed"`
	StackBlue   bool    `json:"stackBlue"`
	StackOrange bool    `json:"stackOrange"`
	StackGreen  bool    `json:"stackGreen"`
	StackBuzzer bool    `json:"stackBuzzer"`
	MatchState      int     `json:"matchState"`
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
	RedHubIP             string
	BlueHubIP            string

	// Health tracking — last time a valid message was received from each node
	scoreTableLastHeard time.Time
	redEstopsLastHeard  time.Time
	blueEstopsLastHeard time.Time
	redHubLastHeard     time.Time
	blueHubLastHeard    time.Time

	// Received state from ESP32s
	fieldEStop   bool
	redEStops    [3]bool
	redAStops    [3]bool
	blueEStops   [3]bool
	blueAStops   [3]bool
	redHubScore  int
	blueHubScore int

	// Outbound command state (set by the arena, sent periodically)
	redHubCommand   NodeCommandMessage
	blueHubCommand  NodeCommandMessage
	stackLights     NodeCommandMessage // Shared stack light + match state for estop/table nodes
	matchState      int
	fieldResetLight bool

	// UDP connection
	listenConn *net.UDPConn

	mu sync.Mutex
}

// --- Address setters ---

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

func (esp32 *Esp32IO) SetRedHubAddress(address string) {
	address = strings.TrimSpace(address)
	if address != "" && net.ParseIP(address) == nil {
		log.Printf("Invalid Red Hub IP address: %s", address)
		return
	}
	esp32.mu.Lock()
	esp32.RedHubIP = address
	esp32.mu.Unlock()
	if address != "" {
		log.Printf("Set Red Hub IP to: %s", address)
	}
}

func (esp32 *Esp32IO) SetBlueHubAddress(address string) {
	address = strings.TrimSpace(address)
	if address != "" && net.ParseIP(address) == nil {
		log.Printf("Invalid Blue Hub IP address: %s", address)
		return
	}
	esp32.mu.Lock()
	esp32.BlueHubIP = address
	esp32.mu.Unlock()
	if address != "" {
		log.Printf("Set Blue Hub IP to: %s", address)
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

func (esp32 *Esp32IO) IsRedHubEnabled() bool {
	esp32.mu.Lock()
	defer esp32.mu.Unlock()
	return esp32.RedHubIP != ""
}

func (esp32 *Esp32IO) IsBlueHubEnabled() bool {
	esp32.mu.Lock()
	defer esp32.mu.Unlock()
	return esp32.BlueHubIP != ""
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

func (esp32 *Esp32IO) IsRedHubHealthy() bool {
	esp32.mu.Lock()
	defer esp32.mu.Unlock()
	if esp32.RedHubIP == "" {
		return false
	}
	return time.Since(esp32.redHubLastHeard) < time.Millisecond*udpHealthTimeoutMs
}

func (esp32 *Esp32IO) IsBlueHubHealthy() bool {
	esp32.mu.Lock()
	defer esp32.mu.Unlock()
	if esp32.BlueHubIP == "" {
		return false
	}
	return time.Since(esp32.blueHubLastHeard) < time.Millisecond*udpHealthTimeoutMs
}

func (esp32 *Esp32IO) SetFieldResetLight(state bool) {
	esp32.mu.Lock()
	defer esp32.mu.Unlock()
	esp32.fieldResetLight = state
}

// --- Data accessors (read state received from ESP32s) ---

func (esp32 *Esp32IO) GetFieldEStop() bool {
	esp32.mu.Lock()
	defer esp32.mu.Unlock()
	if esp32.ScoreTableIP == "" {
		return false
	} else if time.Since(esp32.scoreTableLastHeard) >= time.Millisecond*udpHealthTimeoutMs {
		return true
	}
	return esp32.fieldEStop
}

func (esp32 *Esp32IO) GetTeamEStops() ([3]bool, [3]bool) {
	esp32.mu.Lock()
	defer esp32.mu.Unlock()

	redEStops := esp32.redEStops
	if esp32.RedAllianceEstopsIP == "" {
		redEStops = [3]bool{false, false, false}
	} else if time.Since(esp32.redEstopsLastHeard) >= time.Millisecond*udpHealthTimeoutMs {
		redEStops = [3]bool{true, true, true}
	}

	blueEStops := esp32.blueEStops
	if esp32.BlueAllianceEstopsIP == "" {
		blueEStops = [3]bool{false, false, false}
	} else if time.Since(esp32.blueEstopsLastHeard) >= time.Millisecond*udpHealthTimeoutMs {
		blueEStops = [3]bool{true, true, true}
	}

	return redEStops, blueEStops
}

func (esp32 *Esp32IO) GetTeamAStops() ([3]bool, [3]bool) {
	esp32.mu.Lock()
	defer esp32.mu.Unlock()

	redAStops := esp32.redAStops
	if esp32.RedAllianceEstopsIP == "" {
		redAStops = [3]bool{false, false, false}
	} else if time.Since(esp32.redEstopsLastHeard) >= time.Millisecond*udpHealthTimeoutMs {
		redAStops = [3]bool{true, true, true}
	}

	blueAStops := esp32.blueAStops
	if esp32.BlueAllianceEstopsIP == "" {
		blueAStops = [3]bool{false, false, false}
	} else if time.Since(esp32.blueEstopsLastHeard) >= time.Millisecond*udpHealthTimeoutMs {
		blueAStops = [3]bool{true, true, true}
	}

	return redAStops, blueAStops
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

func (esp32 *Esp32IO) SetRedHubState(hubState string, motorDuty float64) {
	esp32.mu.Lock()
	defer esp32.mu.Unlock()
	esp32.redHubCommand.HubState = hubState
	esp32.redHubCommand.MotorDuty = motorDuty
}

func (esp32 *Esp32IO) SetBlueHubState(hubState string, motorDuty float64) {
	esp32.mu.Lock()
	defer esp32.mu.Unlock()
	esp32.blueHubCommand.HubState = hubState
	esp32.blueHubCommand.MotorDuty = motorDuty
}

func (esp32 *Esp32IO) SetRedHubLedPattern(ledPattern string) {
	esp32.mu.Lock()
	defer esp32.mu.Unlock()
	esp32.redHubCommand.LedPattern = ledPattern
}

func (esp32 *Esp32IO) SetBlueHubLedPattern(ledPattern string) {
	esp32.mu.Lock()
	defer esp32.mu.Unlock()
	esp32.blueHubCommand.LedPattern = ledPattern
}

func (esp32 *Esp32IO) SendStackLights(red, blue, orange, green, buzzer bool) {
	esp32.mu.Lock()
	defer esp32.mu.Unlock()
	esp32.stackLights.StackRed = red
	esp32.stackLights.StackBlue = blue
	esp32.stackLights.StackOrange = orange
	esp32.stackLights.StackGreen = green
	esp32.stackLights.StackBuzzer = buzzer
}

func (esp32 *Esp32IO) SendMatchState(state int) {
	esp32.mu.Lock()
	defer esp32.mu.Unlock()
	esp32.matchState = state
}

// --- Outbound command getters (for debugging/display) ---

func (esp32 *Esp32IO) GetStackLights() (red, blue, orange, green, buzzer bool) {
	esp32.mu.Lock()
	defer esp32.mu.Unlock()
	return esp32.stackLights.StackRed, esp32.stackLights.StackBlue,
		esp32.stackLights.StackOrange, esp32.stackLights.StackGreen, esp32.stackLights.StackBuzzer
}

func (esp32 *Esp32IO) GetMatchState() int {
	esp32.mu.Lock()
	defer esp32.mu.Unlock()
	return esp32.matchState
}

func (esp32 *Esp32IO) GetRedHubCommand() (hubState string, motorDuty float64, ledPattern string) {
	esp32.mu.Lock()
	defer esp32.mu.Unlock()
	return esp32.redHubCommand.HubState, esp32.redHubCommand.MotorDuty, esp32.redHubCommand.LedPattern
}

func (esp32 *Esp32IO) GetBlueHubCommand() (hubState string, motorDuty float64, ledPattern string) {
	esp32.mu.Lock()
	defer esp32.mu.Unlock()
	return esp32.blueHubCommand.HubState, esp32.blueHubCommand.MotorDuty, esp32.blueHubCommand.LedPattern
}

// --- Main run loop ---

// Run starts the UDP listener and periodic sender. Should be called as a goroutine.
func (esp32 *Esp32IO) Run() {
	// Initialize default command states
	esp32.redHubCommand = NodeCommandMessage{Type: "node_command", HubState: "DISABLED"}
	esp32.blueHubCommand = NodeCommandMessage{Type: "node_command", HubState: "DISABLED"}
	esp32.stackLights = NodeCommandMessage{Type: "node_command"}

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
	case "node_status":
		var statusMsg NodeStatusMessage
		if err := json.Unmarshal(data, &statusMsg); err != nil {
			log.Printf("ESP32 UDP: Failed to parse node_status from %s: %v", remoteAddr, err)
			return
		}
		esp32.handleNodeStatus(&statusMsg, remoteAddr)

	default:
		log.Printf("ESP32 UDP: Unknown message type '%s' from %s", msg.Type, remoteAddr)
	}
}

// handleNodeStatus processes a unified node_status message, routing by role.
func (esp32 *Esp32IO) handleNodeStatus(msg *NodeStatusMessage, remoteAddr *net.UDPAddr) {
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

	case "RED_HUB":
		esp32.redHubScore = msg.Score
		esp32.redHubLastHeard = time.Now()

	case "BLUE_HUB":
		esp32.blueHubScore = msg.Score
		esp32.blueHubLastHeard = time.Now()

	default:
		log.Printf("ESP32 UDP: Unknown role '%s' from %s", msg.Role, remoteAddr)
	}
}

// sendCommandsToAll sends the current command state to all configured ESP32s.
func (esp32 *Esp32IO) sendCommandsToAll() {
	esp32.mu.Lock()
	redHubCmd := esp32.redHubCommand
	blueHubCmd := esp32.blueHubCommand

	// Build the shared command for estop/table nodes with stack lights + match state
	sharedCmd := NodeCommandMessage{
		Type:        "node_command",
		StackRed:    esp32.stackLights.StackRed,
		StackBlue:   esp32.stackLights.StackBlue,
		StackOrange: esp32.stackLights.StackOrange,
		StackGreen:  esp32.stackLights.StackGreen,
		StackBuzzer: esp32.stackLights.StackBuzzer,
		MatchState:  esp32.matchState,
	}

	// Include match state in hub commands
	redHubCmd.MatchState = esp32.matchState
	blueHubCmd.MatchState = esp32.matchState

	scoreTableIP := esp32.ScoreTableIP
	redEstopsIP := esp32.RedAllianceEstopsIP
	blueEstopsIP := esp32.BlueAllianceEstopsIP
	redHubIP := esp32.RedHubIP
	blueHubIP := esp32.BlueHubIP
	esp32.mu.Unlock()

	// Send to score table
	if scoreTableIP != "" {
		esp32.sendJSON(scoreTableIP, sharedCmd)
	}

	// Send to alliance estop panels
	if redEstopsIP != "" {
		esp32.sendJSON(redEstopsIP, sharedCmd)
	}
	if blueEstopsIP != "" {
		esp32.sendJSON(blueEstopsIP, sharedCmd)
	}

	// Send per-hub commands to each hub
	if redHubIP != "" {
		esp32.sendJSON(redHubIP, redHubCmd)
	}
	if blueHubIP != "" {
		esp32.sendJSON(blueHubIP, blueHubCmd)
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
