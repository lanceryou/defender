package defender

// 治理接口
type Defender interface {
	Check(resource string) error
	String() string
}

var (
	defenderMap = make(map[string]Defender)
)

func SetDefender(defender Defender) {
	defenderMap[defender.String()] = defender
}

// 检测
func Check(resource string) error {
	for _, v := range defenderMap {
		if err := v.Check(resource); err != nil {
			return err
		}
	}
	return nil
}
