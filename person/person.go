package person

type Gender int

const (
	GenderMan Gender = iota
	GenderWoman
)

type Person interface {
	Name() string
	Age() int
	Address() string
	Country() string
	Gender() any
}

type Join struct{}

func (*Join) Name() string    { return "john" }
func (*Join) Age() int        { return 25 }
func (*Join) Address() string { return "NewYork" }
func (*Join) Country() string { return "America" }
func (*Join) Gender() any     { return 1 }

func NewPerson() Person {
	return &Join{}
}
