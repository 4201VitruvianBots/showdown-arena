// Copyright 20## Team ###. All Rights Reserved.
//
// Contains a fake implementation of the Esp32 interface for testing.

package field

type FakeEsp32 struct {
	// Configured addresses
	scoreTableIP string
	redEstopsIP  string
	blueEstopsIP string
	redHubIP     string
	blueHubIP    string

	// Simulated health
	scoreTableHealthy bool
	redEstopsHealthy  bool
	blueEstopsHealthy bool
	redHubHealthy     bool
	blueHubHealthy    bool

	// Received state (set by tests to simulate ESP32 data)
	fieldEStop   bool
	redEStops    [3]bool
	redAStops    [3]bool
	blueEStops   [3]bool
	blueAStops   [3]bool
	redHubScore  int
	blueHubScore int

	// Sent commands (set by arena, readable by tests for assertions)
	RedHubState   string
	RedMotorDuty  float64
	RedRedLight   bool
	RedBlueLight  bool
	BlueHubState  string
	BlueMotorDuty float64
	BlueRedLight  bool
	BlueBlueLight bool
	StackRed      bool
	StackBlue     bool
	StackOrange   bool
	StackGreen    bool
	StackBuzzer   bool
	MatchStateVal int
}

func (esp32 *FakeEsp32) Run() {
}

// --- Health / Enabled ---

func (esp32 *FakeEsp32) IsScoreTableIOEnabled() bool {
	return esp32.scoreTableIP != ""
}

func (esp32 *FakeEsp32) IsRedEstopsEnabled() bool {
	return esp32.redEstopsIP != ""
}

func (esp32 *FakeEsp32) IsBlueEstopsEnabled() bool {
	return esp32.blueEstopsIP != ""
}

func (esp32 *FakeEsp32) IsRedHubEnabled() bool {
	return esp32.redHubIP != ""
}

func (esp32 *FakeEsp32) IsBlueHubEnabled() bool {
	return esp32.blueHubIP != ""
}

func (esp32 *FakeEsp32) IsScoreTableHealthy() bool {
	return esp32.scoreTableHealthy
}

func (esp32 *FakeEsp32) IsRedEstopsHealthy() bool {
	return esp32.redEstopsHealthy
}

func (esp32 *FakeEsp32) IsBlueEstopsHealthy() bool {
	return esp32.blueEstopsHealthy
}

func (esp32 *FakeEsp32) IsRedHubHealthy() bool {
	return esp32.redHubHealthy
}

func (esp32 *FakeEsp32) IsBlueHubHealthy() bool {
	return esp32.blueHubHealthy
}

// --- Address setters ---

func (esp32 *FakeEsp32) SetScoreTableAddress(address string) {
	esp32.scoreTableIP = address
}

func (esp32 *FakeEsp32) SetRedAllianceStationEstopAddress(address string) {
	esp32.redEstopsIP = address
}

func (esp32 *FakeEsp32) SetBlueAllianceStationEstopAddress(address string) {
	esp32.blueEstopsIP = address
}

func (esp32 *FakeEsp32) SetRedHubAddress(address string) {
	esp32.redHubIP = address
}

func (esp32 *FakeEsp32) SetBlueHubAddress(address string) {
	esp32.blueHubIP = address
}

// --- Data accessors ---

func (esp32 *FakeEsp32) GetFieldEStop() bool {
	return esp32.fieldEStop
}

func (esp32 *FakeEsp32) GetTeamEStops() ([3]bool, [3]bool) {
	return esp32.redEStops, esp32.blueEStops
}

func (esp32 *FakeEsp32) GetTeamAStops() ([3]bool, [3]bool) {
	return esp32.redAStops, esp32.blueAStops
}

func (esp32 *FakeEsp32) GetRedHubScore() int {
	return esp32.redHubScore
}

func (esp32 *FakeEsp32) GetBlueHubScore() int {
	return esp32.blueHubScore
}

// --- Command setters ---

func (esp32 *FakeEsp32) SendRedHubCommand(hubState string, motorDuty float64, redLight, blueLight bool) {
	esp32.RedHubState = hubState
	esp32.RedMotorDuty = motorDuty
	esp32.RedRedLight = redLight
	esp32.RedBlueLight = blueLight
}

func (esp32 *FakeEsp32) SendBlueHubCommand(hubState string, motorDuty float64, redLight, blueLight bool) {
	esp32.BlueHubState = hubState
	esp32.BlueMotorDuty = motorDuty
	esp32.BlueRedLight = redLight
	esp32.BlueBlueLight = blueLight
}

func (esp32 *FakeEsp32) SendStackLights(red, blue, orange, green, buzzer bool) {
	esp32.StackRed = red
	esp32.StackBlue = blue
	esp32.StackOrange = orange
	esp32.StackGreen = green
	esp32.StackBuzzer = buzzer
}

func (esp32 *FakeEsp32) SendMatchState(state int) {
	esp32.MatchStateVal = state
}

// --- Outbound command getters ---

func (esp32 *FakeEsp32) GetStackLights() (red, blue, orange, green, buzzer bool) {
	return esp32.StackRed, esp32.StackBlue, esp32.StackOrange, esp32.StackGreen, esp32.StackBuzzer
}

func (esp32 *FakeEsp32) GetMatchState() int {
	return esp32.MatchStateVal
}

func (esp32 *FakeEsp32) GetRedHubCommand() (hubState string, motorDuty float64, redLight, blueLight bool) {
	return esp32.RedHubState, esp32.RedMotorDuty, esp32.RedRedLight, esp32.RedBlueLight
}

func (esp32 *FakeEsp32) GetBlueHubCommand() (hubState string, motorDuty float64, redLight, blueLight bool) {
	return esp32.BlueHubState, esp32.BlueMotorDuty, esp32.BlueRedLight, esp32.BlueBlueLight
}
