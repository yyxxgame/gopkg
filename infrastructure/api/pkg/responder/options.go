package responder

type (
	Option func(*Responder)
)

func WithEncryptionHeader(encryptionHeader string) Option {
	return func(m *Responder) {
		m.encryptionHeader = encryptionHeader
	}
}

func WithEncryptionKey(encryptionKey string) Option {
	return func(m *Responder) {
		m.encryptionKey = encryptionKey
	}
}

func WithAccessLogDetails(enable bool) Option {
	return func(m *Responder) {
		m.accessLogDetails = enable
	}
}
