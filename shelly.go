package smarthome

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type ShellyInfo struct {
	Settings ShellySettings
	Status   ShellyStatus
}

type ShellySettings struct {
	// Common HTTP API
	Device struct {
		Type       string
		MAC        string
		Hostname   string
		NumOutputs int `json:"num_outputs"`
		NumMeters  int `json:"num_meters"`
		NumRollers int `json:"num_rollers"`
	}
	WiFiAP struct {
		Enabled bool
		SSID    string
		Key     string
	} `json:"wifi_ap"`
	WiFiStatus struct {
		Enabled    bool
		SSID       string
		IPv4Method string `json:"ipv4_method"`
		IP         string
		GW         string
		Mask       string
		DNS        string
	} `json:"wifi_sta"`
	WiFiStatus1 struct {
		Enabled    bool
		SSID       string
		IPv4Method string `json:"ipv4_method"`
		IP         string
		GW         string
		Mask       string
		DNS        string
	} `json:"wifi_sta1"`
	APRoaming struct {
		Enabled   bool
		Threshold int
	} `json:"ap_roaming"`
	MQTT struct {
		Enabled             bool
		Server              string
		User                string
		ID                  string
		ReconnectTimeoutMax float64 `json:"reconnect_timeout_max"`
		ReconnectTimeoutMin float64 `json:"reconnect_timeout_min"`
		CleanSession        bool    `json:"clean_session"`
		KeepAlive           int     `json:"keep_alive"`
		MaxQoS              int     `json:"max_qos"`
		Retain              bool
		UpdatePeriod        int `json:"update_period"`
	}
	CoIoT struct {
		Enabled      bool
		UpdatePeriod int `json:"update_period"`
		Peer         string
	}
	SNTP struct {
		Server  string
		Enabled bool
	}
	Login struct {
		Enabled     bool
		Unprotected bool
		Username    string
	}
	PinCode      string `json:"pin_code"`
	Name         string
	FW           string
	Discoverable bool
	BuildInfo    struct {
		BuildID        string `json:"build_id"`
		BuildTimestamp string `json:"build_timestamp"`
		BuildVersion   string `json:"build_version"`
	} `json:"build_info"`
	Cloud struct {
		Enabled   bool
		Connected bool
	}
	Timezone                  string
	Lat                       float64
	Lng                       float64
	TZAutodetect              bool
	TZUTCOffset               int  `json:"tz_utc_offset"`
	TZDST                     bool `json:"tz_dst"`
	TZDSTAuto                 bool `json:"tz_dst_auto"`
	Time                      string
	Unixtime                  int
	DebugEnable               bool `json:"debug_enable"`
	AllowCrossOrigin          bool `json:"allow_cross_origin"`
	WiFiRecoveryRebootEnabled bool `json:"wifirecovery_reboot_enabled"`

	// Shelly 1/2.5/...
	Mode string
	// ...many more missing...
}

type ShellyStatus struct {
	// Common HTTP API
	WiFiStatus struct {
		Connected bool
		SSID      string
		IP        string
		RSSI      int
	} `json:"wifi_sta"`

	// Extra HTTP API for some Shellies
	Temperature float64
	Voltage     float64
	Relays      []struct {
		IsOn          bool
		HasTimer      bool `json:"has_timer"`
		TimerDuration int  `json:"timer_duration"`
		IsValid       bool `json:"is_valid"`
		Source        string
	}
	Rollers []struct {
		State         string
		Source        string
		Power         float64
		IsValid       bool   `json:"is_valid"`
		LastDirection string `json:"last_direction"`
		CurrentPos    int    `json:"current_pos"`
		Positioning   bool
	}
	Meters []struct {
		Power     float64
		IsValid   bool `json:"is_valid"`
		Timestamp int
		Counters  []float64
		Total     int
	}
	Inputs []struct {
		Input    int
		Event    string
		EventCnt int `json:"event_cnt"`
	}
	Emeters []struct {
		Power         float64
		Reactive      float64
		Voltage       float64
		IsValid       bool `json:"is_valid"`
		Total         float64
		TotalReturned float64 `json:"total_returned"`
	}
}

func ShellyGetInfo(host string) (*ShellyInfo, error) {
	resp1, err := http.Get(fmt.Sprintf("http://%s/settings", host))
	if err != nil {
		return nil, err
	}
	defer resp1.Body.Close()
	body, err := io.ReadAll(resp1.Body)
	if err != nil {
		return nil, err
	}

	var shelly ShellyInfo

	err = json.Unmarshal(body, &shelly.Settings)
	if err != nil {
		return nil, err
	}

	// fmt.Printf("smarthome.ShellyInfo(): rawSettings=%v\n", string(body))
	// fmt.Printf("smarthome.ShellyInfo(): settings=%+v\n", shelly.Settings)

	resp2, err := http.Get(fmt.Sprintf("http://%s/status", host))
	if err != nil {
		return nil, err
	}
	defer resp2.Body.Close()
	body, err = io.ReadAll(resp2.Body)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(body, &shelly.Status)
	if err != nil {
		return nil, err
	}

	// fmt.Printf("smarthome.ShellyInfo(): rawStatus=%v\n", string(body))
	// fmt.Printf("smarthome.ShellyInfo(): status=%+v\n", shelly.Status)

	return &shelly, nil
}
