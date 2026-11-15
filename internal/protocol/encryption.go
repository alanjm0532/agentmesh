package protocol

// EncryptionAlgorithm identifies the versioned encryption scheme used by a message.
// Its string representation is also the JSON wire value.
type EncryptionAlgorithm string

const (
	// EncryptionAlgorithmUnknown represents an unspecified or unsupported scheme.
	EncryptionAlgorithmUnknown EncryptionAlgorithm = "unknown"
	// EncryptionAlgorithmNone indicates that the body contains unencrypted text.
	EncryptionAlgorithmNone EncryptionAlgorithm = "none"
)

// String returns the protocol name, falling back to unknown for unrecognized values.
func (a EncryptionAlgorithm) String() string {
	switch a {
	case EncryptionAlgorithmNone:
		return string(EncryptionAlgorithmNone)
	default:
		return string(EncryptionAlgorithmUnknown)
	}
}
