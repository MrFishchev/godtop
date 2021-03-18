package colorschemes

const (
	Bold int = 1 << (iota + 9)
	Underline
	Reverse
)

type Colorscheme struct {
	Name       string
	MainFg     int
	MainBg     int
	BorderFg   int
	BorderBg   int
	Containers []int
	Cursor     int
	ValueLow   int
	ValueHigh int
	CpuLines   []int
}
