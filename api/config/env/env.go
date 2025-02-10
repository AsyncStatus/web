package env

type Env string

const (
	Local Env = "local"
	Dev   Env = "development"
	Prod  Env = "production"
)

func (e Env) Equals(e2 Env) bool {
	return e == e2
}

func FromString(s string) Env {
	if s == "" {
		return Local
	}

	e := Env(s)

	if e.Equals(Dev) {
		return Dev
	}

	if e.Equals(Prod) {
		return Prod
	}

	return Local
}

func (e *Env) SetValue(s string) error {
	*e = FromString(s)
	return nil
}
