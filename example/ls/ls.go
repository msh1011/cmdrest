package ls

// LS defines all the ls args we may want
type LS struct {
	Long bool   `rcmd:"l"`
	All  bool   `rcmd:"a"`
	Time bool   `rcmd:"t"`
	Dir  string `rcmd:"0"`
}

// DefaultLS defines default values to use for LS command
var DefaultLS = &LS{
	Long: true,
	All:  false,
	Dir:  "",
}

// Name returns the binary name of the command
func (ls *LS) Name() string {
	return "ls"
}
