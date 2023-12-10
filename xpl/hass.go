package xpl

type HAConfig struct {
	Name                   string   `json:"name,omitempty"`
	UniqueID               string   `json:"unique_id,omitempty"`
	DeviceClass            string   `json:"device_class,omitempty"`
	StateTopic             string   `json:"state_topic,omitempty"`
	Unit                   string   `json:"unit_of_measurement,omitempty"`
	CommandTopic           string   `json:"command_topic,omitempty"`
	BrightnessScale        int      `json:"brightness_scale,omitempty"`
	BrightnessStateTopic   string   `json:"brightness_state_topic,omitempty"`
	BrightnessCommandTopic string   `json:"brightness_command_topic,omitempty"`
	Icon                   string   `json:"icon,omitempty"`
	Device                 HADevice `json:"device,omitempty"`
	SupportedFeatures      []string `json:"supported_features,omitempty"`
	CodeArmRequired        bool     `json:"code_arm_required,omitempty"`
	CodeDisarmRequired     bool     `json:"code_disarm_required,omitempty"`
	CodeTriggerRequired    bool     `json:"code_trigger_required,omitempty"`
}

type HADevice struct {
	Identifiers  []string `json:"identifiers,omitempty"`
	Manifacturer string   `json:"manufacturer"`
	Name         string   `json:"name"`
	Model        string   `json:"model"`
}
