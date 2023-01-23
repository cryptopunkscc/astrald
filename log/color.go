package log

var reset = "\033[0m"
var red = "\033[31m"
var green = "\033[32m"
var yellow = "\033[33m"
var blue = "\033[34m"
var purple = "\033[35m"
var cyan = "\033[36m"
var gray = "\033[37m"
var white = "\033[97m"

func (l *Logger) Reset() string  { return reset }
func (l *Logger) Red() string    { return red }
func (l *Logger) Green() string  { return green }
func (l *Logger) Yellow() string { return yellow }
func (l *Logger) Blue() string   { return blue }
func (l *Logger) Purple() string { return purple }
func (l *Logger) Cyan() string   { return cyan }
func (l *Logger) Gray() string   { return gray }
func (l *Logger) White() string  { return white }

func Reset() string  { return instance.Reset() }
func Red() string    { return instance.Red() }
func Green() string  { return instance.Green() }
func Yellow() string { return instance.Yellow() }
func Blue() string   { return instance.Blue() }
func Purple() string { return instance.Purple() }
func Cyan() string   { return instance.Cyan() }
func Gray() string   { return instance.Gray() }
func White() string  { return instance.White() }
