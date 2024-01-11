package secretstring

type SecretString struct {
	value string
}

func NewSecretString(secret string) SecretString {
	return SecretString{secret}
}

// String implements fmt.Stringer.
func (s SecretString) String() string {
	return "SecretString(******)"
}

// ExposeSecret returns the underlying secret.
// It is best practice to never store the secret and expose it only when needed.
func (s SecretString) ExposeSecret() string {
	return s.value
}
