package registers

// Status represents the verification state of a register.
type Status string

const (
	Verified  Status = "verified"
	Candidate Status = "candidate"
	Invalid   Status = "invalid"
)

// Register defines a single Modbus holding register.
type Register struct {
	Address     uint16
	Multiplier  float64
	Unit        string
	Description string
	Status      Status
}

// All registers in display order.
var All = []struct {
	Name string
	Reg  Register
}{
	// Battery
	{"bat_V", Register{0xB7, 0.01, "V", "Напряжение АКБ", Verified}},
	{"bat_A", Register{0xB8, 0.05, "A", "Ток АКБ — не совпадает с логгером (3A vs 11A)", Invalid}},
	{"bat_W", Register{0xBE, 1, "W", "Мощность АКБ", Verified}},
	{"bat_SOC", Register{0xA4, 1, "%", "Заряд АКБ", Verified}},
	{"bat_T", Register{0xAA, 1, "C", "Температура АКБ — неверный адрес", Invalid}},
	// Grid
	{"grid_V", Register{0x96, 0.1, "V", "Напряжение сети", Verified}},
	{"grid_Hz", Register{0x4F, 0.01, "Hz", "Частота сети (0x4F = 50.03)", Candidate}},
	{"grid_W", Register{0x97, 1, "W", "Мощность сети", Invalid}},
	// Load
	{"load_W", Register{0xB2, 1, "W", "Нагрузка дома", Verified}},
	{"load_A", Register{0xB3, 0.1, "A", "Ток нагрузки", Invalid}},
	// Inverter
	{"inv_T", Register{0xAB, 1, "C", "Температура инвертора (общая)", Invalid}},
	{"inv_T_AC", Register{0xDB, 1, "C", "Температура AC (радиатор)", Candidate}},
	{"inv_T_DC", Register{0xDA, 1, "C", "Температура DC (радиатор)", Candidate}},
	{"inv_st", Register{0x6A, 1, "", "Статус инвертора", Candidate}},
	// PV
	{"pv1_W", Register{0x9C, 1, "W", "PV1 вход", Invalid}},
	{"pv2_W", Register{0x9F, 1, "W", "PV2 вход", Candidate}},
	// Energy
	{"today_kWh", Register{0x79, 0.01, "kWh", "За сегодня", Candidate}},
	{"total_kWh", Register{0x7C, 0.1, "kWh", "Всего", Verified}},
}

// Active returns registers safe for regular collection.
func Active(includeUnverified bool) map[string]Register {
	out := make(map[string]Register, len(All))
	for _, entry := range All {
		if includeUnverified || entry.Reg.Status == Verified || entry.Reg.Status == Candidate {
			out[entry.Name] = entry.Reg
		}
	}
	return out
}

// Names returns ordered register names.
func Names(includeUnverified bool) []string {
	var names []string
	for _, entry := range All {
		if includeUnverified || entry.Reg.Status == Verified || entry.Reg.Status == Candidate {
			names = append(names, entry.Name)
		}
	}
	return names
}
