package ipaynow

// redactSensitive returns a shallow copy of fields with the mhtSignature/
// signature values masked to "***" so request/response payloads can be
// logged at DEBUG without leaking signature material (spec §8). appKey is
// never part of fields and must never be logged at any level.
func redactSensitive(fields map[string]string) map[string]string {
	redacted := make(map[string]string, len(fields))
	for k, v := range fields {
		if k == "mhtSignature" || k == "signature" {
			v = "***"
		}
		redacted[k] = v
	}
	return redacted
}
