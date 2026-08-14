package crypto

const EncryptedPrefix = "enc:"

const RedactedValue = "********"

func RedactValue(value string, isSecret bool) string {
	if isSecret {
		return RedactedValue
	}
	return value
}
